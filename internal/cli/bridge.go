package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/tianyi-zhang-02/coact/internal/codex"
	"github.com/tianyi-zhang-02/coact/internal/inbox"
	"github.com/tianyi-zhang-02/coact/internal/journal"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

// cmdBridge runs the real-time bridge to another agent's live process. Today
// only Codex is supported (Model B: Codex as a stdio app-server executor).
func cmdBridge(args []string) int {
	if len(args) == 0 || args[0] != "codex" {
		fmt.Fprintln(os.Stderr, "usage: coact bridge codex")
		return 2
	}
	return bridgeCodex(args[1:])
}

// stdioRW adapts a subprocess's stdin/stdout pipes to one io.ReadWriter.
type stdioRW struct {
	in  io.Writer
	out io.Reader
}

func (s stdioRW) Read(p []byte) (int, error)  { return s.out.Read(p) }
func (s stdioRW) Write(p []byte) (int, error) { return s.in.Write(p) }

func bridgeCodex(args []string) int {
	fs := flag.NewFlagSet("bridge codex", flag.ContinueOnError)
	interval := fs.Int("interval", 1, "inbox poll interval in seconds")
	if _, err := parseInterspersed(fs, args); err != nil {
		return 2
	}
	p, _, ok := loadProject()
	if !ok {
		return 1
	}

	// Spawn the Codex app-server (stdio). The exact invocation and framing are
	// confirmed against real codex at integration time; override the command
	// with COACT_CODEX_CMD (space-separated) if needed.
	parts := strings.Fields(os.Getenv("COACT_CODEX_CMD"))
	if len(parts) == 0 {
		parts = []string{"codex", "app-server"}
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact bridge: %v\n", err)
		return 1
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact bridge: %v\n", err)
		return 1
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "coact bridge: could not start %q: %v\n", strings.Join(parts, " "), err)
		return 1
	}

	// Codex's completed messages → Claude's inbox, where the channel pushes them
	// into Claude's session mid-turn.
	onMsg := func(text string) {
		_ = inbox.Send(p.InboxDir(), "codex", "claude", text)
		_ = journal.Append(p.JournalDir(), "codex", "msg.send", map[string]string{"to": "claude", "via": "bridge"})
	}
	client := codex.New(stdioRW{in: stdin, out: stdout}, codex.NewlineCodec{}, onMsg)

	ua, err := client.Initialize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact bridge: codex initialize failed: %v\n", err)
		_ = cmd.Process.Kill()
		return 1
	}
	fmt.Fprintf(os.Stderr, "coact bridge: connected to %s\n", ua)
	if _, err := client.StartThread(p.WorkRoot()); err != nil {
		fmt.Fprintf(os.Stderr, "coact bridge: codex thread/start failed: %v\n", err)
	}

	stop := make(chan struct{})
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() { <-sig; close(stop) }()

	fmt.Fprintln(os.Stderr, "coact bridge codex: relaying (Ctrl-C to stop)")
	runCodexBridge(p, client, time.Duration(*interval)*time.Second, stop)

	_ = client.Close()
	_ = cmd.Process.Kill()
	return 0
}

// runCodexBridge polls codex's inbox and forwards each message into the live
// Codex turn (start or steer).
func runCodexBridge(p *project.Project, client *codex.Client, interval time.Duration, stop <-chan struct{}) {
	if interval <= 0 {
		interval = time.Second
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			content, err := inbox.Read(p.InboxDir(), "codex", false)
			if err == nil && strings.TrimSpace(content) != "" {
				_ = journal.Append(p.JournalDir(), "codex", "msg.read", map[string]string{"via": "bridge"})
				_ = client.Send(bridgeText(content))
			}
		}
	}
}

// bridgeText strips the inbox "### from X · ts" headers, leaving the body.
func bridgeText(content string) string {
	var out []string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "### from ") {
			continue
		}
		out = append(out, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
