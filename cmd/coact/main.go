// Command coact governs two coding agents working in one repository.
package main

import (
	"os"

	"github.com/coactdev/coact/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
