package cli

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coactdev/coact/internal/journal"
	"github.com/coactdev/coact/internal/presence"
)

// cmdSidecar runs the per-session presence heartbeat until interrupted. An
// agent launcher runs this in the background so the agent counts as live and
// its locks are not stolen mid-work (SPEC §2.9). This is the sidecar; the
// alternative hook-only mode beats on each tool call instead.
func cmdSidecar(args []string) int {
	fs := flag.NewFlagSet("sidecar", flag.ContinueOnError)
	agentFlag := fs.String("agent", "", "participant id")
	interval := fs.Int("interval", 0, "beat interval in seconds (default from config)")
	if _, err := parseInterspersed(fs, args); err != nil {
		return 2
	}
	p, cfg, ok := loadProject()
	if !ok {
		return 1
	}
	agent := agentID(*agentFlag)

	iv := *interval
	if iv <= 0 {
		iv = cfg.Presence.IntervalSeconds
	}
	if iv <= 0 {
		iv = 20
	}

	_ = presence.Register(p.SessionDir(), agent, "working")
	_ = journal.Append(p.JournalDir(), agent, "session.start", map[string]string{"mode": "sidecar"})
	fmt.Printf("coact sidecar: %s beating every %ds (ctrl-c to stop)\n", agent, iv)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	ticker := time.NewTicker(time.Duration(iv) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = presence.Register(p.SessionDir(), agent, "working")
		case <-sig:
			_ = presence.Register(p.SessionDir(), agent, "stopped")
			_ = journal.Append(p.JournalDir(), agent, "session.stop", nil)
			fmt.Println("\ncoact sidecar: stopped")
			return 0
		}
	}
}
