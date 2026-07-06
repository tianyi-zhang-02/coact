// Package buildinfo carries version metadata injected at build time.
package buildinfo

// Version and Commit are overridden via -ldflags in release builds.
var (
	Version = "dev"
	Commit  = "none"
)
