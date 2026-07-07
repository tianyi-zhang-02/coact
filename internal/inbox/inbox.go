// Package inbox is CoAct's turn-based messaging plane: agents send each other
// messages through the shared filesystem (.coact/inbox/<agent>.md), serialized
// under a metalock. It is not real-time push — a recipient reads its inbox at
// the start of a turn — but it removes the human from relaying between agents.
package inbox

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tianyi-zhang-02/coact/internal/metalock"
)

func msgFile(dir, agent string) string  { return filepath.Join(dir, agent+".md") }
func readFile(dir, agent string) string { return filepath.Join(dir, agent+".read.md") }
func lockFile(dir, agent string) string { return filepath.Join(dir, agent+".lock") }

func now() string { return time.Now().UTC().Format(time.RFC3339) }

func lock(dir, agent string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return metalock.Acquire(lockFile(dir, agent), 5*time.Second, 10*time.Second)
}

// Send appends a message from `from` to `to`'s inbox.
func Send(inboxDir, from, to, text string) error {
	if err := lock(inboxDir, to); err != nil {
		return err
	}
	defer metalock.Release(lockFile(inboxDir, to))

	block := fmt.Sprintf("### from %s · %s\n%s\n\n", from, now(), strings.TrimSpace(text))
	f, err := os.OpenFile(msgFile(inboxDir, to), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(block)
	return err
}

// Read returns agent's unread messages. Unless peek, it archives them to
// <agent>.read.md and clears the inbox so they aren't shown again.
func Read(inboxDir, agent string, peek bool) (string, error) {
	if err := lock(inboxDir, agent); err != nil {
		return "", err
	}
	defer metalock.Release(lockFile(inboxDir, agent))

	data, err := os.ReadFile(msgFile(inboxDir, agent))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	content := string(data)
	if peek || strings.TrimSpace(content) == "" {
		return content, nil
	}

	if af, err := os.OpenFile(readFile(inboxDir, agent), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
		af.WriteString(content)
		af.Close()
	}
	if err := os.WriteFile(msgFile(inboxDir, agent), []byte{}, 0o644); err != nil {
		return content, err
	}
	return content, nil
}
