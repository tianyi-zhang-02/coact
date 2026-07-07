package codex

import (
	"bufio"
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"
)

// fakeCodex is a minimal Codex app-server test double: it answers the handshake,
// streams an agent message on turn/start, and answers turn/steer with another.
func fakeCodex(conn net.Conn) {
	r := bufio.NewReader(conn)
	codec := NewlineCodec{}
	write := func(v any) { b, _ := json.Marshal(v); _ = codec.WriteMessage(conn, b) }
	for {
		msg, err := codec.ReadMessage(r)
		if len(msg) > 0 {
			var req struct {
				ID     int    `json:"id"`
				Method string `json:"method"`
			}
			_ = json.Unmarshal(msg, &req)
			switch req.Method {
			case "initialize":
				write(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]any{"userAgent": "codex_cli_rs/0.139.0 (Mac OS 15; arm64)"}})
			case "thread/start":
				write(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]any{"thread": map[string]any{"id": "t1"}}})
			case "turn/start":
				write(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]any{"turn": map[string]any{"id": "turn1"}}})
				write(map[string]any{"jsonrpc": "2.0", "method": "turn/started", "params": map[string]any{"turn": map[string]any{"id": "turn1"}}})
				write(map[string]any{"jsonrpc": "2.0", "method": "item/agentMessage/delta", "params": map[string]any{"itemId": "i1", "delta": "Hello "}})
				write(map[string]any{"jsonrpc": "2.0", "method": "item/agentMessage/delta", "params": map[string]any{"itemId": "i1", "delta": "world"}})
				write(map[string]any{"jsonrpc": "2.0", "method": "item/completed", "params": map[string]any{"item": map[string]any{"id": "i1"}}})
				// intentionally no turn/completed: the turn stays active so a
				// follow-up Send() must steer.
			case "turn/steer":
				write(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]any{}})
				write(map[string]any{"jsonrpc": "2.0", "method": "item/agentMessage/delta", "params": map[string]any{"itemId": "i2", "delta": "steered ok"}})
				write(map[string]any{"jsonrpc": "2.0", "method": "item/completed", "params": map[string]any{"item": map[string]any{"id": "i2"}}})
			}
		}
		if err != nil {
			return
		}
	}
}

func recv(t *testing.T, ch chan string) string {
	t.Helper()
	select {
	case s := <-ch:
		return s
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for a Codex message")
		return ""
	}
}

func TestClientAgainstFakeCodex(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()
	go fakeCodex(c2)

	msgs := make(chan string, 8)
	client := New(c1, NewlineCodec{}, func(s string) { msgs <- s })
	defer client.Close()

	ua, err := client.Initialize()
	if err != nil || !strings.Contains(ua, "0.139") {
		t.Fatalf("initialize = %q, %v", ua, err)
	}
	if tid, err := client.StartThread("/repo"); err != nil || tid != "t1" {
		t.Fatalf("StartThread = %q, %v", tid, err)
	}
	if turn, err := client.StartTurn("build the gateway"); err != nil || turn != "turn1" {
		t.Fatalf("StartTurn = %q, %v", turn, err)
	}

	// Streamed deltas are assembled into one complete message.
	if got := recv(t, msgs); got != "Hello world" {
		t.Fatalf("first message = %q, want %q", got, "Hello world")
	}

	// The turn is still active, so Send must steer it (not start a new turn).
	if err := client.Send("also add tests"); err != nil {
		t.Fatalf("Send (steer): %v", err)
	}
	if got := recv(t, msgs); got != "steered ok" {
		t.Fatalf("steer message = %q, want %q", got, "steered ok")
	}
}
