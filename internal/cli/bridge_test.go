package cli

import (
	"bufio"
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/tianyi-zhang-02/coact/internal/codex"
	"github.com/tianyi-zhang-02/coact/internal/inbox"
)

// fakeCodex answers the handshake and streams a reply on turn/start.
func fakeCodex(conn net.Conn) {
	r := bufio.NewReader(conn)
	c := codex.NewlineCodec{}
	write := func(v any) { b, _ := json.Marshal(v); _ = c.WriteMessage(conn, b) }
	for {
		msg, err := c.ReadMessage(r)
		if len(msg) > 0 {
			var req struct {
				ID     int    `json:"id"`
				Method string `json:"method"`
			}
			_ = json.Unmarshal(msg, &req)
			switch req.Method {
			case "initialize":
				write(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]any{"userAgent": "codex_cli_rs/0.139.0"}})
			case "thread/start":
				write(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]any{"thread": map[string]any{"id": "t1"}}})
			case "turn/start":
				write(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]any{"turn": map[string]any{"id": "turn1"}}})
				write(map[string]any{"jsonrpc": "2.0", "method": "turn/started", "params": map[string]any{"turn": map[string]any{"id": "turn1"}}})
				write(map[string]any{"jsonrpc": "2.0", "method": "item/agentMessage/delta", "params": map[string]any{"itemId": "i1", "delta": "Hello "}})
				write(map[string]any{"jsonrpc": "2.0", "method": "item/agentMessage/delta", "params": map[string]any{"itemId": "i1", "delta": "world"}})
				write(map[string]any{"jsonrpc": "2.0", "method": "item/completed", "params": map[string]any{"item": map[string]any{"id": "i1"}}})
			}
		}
		if err != nil {
			return
		}
	}
}

func TestCodexBridgeRoundTrip(t *testing.T) {
	p := setupProject(t)

	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()
	go fakeCodex(c2)

	// onMsg mirrors bridgeCodex: Codex output → Claude's inbox.
	onMsg := func(text string) { _ = inbox.Send(p.InboxDir(), "codex", "claude", text) }
	client := codex.New(c1, codex.NewlineCodec{}, onMsg)
	if _, err := client.Initialize(); err != nil {
		t.Fatal(err)
	}
	if _, err := client.StartThread("/repo"); err != nil {
		t.Fatal(err)
	}

	stop := make(chan struct{})
	go runCodexBridge(p, client, 20*time.Millisecond, stop)
	defer close(stop)

	// Claude sends Codex a message (as the channel's reply tool would).
	if err := inbox.Send(p.InboxDir(), "claude", "codex", "build the gateway"); err != nil {
		t.Fatal(err)
	}

	// The bridge forwards it to Codex, whose reply must land in Claude's inbox.
	deadline := time.After(3 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("Codex reply never reached Claude's inbox")
		default:
			if got, _ := inbox.Read(p.InboxDir(), "claude", true); strings.Contains(got, "Hello world") {
				return
			}
			time.Sleep(25 * time.Millisecond)
		}
	}
}

func TestBridgeTextStripsHeaders(t *testing.T) {
	in := "### from claude · 2026-07-07T00:00:00Z\nbuild the gateway\n"
	if got := bridgeText(in); got != "build the gateway" {
		t.Fatalf("bridgeText = %q", got)
	}
}
