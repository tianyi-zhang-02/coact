// Command coact governs two coding agents working in one repository.
package main

import (
	"os"

	"github.com/tianyi-zhang-02/coact/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
