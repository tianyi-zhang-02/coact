//go:build windows

package platform

import "os"

// ProcessAlive reports whether a process with the given pid is currently
// running. On Windows, os.FindProcess opens a handle via OpenProcess and
// returns an error when the process does not exist, which is adequate for
// coact's liveness heuristic.
func ProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	_ = p.Release()
	return true
}
