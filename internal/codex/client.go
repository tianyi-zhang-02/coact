package codex

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Client speaks JSON-RPC 2.0 to a Codex app-server over a byte stream.
type Client struct {
	w     io.Writer
	r     *bufio.Reader
	codec Codec

	mu      sync.Mutex
	nextID  int
	pending map[int]chan rpcResponse

	onMessage func(text string) // called with each completed agent message

	stateMu sync.Mutex
	thread  string
	turn    string // active turn id, "" when idle
	deltas  map[string]*strings.Builder

	closeOnce sync.Once
	done      chan struct{}
}

type rpcResponse struct {
	Result json.RawMessage
	Err    *rpcError
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// New builds a client over rw (Codex app-server stdio, or a test pipe).
// onMessage is invoked with each completed agent message from Codex.
func New(rw io.ReadWriter, codec Codec, onMessage func(text string)) *Client {
	if codec == nil {
		codec = NewlineCodec{}
	}
	c := &Client{
		w:         rw,
		r:         bufio.NewReaderSize(rw, 64*1024),
		codec:     codec,
		pending:   map[int]chan rpcResponse{},
		onMessage: onMessage,
		deltas:    map[string]*strings.Builder{},
		done:      make(chan struct{}),
	}
	go c.readLoop()
	return c
}

func (c *Client) readLoop() {
	defer c.Close()
	for {
		msg, err := c.codec.ReadMessage(c.r)
		if len(msg) > 0 {
			c.handle(msg)
		}
		if err != nil {
			return
		}
	}
}

func (c *Client) handle(msg []byte) {
	var probe struct {
		ID     *int   `json:"id"`
		Method string `json:"method"`
	}
	if json.Unmarshal(msg, &probe) != nil {
		return
	}
	switch {
	case probe.Method != "" && probe.ID == nil:
		c.handleNotification(probe.Method, msg)
	case probe.Method != "" && probe.ID != nil:
		c.handleServerRequest(*probe.ID, probe.Method)
	case probe.ID != nil:
		var resp struct {
			ID     int             `json:"id"`
			Result json.RawMessage `json:"result"`
			Error  *rpcError       `json:"error"`
		}
		if json.Unmarshal(msg, &resp) == nil {
			c.deliver(resp.ID, rpcResponse{Result: resp.Result, Err: resp.Error})
		}
	}
}

func (c *Client) handleNotification(method string, msg []byte) {
	switch method {
	case "turn/started":
		var p struct {
			Params struct {
				Turn struct {
					ID string `json:"id"`
				} `json:"turn"`
			} `json:"params"`
		}
		if json.Unmarshal(msg, &p) == nil {
			c.setTurn(p.Params.Turn.ID)
		}
	case "item/agentMessage/delta":
		var p struct {
			Params struct {
				ItemID string `json:"itemId"`
				Delta  string `json:"delta"`
			} `json:"params"`
		}
		if json.Unmarshal(msg, &p) == nil {
			c.appendDelta(p.Params.ItemID, p.Params.Delta)
		}
	case "item/completed":
		var p struct {
			Params struct {
				Item struct {
					ID string `json:"id"`
				} `json:"item"`
			} `json:"params"`
		}
		if json.Unmarshal(msg, &p) == nil {
			c.flush(p.Params.Item.ID)
		}
	case "turn/completed":
		c.setTurn("")
	}
}

// handleServerRequest answers Codex's approval requests. For the executor model
// CoAct auto-approves (the user launched Codex as a bridge worker); the exact
// approval result shape is verified against real codex at integration time.
func (c *Client) handleServerRequest(id int, method string) {
	if strings.HasSuffix(method, "requestApproval") {
		c.reply(id, map[string]any{"decision": "approved"})
		return
	}
	c.reply(id, map[string]any{})
}

func (c *Client) setTurn(id string) {
	c.stateMu.Lock()
	c.turn = id
	c.stateMu.Unlock()
}

func (c *Client) activeTurn() string {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	return c.turn
}

func (c *Client) appendDelta(itemID, delta string) {
	c.stateMu.Lock()
	b := c.deltas[itemID]
	if b == nil {
		b = &strings.Builder{}
		c.deltas[itemID] = b
	}
	b.WriteString(delta)
	c.stateMu.Unlock()
}

func (c *Client) flush(itemID string) {
	c.stateMu.Lock()
	b := c.deltas[itemID]
	delete(c.deltas, itemID)
	c.stateMu.Unlock()
	if b != nil && strings.TrimSpace(b.String()) != "" && c.onMessage != nil {
		c.onMessage(b.String())
	}
}

// --- requests -------------------------------------------------------------

func (c *Client) reply(id int, result any) {
	b, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "result": result})
	c.mu.Lock()
	_ = c.codec.WriteMessage(c.w, b)
	c.mu.Unlock()
}

func (c *Client) call(method string, params any) (json.RawMessage, error) {
	c.mu.Lock()
	c.nextID++
	id := c.nextID
	ch := make(chan rpcResponse, 1)
	c.pending[id] = ch
	req, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params})
	err := c.codec.WriteMessage(c.w, req)
	c.mu.Unlock()
	if err != nil {
		return nil, err
	}
	select {
	case resp := <-ch:
		if resp.Err != nil {
			return nil, fmt.Errorf("codex %s: %s", method, resp.Err.Message)
		}
		return resp.Result, nil
	case <-time.After(60 * time.Second):
		return nil, fmt.Errorf("codex %s: timeout", method)
	case <-c.done:
		return nil, errors.New("codex: connection closed")
	}
}

func (c *Client) deliver(id int, resp rpcResponse) {
	c.mu.Lock()
	ch := c.pending[id]
	delete(c.pending, id)
	c.mu.Unlock()
	if ch != nil {
		ch <- resp
	}
}

func textInput(text string) []any {
	return []any{map[string]any{"type": "text", "text": text}}
}

// Initialize performs the app-server handshake and returns the userAgent.
func (c *Client) Initialize() (string, error) {
	res, err := c.call("initialize", map[string]any{"clientInfo": map[string]any{"name": "coact", "version": "0.1"}})
	if err != nil {
		return "", err
	}
	var r struct {
		UserAgent string `json:"userAgent"`
	}
	_ = json.Unmarshal(res, &r)
	return r.UserAgent, nil
}

// StartThread opens a thread and records its id.
func (c *Client) StartThread(cwd string) (string, error) {
	res, err := c.call("thread/start", map[string]any{"cwd": cwd})
	if err != nil {
		return "", err
	}
	var r struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if json.Unmarshal(res, &r) != nil || r.Thread.ID == "" {
		return "", errors.New("codex thread/start: no thread id")
	}
	c.stateMu.Lock()
	c.thread = r.Thread.ID
	c.stateMu.Unlock()
	return r.Thread.ID, nil
}

func (c *Client) threadID() string {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	return c.thread
}

// StartTurn begins a new turn with text and records the active turn id.
func (c *Client) StartTurn(text string) (string, error) {
	res, err := c.call("turn/start", map[string]any{"threadId": c.threadID(), "input": textInput(text)})
	if err != nil {
		return "", err
	}
	var r struct {
		Turn struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	_ = json.Unmarshal(res, &r)
	if r.Turn.ID != "" {
		c.setTurn(r.Turn.ID)
	}
	return r.Turn.ID, nil
}

// Steer feeds text into the currently running turn without interrupting it.
func (c *Client) Steer(text string) error {
	turn := c.activeTurn()
	if turn == "" {
		return errors.New("codex: no active turn to steer")
	}
	_, err := c.call("turn/steer", map[string]any{
		"threadId":       c.threadID(),
		"expectedTurnId": turn,
		"input":          textInput(text),
	})
	return err
}

// Send routes text to Codex: steer if a turn is running, else start a new one.
func (c *Client) Send(text string) error {
	if c.activeTurn() != "" {
		return c.Steer(text)
	}
	_, err := c.StartTurn(text)
	return err
}

// Close shuts down the client.
func (c *Client) Close() error {
	c.closeOnce.Do(func() { close(c.done) })
	return nil
}
