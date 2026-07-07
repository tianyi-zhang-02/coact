package channel

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tianyi-zhang-02/coact/internal/inbox"
)

func TestInitializeToolsAndReply(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	s := New("claude", "codex", dir, dir, strings.NewReader(""), &out)

	s.dispatch([]byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}`))
	s.dispatch([]byte(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`))
	s.dispatch([]byte(`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"reply","arguments":{"text":"hi codex"}}}`))

	o := out.String()
	if !strings.Contains(o, `"claude/channel"`) {
		t.Fatalf("initialize missing channel capability:\n%s", o)
	}
	if !strings.Contains(o, `"reply"`) {
		t.Fatalf("tools/list missing reply tool:\n%s", o)
	}
	if !strings.Contains(o, "sent to codex") {
		t.Fatalf("reply not confirmed:\n%s", o)
	}

	// The reply must have been delivered to codex's inbox.
	got, _ := inbox.Read(dir, "codex", true)
	if !strings.Contains(got, "hi codex") {
		t.Fatalf("reply not delivered to codex inbox: %q", got)
	}

	// Every output line must be valid JSON-RPC.
	for _, line := range strings.Split(strings.TrimSpace(o), "\n") {
		var m map[string]any
		if json.Unmarshal([]byte(line), &m) != nil || m["jsonrpc"] != "2.0" {
			t.Fatalf("invalid JSON-RPC output line: %q", line)
		}
	}
}

func TestPushEventShape(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	s := New("claude", "codex", dir, dir, strings.NewReader(""), &out)
	s.pushEvent("hello from codex")

	var m struct {
		Method string `json:"method"`
		Params struct {
			Content string            `json:"content"`
			Meta    map[string]string `json:"meta"`
		} `json:"params"`
	}
	if json.Unmarshal(out.Bytes(), &m) != nil {
		t.Fatalf("push event not JSON: %s", out.String())
	}
	if m.Method != "notifications/claude/channel" {
		t.Fatalf("wrong method: %s", m.Method)
	}
	if m.Params.Content != "hello from codex" || m.Params.Meta["from"] != "codex" {
		t.Fatalf("bad params: %+v", m.Params)
	}
}
