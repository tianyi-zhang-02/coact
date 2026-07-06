package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tianyi-zhang-02/coact/internal/board"
	"github.com/tianyi-zhang-02/coact/internal/journal"
	"github.com/tianyi-zhang-02/coact/internal/metalock"
	"github.com/tianyi-zhang-02/coact/internal/presence"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

// withBoardLock serializes board mutations under a dedicated meta-lock so two
// agents never claim the same task at once (SPEC §3.3).
func withBoardLock(p *project.Project, fn func() error) error {
	if err := os.MkdirAll(p.CoactDir(), 0o755); err != nil {
		return err
	}
	lockPath := filepath.Join(p.CoactDir(), "board.lock")
	if err := metalock.Acquire(lockPath, 5*time.Second, 10*time.Second); err != nil {
		return err
	}
	defer metalock.Release(lockPath)
	return fn()
}

func cmdBoard(args []string) int {
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	b, err := board.Load(p.BoardPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: reading board: %v\n", err)
		return 1
	}
	tasks := b.Tasks()
	if len(tasks) == 0 {
		fmt.Println("board is empty — add one with: coact task add \"<title>\"")
		return 0
	}
	for _, t := range tasks {
		owner := t.Owner
		if owner == "" {
			owner = "-"
		}
		fmt.Printf("  %-7s %-9s %-10s %s\n", t.ID, t.State, owner, t.Title)
	}
	return 0
}

func cmdClaim(args []string) int {
	fs := flag.NewFlagSet("claim", flag.ContinueOnError)
	agentFlag := fs.String("agent", "", "participant id")
	ttl := fs.Int("ttl", 1800, "claim ttl in seconds")
	ids, err := parseInterspersed(fs, args)
	if err != nil {
		return 2
	}
	if len(ids) != 1 {
		fmt.Fprintln(os.Stderr, "usage: coact claim [--agent id] [--ttl secs] <task-id>")
		return 2
	}
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	agent := agentID(*agentFlag)

	err = withBoardLock(p, func() error {
		b, err := board.Load(p.BoardPath())
		if err != nil {
			return err
		}
		t, err := b.Claim(ids[0], agent, *ttl)
		if err != nil {
			return err
		}
		if err := b.Save(); err != nil {
			return err
		}
		_ = presence.Beat(p.SessionDir(), agent, "working", t.ID)
		_ = journal.Append(p.JournalDir(), agent, "task.claim", map[string]string{"id": t.ID})
		fmt.Printf("claimed %s (%s) by %s\n", t.ID, t.Title, agent)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}
	return 0
}

func cmdDone(args []string) int {
	fs := flag.NewFlagSet("done", flag.ContinueOnError)
	agentFlag := fs.String("agent", "", "participant id")
	ids, err := parseInterspersed(fs, args)
	if err != nil {
		return 2
	}
	if len(ids) != 1 {
		fmt.Fprintln(os.Stderr, "usage: coact done [--agent id] <task-id>")
		return 2
	}
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	agent := agentID(*agentFlag)

	err = withBoardLock(p, func() error {
		b, err := board.Load(p.BoardPath())
		if err != nil {
			return err
		}
		t, err := b.Finish(ids[0], agent)
		if err != nil {
			return err
		}
		if err := b.Save(); err != nil {
			return err
		}
		_ = journal.Append(p.JournalDir(), agent, "task.finish", map[string]string{"id": t.ID})
		fmt.Printf("done %s (%s)\n", t.ID, t.Title)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}
	return 0
}

func cmdTask(args []string) int {
	if len(args) == 0 || args[0] != "add" {
		fmt.Fprintln(os.Stderr, "usage: coact task add \"<title>\"")
		return 2
	}
	title := strings.TrimSpace(strings.Join(args[1:], " "))
	if title == "" {
		fmt.Fprintln(os.Stderr, "usage: coact task add \"<title>\"")
		return 2
	}
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	err := withBoardLock(p, func() error {
		b, err := board.Load(p.BoardPath())
		if err != nil {
			return err
		}
		t := b.Add(title)
		if err := b.Save(); err != nil {
			return err
		}
		_ = journal.Append(p.JournalDir(), agentID(""), "task.add", map[string]string{"id": t.ID})
		fmt.Printf("added %s: %s\n", t.ID, t.Title)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}
	return 0
}
