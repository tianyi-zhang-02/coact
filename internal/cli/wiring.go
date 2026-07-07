package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/platform"
)

const (
	coactBegin = "<!-- coact:begin -->"
	coactEnd   = "<!-- coact:end -->"
)

// ensureMarkedBlock idempotently appends body (wrapped in coact markers) to the
// file at path, creating it if needed. Returns true if it changed the file.
func ensureMarkedBlock(path, body string) (bool, error) {
	var content string
	if data, err := os.ReadFile(path); err == nil {
		content = string(data)
	} else if !os.IsNotExist(err) {
		return false, err
	}
	if strings.Contains(content, coactBegin) {
		return false, nil // already present
	}
	var b strings.Builder
	b.WriteString(content)
	if content != "" {
		if !strings.HasSuffix(content, "\n") {
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	b.WriteString(coactBegin + "\n")
	b.WriteString(strings.TrimRight(body, "\n") + "\n")
	b.WriteString(coactEnd + "\n")
	return true, platform.AtomicWrite(path, []byte(b.String()), 0o644)
}

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

// ensureClaudeHook idempotently wires coact's PreToolUse gate into the repo's
// .claude/settings.json, merging with any existing settings. Returns true if it
// added the hook.
func ensureClaudeHook(root string) (bool, error) {
	claudeDir := filepath.Join(root, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		return false, err
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")

	settings := map[string]any{}
	if data, err := os.ReadFile(settingsPath); err == nil && len(strings.TrimSpace(string(data))) > 0 {
		if err := json.Unmarshal(data, &settings); err != nil {
			return false, fmt.Errorf(".claude/settings.json is not valid JSON: %w", err)
		}
	}

	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}
	pre, _ := hooks["PreToolUse"].([]any)

	// The command MUST be a single shell string ("<coact> hook claude"). The
	// args-array exec form is ignored by older Claude Code (1.0.x runs the bare
	// binary → prints usage → exit 1 → non-blocking, so edits are NOT gated).
	desired := coactBinary() + " hook claude"

	present := false
	dirty := false
	var kept []any
	for _, entry := range pre {
		em, _ := entry.(map[string]any)
		hs, _ := em["hooks"].([]any)
		var keptHooks []any
		for _, h := range hs {
			hm, ok := h.(map[string]any)
			if ok && isCoactHook(hm) {
				c, _ := hm["command"].(string)
				_, hasArgs := hm["args"]
				if c == desired && !hasArgs && !present {
					present = true
					keptHooks = append(keptHooks, h)
				} else {
					dirty = true // outdated / duplicate coact hook → drop & migrate
				}
				continue
			}
			keptHooks = append(keptHooks, h)
		}
		if len(keptHooks) > 0 {
			em["hooks"] = keptHooks
			kept = append(kept, em)
		} else {
			dirty = true // matcher entry held only coact hooks; drop it
		}
	}

	if present && !dirty {
		return false, nil
	}
	if !present {
		kept = append(kept, map[string]any{
			"matcher": "Edit|Write|MultiEdit|NotebookEdit",
			"hooks":   []any{map[string]any{"type": "command", "command": desired}},
		})
	}
	hooks["PreToolUse"] = kept
	settings["hooks"] = hooks

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return false, err
	}
	data = append(data, '\n')
	if err := platform.AtomicWrite(settingsPath, data, 0o644); err != nil {
		return false, err
	}
	return true, nil
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
