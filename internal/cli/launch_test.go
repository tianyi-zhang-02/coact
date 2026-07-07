//go:build !windows

package cli

import (
	"os"
	"path/filepath"
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
