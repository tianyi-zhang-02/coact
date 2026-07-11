package cli

import (
	"testing"

	"github.com/tianyi-zhang-02/coact/internal/config"
)

func TestMissingProtectedPathsFindsSecurityDrift(t *testing.T) {
	cfg := config.Default()
	cfg.Policy.ProtectedPaths = []string{".coact/config.json"}
	missing := missingProtectedPaths(cfg)
	if len(missing) == 0 {
		t.Fatal("expected missing protected paths")
	}
	foundUsage := false
	for _, path := range missing {
		if path == ".coact/usage/**" {
			foundUsage = true
		}
	}
	if !foundUsage {
		t.Fatalf("usage protection missing from doctor audit: %#v", missing)
	}
}
