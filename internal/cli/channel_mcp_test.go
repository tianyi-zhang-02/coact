package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureChannelMCPRoundTrip(t *testing.T) {
	root := t.TempDir()

	added, err := ensureChannelMCP(root, "claude")
	if err != nil || !added {
		t.Fatalf("first install: added=%v err=%v", added, err)
	}
	if !channelMCPInstalled(root, "claude") {
		t.Fatal("channel should be installed")
	}
	if a, _ := ensureChannelMCP(root, "claude"); a {
		t.Fatal("second install should be a no-op")
	}

	removed, err := removeChannelMCP(root)
	if err != nil || !removed {
		t.Fatalf("remove: removed=%v err=%v", removed, err)
	}
	if channelMCPInstalled(root, "claude") {
		t.Fatal("channel should be gone after remove")
	}
}

func TestEnsureChannelMCPMergesExisting(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".mcp.json"),
		[]byte(`{"mcpServers":{"other":{"command":"foo"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ensureChannelMCP(root, "claude"); err != nil {
		t.Fatal(err)
	}
	servers := readServers(t, root)
	if _, ok := servers["other"]; !ok {
		t.Error("unrelated server was clobbered")
	}
	if _, ok := servers["coact-claude"]; !ok {
		t.Error("coact-claude not added")
	}

	// remove only touches coact-* entries
	if _, err := removeChannelMCP(root); err != nil {
		t.Fatal(err)
	}
	servers = readServers(t, root)
	if _, ok := servers["other"]; !ok {
		t.Error("remove clobbered the unrelated server")
	}
	if _, ok := servers["coact-claude"]; ok {
		t.Error("coact-claude not removed")
	}
}

func readServers(t *testing.T, root string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	servers, _ := m["mcpServers"].(map[string]any)
	return servers
}
