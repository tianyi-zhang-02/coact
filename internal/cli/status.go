package cli

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/tianyi-zhang-02/coact/internal/config"
	"github.com/tianyi-zhang-02/coact/internal/lockmgr"
	"github.com/tianyi-zhang-02/coact/internal/presence"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

func cmdStatus(args []string) int {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	watch := fs.Bool("watch", false, "refresh continuously until Ctrl-C")
	interval := fs.Int("interval", 2, "refresh interval in seconds (with --watch)")
	if _, err := parseInterspersed(fs, args); err != nil {
		return 2
	}
	p, cfg, ok := loadProject()
	if !ok {
		return 1
	}

	if !*watch {
		renderStatus(p, cfg)
		return 0
	}

	iv := *interval
	if iv < 1 {
		iv = 1
	}
	for {
		fmt.Print("\033[2J\033[H") // clear screen, cursor home
		fmt.Printf("coact — %s (refreshing every %ds, Ctrl-C to stop)\n\n",
			time.Now().Format("15:04:05"), iv)
		renderStatus(p, cfg)
		time.Sleep(time.Duration(iv) * time.Second)
	}
}

func renderStatus(p *project.Project, cfg *config.Config) {
	fmt.Printf("coact workspace: %s\n", p.Root)
	fmt.Printf("mode: %s\n\n", cfg.Mode)

	sessions, _ := presence.List(p.SessionDir())
	fmt.Println("participants:")
	if len(sessions) == 0 {
		fmt.Println("  (none — no sessions have checked in)")
	}
	for _, s := range sessions {
		marker := "dead"
		if presence.IsLive(p.SessionDir(), s.Agent, cfg.Presence.TTLSeconds) {
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

	m := lockmgr.New(p, cfg)
	locks, err := m.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: reading locks: %v\n", err)
		return
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
