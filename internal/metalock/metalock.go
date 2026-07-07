// Package metalock is a small crash-safe mutual-exclusion primitive built on
// atomic O_EXCL file creation. It serializes short critical sections (lock
// registry acquisition, board mutations) across processes. The holder refreshes
// nothing; instead a stale lock older than ttl is stolen, so a crashed holder
// never wedges the system.
package metalock

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// ErrTimeout is returned when the lock cannot be taken within the deadline.
var ErrTimeout = errors.New("metalock: timed out acquiring lock")

type record struct {
	PID        int    `json:"pid"`
	AcquiredAt string `json:"acquired_at"`
	Token      string `json:"token,omitempty"`
}

var held sync.Map // path -> token, scoped to this process

// Acquire creates path exclusively. If it already exists but is older than ttl,
// it is assumed abandoned and stolen. Spins with backoff until timeout.
func Acquire(path string, ttl, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	token := newToken()
	for {
		err := create(path, token)
		if err == nil {
			held.Store(path, token)
			return nil
		}
		if !os.IsExist(err) {
			return err
		}
		if stale(path, ttl) && quarantine(path) {
			continue
		}
		if time.Now().After(deadline) {
			return ErrTimeout
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func create(path, token string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	rec := record{
		PID:        os.Getpid(),
		AcquiredAt: time.Now().UTC().Format(time.RFC3339),
		Token:      token,
	}
	data, err := json.Marshal(rec)
	if err == nil {
		_, err = f.Write(data)
	}
	closeErr := f.Close()
	if err != nil {
		_ = os.Remove(path)
		return err
	}
	if closeErr != nil {
		_ = os.Remove(path)
		return closeErr
	}
	return nil
}

func newToken() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err == nil {
		return hex.EncodeToString(b[:])
	}
	return strconv.FormatInt(time.Now().UnixNano(), 36) + "-" + strconv.Itoa(os.Getpid())
}

func stale(path string, ttl time.Duration) bool {
	if ttl <= 0 {
		return false
	}
	data, err := os.ReadFile(path)
	if err == nil {
		var rec record
		if json.Unmarshal(data, &rec) == nil {
			if t, err := time.Parse(time.RFC3339, rec.AcquiredAt); err == nil {
				return time.Since(t) > ttl
			}
		}
	}
	info, err := os.Stat(path)
	return err == nil && time.Since(info.ModTime()) > ttl
}

func quarantine(path string) bool {
	stalePath := fmt.Sprintf("%s.stale.%d.%d", path, os.Getpid(), time.Now().UnixNano())
	if err := os.Rename(path, stalePath); err != nil {
		return false
	}
	_ = os.Remove(stalePath)
	return true
}

// Release drops the lock only if this process still owns the current file.
// Safe to call unconditionally in a defer.
func Release(path string) {
	raw, ok := held.Load(path)
	if !ok {
		return
	}
	defer held.Delete(path)

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var rec record
	if json.Unmarshal(data, &rec) != nil {
		return
	}
	if rec.Token == raw.(string) {
		_ = os.Remove(path)
	}
}
