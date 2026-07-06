// Package metalock is a small crash-safe mutual-exclusion primitive built on
// atomic O_EXCL file creation. It serializes short critical sections (lock
// registry acquisition, board mutations) across processes. The holder refreshes
// nothing; instead a stale lock older than ttl is stolen, so a crashed holder
// never wedges the system.
package metalock

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// ErrTimeout is returned when the lock cannot be taken within the deadline.
var ErrTimeout = errors.New("metalock: timed out acquiring lock")

// Acquire creates path exclusively. If it already exists but is older than ttl,
// it is assumed abandoned and stolen. Spins with backoff until timeout.
func Acquire(path string, ttl, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			fmt.Fprintf(f, `{"pid":%d,"acquired_at":%q}`, os.Getpid(),
				time.Now().UTC().Format(time.RFC3339))
			f.Close()
			return nil
		}
		if !os.IsExist(err) {
			return err
		}
		if info, statErr := os.Stat(path); statErr == nil {
			if time.Since(info.ModTime()) > ttl {
				os.Remove(path) // steal an abandoned lock
				continue
			}
		}
		if time.Now().After(deadline) {
			return ErrTimeout
		}
		time.Sleep(25 * time.Millisecond)
	}
}

// Release drops the lock. Safe to call unconditionally in a defer.
func Release(path string) { os.Remove(path) }
