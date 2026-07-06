package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/policy"
)

func cmdPolicy(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: coact policy <check|show>")
		return 2
	}
	switch args[0] {
	case "check":
		return policyCheck(args[1:])
	case "show":
		return policyShow(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "coact: unknown policy subcommand %q\n", args[0])
		return 2
	}
}

func policyCheck(args []string) int {
	fs := flag.NewFlagSet("policy check", flag.ContinueOnError)
	agentFlag := fs.String("agent", "", "participant id")
	positionals, err := parseInterspersed(fs, args)
	if err != nil {
		return 2
	}
	if len(positionals) != 1 {
		fmt.Fprintln(os.Stderr, "usage: coact policy check [--agent id] <path>")
		return 2
	}
	p, cfg, ok := loadProject()
	if !ok {
		return 1
	}
	rel, err := relToRoot(p.Root, positionals[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}
	agent := agentID(*agentFlag)
	d := policy.New(cfg).Check(agent, rel)
	if d.Allowed {
		fmt.Printf("allow: %s may write %s\n", agent, rel)
		return 0
	}
	fmt.Printf("deny: %s\n", d.Reason)
	return 3
}

func policyShow(args []string) int {
	_, cfg, ok := loadProject()
	if !ok {
		return 1
	}
	fmt.Println("protected paths (no agent may write; human-gated):")
	if len(cfg.Policy.ProtectedPaths) == 0 {
		fmt.Println("  (none)")
	}
	for _, g := range cfg.Policy.ProtectedPaths {
		fmt.Printf("  %s\n", g)
	}
	fmt.Println("\nper-agent write scope (empty = unrestricted):")
	for _, a := range cfg.Agents {
		scope := "unrestricted"
		if len(a.Write) > 0 {
			scope = strings.Join(a.Write, ", ")
		}
		fmt.Printf("  %-10s %s\n", a.ID, scope)
	}
	return 0
}

// relToRoot normalizes a user path to a clean slash path relative to the repo
// root, rejecting anything outside it.
func relToRoot(root, raw string) (string, error) {
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", err
	}
	r, err := filepath.Rel(root, abs)
	if err != nil {
		return "", err
	}
	r = filepath.ToSlash(filepath.Clean(r))
	if r == ".." || strings.HasPrefix(r, "../") {
		return "", fmt.Errorf("path %q is outside the repo root", raw)
	}
	return r, nil
}
