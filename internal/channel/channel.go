// Package channel implements a Claude Code "channel": an MCP server (stdio,
// JSON-RPC 2.0) that pushes messages from the other agent into a running Claude
// session mid-turn (as <channel> events) and exposes a reply tool so Claude can
// send back. This is CoAct's real-time push for the Claude side, in pure Go —
// no MCP SDK, just newline-delimited JSON-RPC.
//
// Requires Claude Code v2.1.80+ (channels are a research preview). Register the
// server in .mcp.json and launch with:
//
//	claude --dangerously-load-development-channels server:coact-<agent>
package channel

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/tianyi-zhang-02/coact/internal/inbox"
	"github.com/tianyi-zhang-02/coact/internal/journal"
)

const protocolVersion = "2025-06-18"

// Server is a stdio MCP channel for one agent, relaying to/from its peer.
type Server struct {
	agent, peer          string
	inboxDir, journalDir string
	in                   io.Reader
	out                  io.Writer
	mu                   sync.Mutex
	poll                 time.Duration
}

// New builds a channel server. in/out are the MCP stdio transport (stdin/stdout
// when spawned by Claude Code).
func New(agent, peer, inboxDir, journalDir string, in io.Reader, out io.Writer) *Server {
	return &Server{
		agent: agent, peer: peer,
		inboxDir: inboxDir, journalDir: journalDir,
		in: in, out: out, poll: 500 * time.Millisecond,
	}
}

type rpcMsg struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Run serves until in is closed. It also polls the agent's inbox and pushes any
// new messages as channel events.
func (s *Server) Run() error {
	stop := make(chan struct{})
	go s.pollInbox(stop)
	defer close(stop)

	sc := bufio.NewScanner(s.in)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		s.dispatch(line)
	}
	return sc.Err()
}

func (s *Server) write(v any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, _ := json.Marshal(v)
	s.out.Write(append(b, '\n'))
}

func (s *Server) result(id json.RawMessage, result any) {
	s.write(map[string]any{"jsonrpc": "2.0", "id": rawOrNull(id), "result": result})
}

func rawOrNull(id json.RawMessage) any {
	if len(id) == 0 {
		return nil
	}
	return id
}

// dispatch handles one JSON-RPC message. Exposed logic is unit-tested.
func (s *Server) dispatch(line []byte) {
	var m rpcMsg
	if json.Unmarshal(line, &m) != nil {
		return
	}
	switch m.Method {
	case "initialize":
		pv := protocolVersion
		var p struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		if json.Unmarshal(m.Params, &p) == nil && p.ProtocolVersion != "" {
			pv = p.ProtocolVersion
		}
		s.result(m.ID, map[string]any{
			"protocolVersion": pv,
			"capabilities": map[string]any{
				"experimental": map[string]any{"claude/channel": map[string]any{}},
				"tools":        map[string]any{},
			},
			"serverInfo": map[string]any{"name": "coact-" + s.agent, "version": "0.1"},
			"instructions": "Messages from the other agent (" + s.peer + ") arrive as " +
				`<channel source="coact-` + s.peer + `">...</channel>. To reply, call the reply ` +
				"tool with the text; it is delivered to " + s.peer + ".",
		})
	case "tools/list":
		s.result(m.ID, map[string]any{"tools": []any{s.replyTool()}})
	case "tools/call":
		s.handleToolCall(m)
	case "ping":
		s.result(m.ID, map[string]any{})
	default:
		// notifications (e.g. notifications/initialized) and unknown methods:
		// notifications carry no id and need no response.
		if len(m.ID) != 0 {
			s.write(map[string]any{"jsonrpc": "2.0", "id": rawOrNull(m.ID),
				"error": map[string]any{"code": -32601, "message": "method not found: " + m.Method}})
		}
	}
}

func (s *Server) replyTool() map[string]any {
	return map[string]any{
		"name":        "reply",
		"description": "Send a message to the other agent (" + s.peer + ")",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"text": map[string]any{"type": "string", "description": "the message to send"},
			},
			"required": []any{"text"},
		},
	}
}

func (s *Server) handleToolCall(m rpcMsg) {
	var p struct {
		Name      string `json:"name"`
		Arguments struct {
			Text string `json:"text"`
		} `json:"arguments"`
	}
	_ = json.Unmarshal(m.Params, &p)
	if p.Name != "reply" {
		s.write(map[string]any{"jsonrpc": "2.0", "id": rawOrNull(m.ID),
			"error": map[string]any{"code": -32602, "message": "unknown tool: " + p.Name}})
		return
	}
	text := strings.TrimSpace(p.Arguments.Text)
	if text != "" {
		_ = inbox.Send(s.inboxDir, s.agent, s.peer, text)
		_ = journal.Append(s.journalDir, s.agent, "msg.send", map[string]string{"to": s.peer, "via": "channel"})
	}
	s.result(m.ID, map[string]any{
		"content": []any{map[string]any{"type": "text", "text": "sent to " + s.peer}},
	})
}

// pushEvent emits a channel notification into the Claude session.
func (s *Server) pushEvent(content string) {
	s.write(map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/claude/channel",
		"params": map[string]any{
			"content": content,
			"meta":    map[string]any{"from": s.peer},
		},
	})
}

func (s *Server) pollInbox(stop <-chan struct{}) {
	t := time.NewTicker(s.poll)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			content, err := inbox.Read(s.inboxDir, s.agent, false)
			if err == nil && strings.TrimSpace(content) != "" {
				_ = journal.Append(s.journalDir, s.agent, "msg.read", map[string]string{"via": "channel"})
				s.pushEvent(content)
			}
		}
	}
}
