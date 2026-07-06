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

	for _, entry := range pre {
		em, _ := entry.(map[string]any)
		hs, _ := em["hooks"].([]any)
		for _, h := range hs {
			if hm, ok := h.(map[string]any); ok && isCoactHook(hm) {
				return false, nil
			}
		}
	}

	entry := map[string]any{
		"matcher": "Edit|Write|MultiEdit|NotebookEdit",
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": coactBinary(),
				"args":    []any{"hook", "claude"},
			},
		},
	}
	hooks["PreToolUse"] = append(pre, entry)
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
	if c, _ := hm["command"].(string); strings.Contains(c, "coact") && strings.Contains(c, "hook") {
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
