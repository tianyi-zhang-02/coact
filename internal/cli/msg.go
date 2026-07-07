package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/inbox"
	"github.com/tianyi-zhang-02/coact/internal/journal"
)

func cmdMsg(args []string) int {
	fs := flag.NewFlagSet("msg", flag.ContinueOnError)
	agentFlag := fs.String("agent", "", "sender id")
	pos, err := parseInterspersed(fs, args)
	if err != nil || len(pos) < 2 {
		fmt.Fprintln(os.Stderr, `usage: coact msg [--agent id] <to> <message...>`)
		return 2
	}
	to := sanitizeAgent(pos[0])
	text := strings.Join(pos[1:], " ")

	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	from := agentID(*agentFlag)
	if err := inbox.Send(p.InboxDir(), from, to, text); err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}
	_ = journal.Append(p.JournalDir(), from, "msg.send", map[string]string{"to": to})
	fmt.Printf("sent to %s\n", to)
	return 0
}

func cmdInbox(args []string) int {
	fs := flag.NewFlagSet("inbox", flag.ContinueOnError)
	agentFlag := fs.String("agent", "", "your id")
	peek := fs.Bool("peek", false, "show without marking read")
	if _, err := parseInterspersed(fs, args); err != nil {
		return 2
	}
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	me := agentID(*agentFlag)
	content, err := inbox.Read(p.InboxDir(), me, *peek)
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}
	if strings.TrimSpace(content) == "" {
		fmt.Printf("no new messages for %s\n", me)
		return 0
	}
	fmt.Print(content)
	if !*peek {
		_ = journal.Append(p.JournalDir(), me, "msg.read", nil)
	}
	return 0
}
