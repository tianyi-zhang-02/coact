// Package platform isolates the few OS-specific behaviors coact relies on so
// the rest of the codebase stays portable across macOS, Linux, and Windows.
package platform

import (
	"os"
	"path/filepath"
)

// AtomicWrite writes data to path atomically: it writes to a temp file in the
// same directory, fsyncs, then renames over the target. os.Rename maps to
// MoveFileEx(REPLACE_EXISTING) on Windows, so the replace is atomic on all
// supported platforms.
func AtomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".coact-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// Best-effort cleanup if we fail before the rename.
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
