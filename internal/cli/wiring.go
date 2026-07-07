package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/platform"
	"github.com/tianyi-zhang-02/coact/internal/setup"
)

const (
	coactBegin = "<!-- coact:begin -->"
	coactEnd   = "<!-- coact:end -->"
)

// This file holds coact's removal + inspection helpers for wiring (used by
// deinit and doctor). The idempotent init-time wiring (ensureClaudeHook,
// ensureMarkedBlock) lives in internal/setup, which the CLI and the UI both call
// — keeping the safety-critical hook format in exactly one place.

// removeMarkedBlock removes the coact-marked block from path. Returns true if it
// changed the file.
func removeMarkedBlock(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	content := string(data)
	start := strings.Index(content, coactBegin)
	end := strings.Index(content, coactEnd)
	if start < 0 || end < start {
		return false, nil
	}
	end += len(coactEnd)

	head := strings.TrimRight(content[:start], " \t\n")
	tail := strings.TrimLeft(content[end:], "\n")
	result := head
	if head != "" && tail != "" {
		result += "\n\n" + tail
	} else {
		result += tail
	}
	if result != "" && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return true, platform.AtomicWrite(path, []byte(result), 0o644)
}

func ensureMarkedBlock(path, body string) (bool, error) {
	return setup.EnsureMarkedBlock(path, body)
}

func ensureClaudeHook(root string) (bool, error) {
	return setup.EnsureClaudeHook(root)
}

// removeClaudeHook removes coact's PreToolUse hook from .claude/settings.json,
// pruning now-empty matcher entries and the hooks map. Returns true if changed.
func removeClaudeHook(root string) (bool, error) {
	settingsPath := filepath.Join(root, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	var settings map[string]any
	if json.Unmarshal(data, &settings) != nil {
		return false, nil
	}
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		return false, nil
	}
	pre, _ := hooks["PreToolUse"].([]any)

	changed := false
	var kept []any
	for _, entry := range pre {
		em, _ := entry.(map[string]any)
		hs, _ := em["hooks"].([]any)
		var keptHooks []any
		for _, h := range hs {
			if hm, ok := h.(map[string]any); ok && isCoactHook(hm) {
				changed = true
				continue
			}
			keptHooks = append(keptHooks, h)
		}
		if len(keptHooks) == 0 {
			continue // drop matcher entry that only held coact
		}
		em["hooks"] = keptHooks
		kept = append(kept, em)
	}
	if !changed {
		return false, nil
	}
	if len(kept) == 0 {
		delete(hooks, "PreToolUse")
	} else {
		hooks["PreToolUse"] = kept
	}
	if len(hooks) == 0 {
		delete(settings, "hooks")
	} else {
		settings["hooks"] = hooks
	}
	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return false, err
	}
	out = append(out, '\n')
	return true, platform.AtomicWrite(settingsPath, out, 0o644)
}

// wiredHookCommand returns the command + args of the coact hook wired into
// .claude/settings.json, if present.
func wiredHookCommand(root string) (cmd string, args []string, found bool) {
	data, err := os.ReadFile(filepath.Join(root, ".claude", "settings.json"))
	if err != nil {
		return "", nil, false
	}
	var settings map[string]any
	if json.Unmarshal(data, &settings) != nil {
		return "", nil, false
	}
	hooks, _ := settings["hooks"].(map[string]any)
	pre, _ := hooks["PreToolUse"].([]any)
	for _, entry := range pre {
		em, _ := entry.(map[string]any)
		hs, _ := em["hooks"].([]any)
		for _, h := range hs {
			hm, _ := h.(map[string]any)
			if !isCoactHook(hm) {
				continue
			}
			cmd, _ = hm["command"].(string)
			if raw, ok := hm["args"].([]any); ok {
				for _, x := range raw {
					if s, ok := x.(string); ok {
						args = append(args, s)
					}
				}
			}
			return cmd, args, true
		}
	}
	return "", nil, false
}

func isCoactHook(hm map[string]any) bool {
	// String form: "<coact> hook claude" — match the subcommand signature
	// (path-independent; "hook claude" won't collide with e.g. "claude-hook.sh").
	if c, _ := hm["command"].(string); strings.Contains(c, "hook claude") {
		return true
	}
	if args, ok := hm["args"].([]any); ok {
		var hasHook, hasClaude bool
		for _, a := range args {
			switch s, _ := a.(string); s {
			case "hook":
				hasHook = true
			case "claude":
				hasClaude = true
			}
		}
		if hasHook && hasClaude {
			return true
		}
	}
	return false
}

// coactBinary returns an absolute path to the running coact binary so the wired
// hook works whether or not coact is on PATH. Falls back to "coact".
func coactBinary() string {
	if exe, err := os.Executable(); err == nil {
		if abs, err := filepath.Abs(exe); err == nil {
			return abs
		}
		return exe
	}
	return "coact"
}

func channelServerName(agent string) string { return "coact-" + agent }

// channelMCPInstalled reports whether the coact channel for agent is registered
// in the repo's .mcp.json.
func channelMCPInstalled(root, agent string) bool {
	data, err := os.ReadFile(filepath.Join(root, ".mcp.json"))
	if err != nil {
		return false
	}
	var m map[string]any
	if json.Unmarshal(data, &m) != nil {
		return false
	}
	servers, _ := m["mcpServers"].(map[string]any)
	_, ok := servers[channelServerName(agent)]
	return ok
}

// ensureChannelMCP registers the coact channel for agent in .mcp.json (merging
// with any existing servers). Returns true if it added the entry.
func ensureChannelMCP(root, agent string) (bool, error) {
	path := filepath.Join(root, ".mcp.json")
	settings := map[string]any{}
	if data, err := os.ReadFile(path); err == nil && len(strings.TrimSpace(string(data))) > 0 {
		if err := json.Unmarshal(data, &settings); err != nil {
			return false, fmt.Errorf(".mcp.json is not valid JSON: %w", err)
		}
	}
	servers, _ := settings["mcpServers"].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	name := channelServerName(agent)
	if _, exists := servers[name]; exists {
		return false, nil
	}
	servers[name] = map[string]any{
		"command": coactBinary(),
		"args":    []any{"channel", agent},
	}
	settings["mcpServers"] = servers

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return false, err
	}
	data = append(data, '\n')
	if err := platform.AtomicWrite(path, data, 0o644); err != nil {
		return false, err
	}
	return true, nil
}

// removeChannelMCP removes coact channel entries from .mcp.json, pruning the
// map when empty. Returns true if it changed the file.
func removeChannelMCP(root string) (bool, error) {
	path := filepath.Join(root, ".mcp.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	var settings map[string]any
	if json.Unmarshal(data, &settings) != nil {
		return false, nil
	}
	servers, _ := settings["mcpServers"].(map[string]any)
	changed := false
	for name := range servers {
		if strings.HasPrefix(name, "coact-") {
			delete(servers, name)
			changed = true
		}
	}
	if !changed {
		return false, nil
	}
	if len(servers) == 0 {
		delete(settings, "mcpServers")
	} else {
		settings["mcpServers"] = servers
	}
	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return false, err
	}
	out = append(out, '\n')
	return true, platform.AtomicWrite(path, out, 0o644)
}
