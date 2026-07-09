package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/inbox"
	"github.com/tianyi-zhang-02/coact/internal/journal"
	"github.com/tianyi-zhang-02/coact/internal/presence"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

func cmdMention(target string, args []string) int {
	fs := flag.NewFlagSet("@"+target, flag.ContinueOnError)
	agentFlag := fs.String("agent", "", "sender id")
	pos, err := parseInterspersed(fs, args)
	if err != nil || len(pos) == 0 {
		fmt.Fprintf(os.Stderr, "usage: coact @%s [--agent id] <message...>\n", target)
		return 2
	}
	text := strings.TrimSpace(strings.Join(pos, " "))
	if text == "" {
		fmt.Fprintf(os.Stderr, "usage: coact @%s [--agent id] <message...>\n", target)
		return 2
	}

	p, cfg, ok := loadProject()
	if !ok {
		return 1
	}
	from := agentID(*agentFlag)
	recipients := mentionRecipients(target, from, p, cfg.Presence.TTLSeconds)
	if len(recipients) == 0 {
		fmt.Fprintf(os.Stderr, "coact: @%s has no recipients\n", target)
		return 1
	}
	for _, to := range recipients {
		if err := inbox.Send(p.InboxDir(), from, to, text); err != nil {
			fmt.Fprintf(os.Stderr, "coact: %v\n", err)
			return 1
		}
		_ = journal.Append(p.JournalDir(), from, "msg.send", map[string]string{"to": to, "via": "mention"})
	}
	fmt.Printf("sent to %s\n", strings.Join(recipients, ", "))
	return 0
}

func mentionRecipients(target, from string, p *project.Project, ttlSeconds int) []string {
	target = sanitizeAgent(strings.TrimPrefix(target, "@"))
	if target != "all" {
		return []string{target}
	}
	var out []string
	sessions, _ := presence.List(p.SessionDir())
	for _, session := range sessions {
		if session.Agent != from && presence.IsLive(p.SessionDir(), session.Agent, ttlSeconds) {
			out = append(out, session.Agent)
		}
	}
	return out
}
