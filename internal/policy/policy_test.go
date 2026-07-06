package policy

import (
	"testing"

	"github.com/tianyi-zhang-02/coact/internal/config"
)

func TestMatch(t *testing.T) {
	cases := []struct {
		pat, name string
		want      bool
	}{
		{".coact/**", ".coact/config.json", true},
		{".coact/**", ".coact/locks/ab.lock", true},
		{".coact", ".coact/config.json", true}, // literal covers subtree
		{".coact", ".coactother/x", false},     // not a false prefix match
		{"*.png", "logo.png", true},
		{"*.png", "src/logo.png", false}, // * does not cross segments
		{"assets/**", "assets/img/logo.png", true},
		{"assets/**", "assets", true}, // ** matches zero segments
		{"src/**/*.go", "src/a/b/x.go", true},
		{"src/**/*.go", "src/x.go", true},
		{"src/*", "src/x", true},
		{"src/*", "src/x/y", false},
		{"web", "src/web", false},
	}
	for _, c := range cases {
		if got := Match(c.pat, c.name); got != c.want {
			t.Errorf("Match(%q,%q)=%v want %v", c.pat, c.name, got, c.want)
		}
	}
}

func TestCheckProtected(t *testing.T) {
	e := New(config.Default()) // default protects .coact/**
	if d := e.Check("claude", ".coact/config.json"); d.Allowed {
		t.Error("writing a protected path should be denied")
	}
	if d := e.Check("claude", "src/main.go"); !d.Allowed {
		t.Errorf("writing a normal path should be allowed: %s", d.Reason)
	}
}

func TestCheckPerAgentScope(t *testing.T) {
	cfg := config.Default()
	cfg.Agents = []config.AgentConfig{
		{ID: "design", Write: []string{"assets/**", "*.md"}},
		{ID: "backend"}, // unrestricted
	}
	e := New(cfg)

	if d := e.Check("design", "assets/logo.png"); !d.Allowed {
		t.Errorf("design writing assets should be allowed: %s", d.Reason)
	}
	if d := e.Check("design", "src/main.go"); d.Allowed {
		t.Error("design writing src should be denied (outside scope)")
	}
	if d := e.Check("backend", "src/main.go"); !d.Allowed {
		t.Errorf("unrestricted agent should be allowed: %s", d.Reason)
	}
	if d := e.Check("backend", ".coact/config.json"); d.Allowed {
		t.Error("protected path denied even for an unrestricted agent")
	}
}
