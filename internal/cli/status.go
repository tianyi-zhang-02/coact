package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/tianyi-zhang-02/coact/internal/lockmgr"
	"github.com/tianyi-zhang-02/coact/internal/presence"
)

func cmdStatus(args []string) int {
	p, cfg, ok := loadProject()
	if !ok {
		return 1
	}

	fmt.Printf("coact workspace: %s\n", p.Root)
	fmt.Printf("mode: %s\n\n", cfg.Mode)

	// Participants
	sessions, _ := presence.List(p.SessionDir())
	fmt.Println("participants:")
	if len(sessions) == 0 {
		fmt.Println("  (none — no sessions have checked in)")
	}
	for _, s := range sessions {
		live := presence.IsLive(p.SessionDir(), s.Agent, cfg.Presence.TTLSeconds)
		marker := "dead"
		if live {
			marker = "live"
		}
		age := "?"
		if d, ok := s.Age(); ok {
			age = shortDur(d) + " ago"
		}
		task := s.CurrentTask
		if task == "" {
			task = "-"
		}
		fmt.Printf("  %-10s %-4s  status=%-8s task=%-10s beat=%s\n",
			s.Agent, marker, orDash(s.Status), task, age)
	}

	// Locks
	m := lockmgr.New(p, cfg)
	locks, err := m.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: reading locks: %v\n", err)
		return 1
	}
	fmt.Println("\nactive locks:")
	if len(locks) == 0 {
		fmt.Println("  (none)")
	}
	for _, lk := range locks {
		state := "held"
		if lockmgr.Expired(lk) {
			state = "expired"
		}
		fmt.Printf("  %-30s %-8s owner=%-10s ttl=%ds\n",
			lk.Path, state, lk.Owner, lk.TTLSeconds)
	}
	return 0
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func shortDur(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh", int(d.Hours()))
}
