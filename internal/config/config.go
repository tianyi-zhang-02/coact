// Package config defines the machine-readable coact project configuration.
//
// The implementation uses JSON (.coact/config.json) to keep the core free of
// external dependencies; the SPEC shows the same schema in YAML for
// readability. See docs/STACK.md.
package config

import (
	"encoding/json"
	"os"

	"github.com/tianyi-zhang-02/coact/internal/platform"
)

type Config struct {
	Version  string         `json:"version"`
	Mode     string         `json:"mode"` // shared | worktree
	Agents   []AgentConfig  `json:"agents"`
	Locks    LockConfig     `json:"locks"`
	Presence PresenceConfig `json:"presence"`
	Policy   PolicyConfig   `json:"policy"`
}

type AgentConfig struct {
	ID      string   `json:"id"`
	Adapter string   `json:"adapter"`
	Write   []string `json:"write,omitempty"` // globs this agent may write; empty = unrestricted
}

type LockConfig struct {
	DefaultTTLSeconds      int `json:"default_ttl_seconds"`
	HeartbeatDivisor       int `json:"heartbeat_divisor"`
	RegistryLockTTLSeconds int `json:"registry_lock_ttl_seconds"`
}

type PresenceConfig struct {
	Mode            string `json:"mode"` // sidecar | hook-only
	TTLSeconds      int    `json:"ttl_seconds"`
	IntervalSeconds int    `json:"interval_seconds"`
}

type PolicyConfig struct {
	OnConflict        string   `json:"on_conflict"` // block | queue | warn
	AllowStealExpired bool     `json:"allow_steal_expired"`
	ProtectedPaths    []string `json:"protected_paths"`
}

// Default returns the built-in configuration used by `coact init`.
func Default() *Config {
	return &Config{
		Version: "0.3",
		Mode:    "shared",
		Agents: []AgentConfig{
			{ID: "claude", Adapter: "claude-code"},
			{ID: "codex", Adapter: "codex"},
			{ID: "antigravity", Adapter: "antigravity-cli"},
		},
		Locks: LockConfig{
			DefaultTTLSeconds:      900,
			HeartbeatDivisor:       3,
			RegistryLockTTLSeconds: 5,
		},
		Presence: PresenceConfig{
			Mode:            "sidecar",
			TTLSeconds:      60,
			IntervalSeconds: 20,
		},
		Policy: PolicyConfig{
			OnConflict:        "block",
			AllowStealExpired: true,
			// Protect coact's machine-mutated coordination state from direct
			// rewrites. Planning runs and local memory remain editable through
			// normal locks so agents can write proposals and durable notes.
			ProtectedPaths: []string{
				".coact/config.json",
				".coact/board.md",
				".coact/locks/**",
				".coact/session/**",
				".coact/journal/**",
				".coact/inbox/**",
				".coact/terminal/**",
				".coact/tasks/**",
				".coact/usage/**",
				".coact/evaluations/**",
			},
		},
	}
}

// Load reads config from path. A missing file yields the default config so
// read-only commands work even before `coact init`.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}
	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Save writes config to path atomically.
func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return platform.AtomicWrite(path, data, 0o644)
}
