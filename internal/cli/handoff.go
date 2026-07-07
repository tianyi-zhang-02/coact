package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/board"
	"github.com/tianyi-zhang-02/coact/internal/inbox"
	"github.com/tianyi-zhang-02/coact/internal/journal"
	"github.com/tianyi-zhang-02/coact/internal/lockmgr"
)

// cmdHandoff hands the current agent's work to another agent: it reassigns the
// agent's active board tasks, releases its locks, and messages the recipient
// with context. This is the explicit "I'm stopping / hitting my plan limit,
// take over" hand-off.
func cmdHandoff(args []string) int {
	fs := flag.NewFlagSet("handoff", flag.ContinueOnError)
	agentFlag := fs.String("agent", "", "your id")
	pos, err := parseInterspersed(fs, args)
	if err != nil || len(pos) < 1 {
		fmt.Fprintln(os.Stderr, "usage: coact handoff [--agent id] <to> [note...]")
		return 2
	}
	to := sanitizeAgent(pos[0])
	note := strings.Join(pos[1:], " ")

	p, cfg, ok := loadProject()
	if !ok {
		return 1
	}
	from := agentID(*agentFlag)

	var moved []string
	err = withBoardLock(p, func() error {
		b, err := board.Load(p.BoardPath())
		if err != nil {
			return err
		}
		moved = b.Reassign(from, to)
		return b.Save()
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}

	// Release the handing-off agent's locks so the recipient can take over.
	m := lockmgr.New(p, cfg)
	_, _ = m.ReleaseAll(from)

	msg := fmt.Sprintf("Handoff from %s.", from)
	if len(moved) > 0 {
		msg += " Tasks now yours: " + strings.Join(moved, ", ") + "."
	}
	if note != "" {
		msg += " Note: " + note
	}
	_ = inbox.Send(p.InboxDir(), from, to, msg)
	_ = journal.Append(p.JournalDir(), from, "handoff", map[string]string{"to": to, "tasks": strings.Join(moved, ",")})

	fmt.Printf("handed off to %s", to)
	if len(moved) > 0 {
		fmt.Printf(" — %d task(s): %s", len(moved), strings.Join(moved, ", "))
	}
	fmt.Println()
	return 0
}
