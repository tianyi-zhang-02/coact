package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/board"
	"github.com/tianyi-zhang-02/coact/internal/presence"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

func cmdHome() int {
	p, err := project.Resolve()
	if err != nil {
		fmt.Println("coact — terminal-native coordination for coding agents")
		fmt.Println()
		fmt.Println("This directory is not initialized yet.")
		fmt.Println("Run: coact init")
		fmt.Println()
		fmt.Println("Then use:")
		fmt.Println("  coact plan \"describe the work\"")
		fmt.Println("  coact @claude \"message\"")
		fmt.Println("  coact @codex \"message\"")
		fmt.Println("  coact board")
		return 0
	}

	fmt.Printf("coact workspace: %s\n", p.Root)
	fmt.Println("mode: terminal-native coordination")
	fmt.Println()
	fmt.Println("shared context:")
	fmt.Printf("  team:   %s\n", homeRel(p.Root, p.TeamPath()))
	fmt.Printf("  memory: %s\n", homeRel(p.Root, p.MemoryDir()))
	fmt.Printf("  runs:   %s\n", homeRel(p.Root, p.RunsDir()))
	fmt.Println()

	if b, err := board.Load(p.BoardPath()); err == nil {
		tasks := b.Tasks()
		fmt.Printf("board: %d task(s)", len(tasks))
		if len(tasks) > 0 {
			var active []string
			for _, t := range tasks {
				if t.State != "done" {
					owner := t.Owner
					if owner == "" {
						owner = "unassigned"
					}
					active = append(active, fmt.Sprintf("%s/%s/%s", t.ID, t.State, owner))
				}
			}
			if len(active) > 0 {
				fmt.Printf(" — %s", strings.Join(active, ", "))
			}
		}
		fmt.Println()
	}

	sessions, _ := presence.List(p.SessionDir())
	fmt.Printf("agents: %d known session(s)\n", len(sessions))
	fmt.Println()
	fmt.Println("next commands:")
	fmt.Println("  coact inbox")
	fmt.Println("  coact @all \"message\"")
	fmt.Println("  coact plan --with codex,claude --distributor codex \"brief\"")
	fmt.Println("  coact board")
	fmt.Println("  coact status")
	return 0
}

func homeRel(root, path string) string {
	if r, err := filepath.Rel(root, path); err == nil {
		return r
	}
	return path
}
