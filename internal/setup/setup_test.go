package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tianyi-zhang-02/coact/internal/config"
)

func TestEnsureGitignoreIncludesRuntimeState(t *testing.T) {
	root := t.TempDir()

	ensureGitignore(root)

	data, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	content := string(data)
	for _, want := range []string{
		".coact/locks/",
		".coact/session/",
		".coact/journal/",
		".coact/inbox/",
		".coact/terminal/",
		".coact/runs/",
		".coact/memory/",
		".coact/usage/",
		".coact/evaluations/",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf(".gitignore missing %q in:\n%s", want, content)
		}
	}
}

func TestEnsureMarkedBlockUpdatesExistingContract(t *testing.T) {
	path := filepath.Join(t.TempDir(), "AGENTS.md")
	initial := "user text\n\n<!-- coact:begin -->\nold contract\n<!-- coact:end -->\n"
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}
	changed, err := ensureMarkedBlock(path, "new contract")
	if err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "user text") || !strings.Contains(content, "new contract") || strings.Contains(content, "old contract") {
		t.Fatalf("unexpected updated contract:\n%s", content)
	}
}

func TestInitializeUsesAntigravityAsThirdDefault(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, "human"); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(filepath.Join(root, ".coact", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Agents) != 3 || cfg.Agents[2].ID != "antigravity" {
		t.Fatalf("default agents = %#v", cfg.Agents)
	}
	contract, err := os.ReadFile(filepath.Join(root, "ANTIGRAVITY.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contract), "COACT_AGENT=antigravity") {
		t.Fatalf("unexpected Antigravity contract:\n%s", contract)
	}
	if _, err := os.Stat(filepath.Join(root, "GEMINI.md")); !os.IsNotExist(err) {
		t.Fatalf("fresh workspace should not wire Gemini by default: %v", err)
	}
}

func TestEnsureGitignoreDoesNotDuplicateWhenCoactDirIgnored(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".gitignore")
	initial := "/.coact/\n"
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	ensureGitignore(root)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != initial {
		t.Fatalf("whole .coact ignore should already cover runtime paths; got:\n%s", string(data))
	}
}

func TestMigrateLegacyProtectedCoactGlob(t *testing.T) {
	cfg := config.Default()
	cfg.Policy.ProtectedPaths = []string{".coact/**"}

	if !migrateProtectedPaths(cfg) {
		t.Fatal("expected legacy .coact/** policy to migrate")
	}
	for _, protected := range cfg.Policy.ProtectedPaths {
		if protected == ".coact/**" {
			t.Fatalf("legacy broad protection should be removed: %#v", cfg.Policy.ProtectedPaths)
		}
	}
	if !containsString(cfg.Policy.ProtectedPaths, ".coact/config.json") {
		t.Fatalf("config should remain protected: %#v", cfg.Policy.ProtectedPaths)
	}
}

func TestMigrateVersionAddsNewProtectedStateWithoutDroppingCustomPaths(t *testing.T) {
	cfg := config.Default()
	cfg.Version = "0.1"
	cfg.Policy.ProtectedPaths = []string{"secrets/**"}

	if !migrateProtectedPaths(cfg) {
		t.Fatal("expected old config version to migrate")
	}
	for _, want := range []string{"secrets/**", ".coact/usage/**", ".coact/evaluations/**"} {
		if !containsString(cfg.Policy.ProtectedPaths, want) {
			t.Fatalf("migration missing %q: %#v", want, cfg.Policy.ProtectedPaths)
		}
	}
	if cfg.Version != config.Default().Version {
		t.Fatalf("version = %q", cfg.Version)
	}
}

func TestMigrateRepairsMissingRequiredPathAtCurrentVersion(t *testing.T) {
	cfg := config.Default()
	cfg.Policy.ProtectedPaths = []string{".coact/config.json"}
	if !migrateProtectedPaths(cfg) {
		t.Fatal("expected missing mandatory protection to be repaired")
	}
	if !containsString(cfg.Policy.ProtectedPaths, ".coact/evaluations/**") {
		t.Fatalf("evaluation protection was not restored: %#v", cfg.Policy.ProtectedPaths)
	}
}

func TestMigrateRetiredGeminiAgentToAntigravity(t *testing.T) {
	cfg := config.Default()
	cfg.Agents = []config.AgentConfig{
		{ID: "claude", Adapter: "claude-code"},
		{ID: "codex", Adapter: "codex"},
		{ID: "gemini", Adapter: "gemini-cli"},
	}
	if !migrateRetiredGeminiAgent(cfg) {
		t.Fatal("expected the old built-in agent layout to migrate")
	}
	if got := cfg.Agents[2]; got.ID != "antigravity" || got.Adapter != "antigravity-cli" {
		t.Fatalf("third agent = %+v", got)
	}
}

func TestMigrateRetiredGeminiPreservesWriteRestrictions(t *testing.T) {
	cfg := config.Default()
	cfg.Agents = []config.AgentConfig{
		{ID: "claude", Adapter: "claude-code"},
		{ID: "codex", Adapter: "codex"},
		{ID: "gemini", Adapter: "gemini-cli", Write: []string{"research/**"}},
	}
	if !migrateRetiredGeminiAgent(cfg) {
		t.Fatal("custom retired agent should migrate")
	}
	if cfg.Agents[2].ID != "antigravity" || cfg.Agents[2].Adapter != "antigravity-cli" || len(cfg.Agents[2].Write) != 1 || cfg.Agents[2].Write[0] != "research/**" {
		t.Fatalf("custom restrictions were not preserved: %+v", cfg.Agents[2])
	}
}

func TestMigrateRetiredGeminiDropsDuplicateWhenAntigravityExists(t *testing.T) {
	cfg := config.Default()
	cfg.Agents = append(cfg.Agents, config.AgentConfig{ID: "gemini", Adapter: "gemini-cli"})
	if !migrateRetiredGeminiAgent(cfg) {
		t.Fatal("duplicate retired agent should be removed")
	}
	if len(cfg.Agents) != 3 {
		t.Fatalf("agents = %#v", cfg.Agents)
	}
	for _, agent := range cfg.Agents {
		if agent.ID == "gemini" {
			t.Fatalf("retired agent remained: %#v", cfg.Agents)
		}
	}
}

func TestRemoveMarkedBlockFilePreservesUserContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "retired.md")
	content := "before\n\n" + coactBegin + "\nold\n" + coactEnd + "\n\nafter\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	changed, err := removeMarkedBlockFile(path)
	if err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "before\n\nafter\n" {
		t.Fatalf("unexpected preserved content: %q", data)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
