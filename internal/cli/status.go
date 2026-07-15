package cli

import (
	"flag"
	"fmt"
	"os"
	"sort"
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
	all := fs.Bool("all", false, "include historical sessions not in the current agent configuration")
	if _, err := parseInterspersed(fs, args); err != nil {
		return 2
	}
	p, cfg, ok := loadProject()
	if !ok {
		return 1
	}

	if !*watch {
		renderStatus(p, cfg, *all)
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
		renderStatus(p, cfg, *all)
		time.Sleep(time.Duration(iv) * time.Second)
	}
}

func renderStatus(p *project.Project, cfg *config.Config, includeHistory bool) {
	fmt.Printf("coact workspace: %s\n", p.Root)
	fmt.Printf("mode: %s\n\n", cfg.Mode)

	sessions, _ := presence.List(p.SessionDir())
	sessionByAgent := make(map[string]*presence.Session, len(sessions))
	for _, session := range sessions {
		sessionByAgent[session.Agent] = session
	}
	ordered := make([]*presence.Session, 0, len(cfg.Agents)+len(sessions))
	configured := make(map[string]bool, len(cfg.Agents))
	for _, agent := range cfg.Agents {
		configured[agent.ID] = true
		if session := sessionByAgent[agent.ID]; session != nil {
			ordered = append(ordered, session)
		} else {
			ordered = append(ordered, &presence.Session{Agent: agent.ID, Status: "offline"})
		}
	}
	if includeHistory {
		var historical []*presence.Session
		for _, session := range sessions {
			if !configured[session.Agent] {
				historical = append(historical, session)
			}
		}
		sort.Slice(historical, func(i, j int) bool { return historical[i].Agent < historical[j].Agent })
		ordered = append(ordered, historical...)
	}
	fmt.Println("participants:")
	if len(ordered) == 0 {
		fmt.Println("  (none — no sessions have checked in)")
	}
	for _, s := range ordered {
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
	fmt.Println("\nlocks:")
	if len(locks) == 0 {
		fmt.Println("  (none)")
	}
	for _, lk := range locks {
		state := "held"
		if lockmgr.Expired(lk) {
			state = "expired"
			if presence.IsLive(p.SessionDir(), lk.Owner, cfg.Presence.TTLSeconds) {
				state = "expired/live-owner"
			}
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
