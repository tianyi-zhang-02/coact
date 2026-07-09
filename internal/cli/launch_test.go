//go:build !windows

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tianyi-zhang-02/coact/internal/config"
	"github.com/tianyi-zhang-02/coact/internal/lockmgr"
	"github.com/tianyi-zhang-02/coact/internal/presence"
)

func writeStub(t *testing.T, script string) string {
	t.Helper()
	stub := filepath.Join(t.TempDir(), "stub.sh")
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return stub
}

func TestRunWrappedReleasesLocksAndStopsPresence(t *testing.T) {
	p := setupProject(t)
	cfg, _ := config.Load(p.ConfigPath())
	m := lockmgr.New(p, cfg)
	if _, err := m.Acquire("claude", filepath.Join(p.Root, "src")); err != nil {
		t.Fatal(err)
	}

	if code := runWrapped(p, cfg, "claude", writeStub(t, "#!/bin/sh\nexit 0\n"), nil, ""); code != 0 {
		t.Fatalf("runWrapped exit code = %d", code)
	}

	locks, _ := m.List()
	for _, lk := range locks {
		if lk.Owner == "claude" {
			t.Fatal("claude locks were not released on session exit")
		}
	}
	s, err := presence.Read(p.SessionDir(), "claude")
	if err != nil || s.Status != "stopped" {
		t.Fatalf("presence should be stopped, got %+v (err %v)", s, err)
	}
}

func TestRunWrappedPropagatesExitCode(t *testing.T) {
	p := setupProject(t)
	cfg, _ := config.Load(p.ConfigPath())
	if code := runWrapped(p, cfg, "claude", writeStub(t, "#!/bin/sh\nexit 7\n"), nil, ""); code != 7 {
		t.Fatalf("want exit code 7, got %d", code)
	}
}

func TestCoactAgentEnvExposesCurrentBinary(t *testing.T) {
	env := coactAgentEnv([]string{"PATH=/usr/bin", "COACT_AGENT=old"}, "codex")
	got := map[string]string{}
	for _, item := range env {
		if k, v, ok := strings.Cut(item, "="); ok {
			got[k] = v
		}
	}
	if got["COACT_AGENT"] != "codex" {
		t.Fatalf("COACT_AGENT = %q, want codex", got["COACT_AGENT"])
	}
	if got["COACT_BIN"] == "" {
		t.Fatal("COACT_BIN should point at the current coact binary")
	}
	binDir := filepath.Dir(got["COACT_BIN"])
	pathParts := strings.Split(got["PATH"], string(os.PathListSeparator))
	if len(pathParts) == 0 || pathParts[0] != binDir {
		t.Fatalf("PATH should start with current binary dir %q, got %q", binDir, got["PATH"])
	}
}
