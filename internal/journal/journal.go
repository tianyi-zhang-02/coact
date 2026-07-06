// Package journal appends protocol-significant events to a per-day JSONL log.
// The journal is coact's audit substrate: it is sufficient to reconstruct who
// held what, who was blocked by whom, and where a conflict originated.
package journal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Append writes one event as a JSON line to today's journal file. Extra fields
// are merged alongside the standard ts/agent/event keys. A nil or empty
// journalDir is a no-op so commands run before `coact init` don't fail.
func Append(journalDir, agent, event string, fields map[string]string) error {
	if journalDir == "" {
		return nil
	}
	if err := os.MkdirAll(journalDir, 0o755); err != nil {
		return err
	}

	rec := map[string]string{
		"ts":    time.Now().UTC().Format(time.RFC3339),
		"agent": agent,
		"event": event,
	}
	for k, v := range fields {
		rec[k] = v
	}

	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	line := append(data, '\n')

	name := time.Now().UTC().Format("2006-01-02") + ".jsonl"
	f, err := os.OpenFile(filepath.Join(journalDir, name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(line)
	return err
}
