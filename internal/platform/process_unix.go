//go:build !windows

package platform

import "syscall"

// ProcessAlive reports whether a process with the given pid is currently
// running. On Unix it sends signal 0, which performs permission/existence
// checking without delivering a signal.
func ProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	if err == nil {
		return true
	}
	// EPERM means the process exists but is owned by another user.
	return err == syscall.EPERM
}
