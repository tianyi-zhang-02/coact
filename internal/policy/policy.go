// Package policy decides whether an agent may write a given path, independent
// of who currently holds a lock. It enforces two things: protected paths that
// no agent may write (they require a human), and optional per-agent write
// scopes. This is the capability layer of coact's governance.
package policy

import (
	"path"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/config"
)

// Decision is the outcome of a policy check.
type Decision struct {
	Allowed bool
	Reason  string
}

// Engine evaluates write permission from a project's config.
type Engine struct {
	agents    map[string][]string // agent id -> write-scope globs
	protected []string
}

// New builds an Engine from config. A nil config yields an allow-all engine.
func New(cfg *config.Config) *Engine {
	e := &Engine{agents: map[string][]string{}}
	if cfg == nil {
		return e
	}
	for _, a := range cfg.Agents {
		e.agents[a.ID] = a.Write
	}
	e.protected = cfg.Policy.ProtectedPaths
	return e
}

// Check reports whether agent may write relPath (repo-relative, slash form).
func (e *Engine) Check(agent, relPath string) Decision {
	for _, p := range e.protected {
		if Match(p, relPath) {
			return Decision{
				Allowed: false,
				Reason:  relPath + " is a protected path — it requires a human (see policy.protected_paths in .coact/config.json)",
			}
		}
	}
	if globs, ok := e.agents[agent]; ok && len(globs) > 0 {
		for _, g := range globs {
			if Match(g, relPath) {
				return Decision{Allowed: true}
			}
		}
		return Decision{
			Allowed: false,
			Reason:  relPath + " is outside " + agent + "'s write scope (see agents[].write in .coact/config.json)",
		}
	}
	return Decision{Allowed: true}
}

// Protected returns the configured protected-path globs.
func (e *Engine) Protected() []string { return e.protected }

// Scope returns the write-scope globs for an agent (nil = unrestricted).
func (e *Engine) Scope(agent string) []string { return e.agents[agent] }

// Match reports whether a glob pattern matches a repo-relative slash path. It
// supports ** (any number of path segments), * and ? within a segment, and
// treats a wildcard-free pattern as a literal that also matches its subtree
// (so ".coact" matches ".coact/config.json").
func Match(pattern, name string) bool {
	pattern = strings.TrimSuffix(strings.TrimPrefix(pattern, "./"), "/")
	name = strings.TrimPrefix(name, "./")
	if pattern == "" {
		return false
	}
	if !strings.ContainsAny(pattern, "*?[") {
		return name == pattern || strings.HasPrefix(name, pattern+"/")
	}
	return matchSegments(strings.Split(pattern, "/"), strings.Split(name, "/"))
}

func matchSegments(pat, name []string) bool {
	if len(pat) == 0 {
		return len(name) == 0
	}
	if pat[0] == "**" {
		for i := 0; i <= len(name); i++ {
			if matchSegments(pat[1:], name[i:]) {
				return true
			}
		}
		return false
	}
	if len(name) == 0 {
		return false
	}
	if ok, _ := path.Match(pat[0], name[0]); !ok {
		return false
	}
	return matchSegments(pat[1:], name[1:])
}
