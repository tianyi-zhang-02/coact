package metalock

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAcquireRelease(t *testing.T) {
	path := filepath.Join(t.TempDir(), "board.lock")
	if err := Acquire(path, time.Second, time.Second); err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("lock file missing after acquire: %v", err)
	}
	Release(path)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("lock file should be gone after release, stat err=%v", err)
	}
}

func TestReleaseDoesNotRemoveForeignLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "board.lock")
	if err := Acquire(path, time.Second, time.Second); err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	writeRecord(t, path, record{
		PID:        os.Getpid(),
		AcquiredAt: time.Now().UTC().Format(time.RFC3339),
		Token:      "foreign-token",
	})
	Release(path)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("foreign lock should remain after release, stat err=%v", err)
	}
}

func TestAcquireStealsStaleLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "board.lock")
	writeRecord(t, path, record{
		PID:        12345,
		AcquiredAt: time.Now().Add(-time.Hour).UTC().Format(time.RFC3339),
		Token:      "stale-token",
	})
	if err := Acquire(path, time.Millisecond, time.Second); err != nil {
		t.Fatalf("Acquire should steal stale lock: %v", err)
	}
	Release(path)
}

func TestAcquireDoesNotStealFreshLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "board.lock")
	writeRecord(t, path, record{
		PID:        12345,
		AcquiredAt: time.Now().UTC().Format(time.RFC3339),
		Token:      "fresh-token",
	})
	if err := Acquire(path, time.Hour, 50*time.Millisecond); err != ErrTimeout {
		t.Fatalf("Acquire fresh lock err=%v, want ErrTimeout", err)
	}
}

func writeRecord(t *testing.T, path string, rec record) {
	t.Helper()
	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
