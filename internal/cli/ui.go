package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/tianyi-zhang-02/coact/internal/ui"
)

func cmdUI(args []string) int {
	fs := flag.NewFlagSet("ui", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	addr := fs.String("addr", "127.0.0.1", "local address to bind")
	port := fs.Int("port", 7331, "local port to bind")
	noOpen := fs.Bool("no-open", false, "print the URL without opening a browser")
	lang := fs.String("lang", "en", "UI language hint: en or zh")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "coact ui: unexpected argument %q\n", fs.Arg(0))
		return 2
	}
	if err := ui.Serve(ui.Options{
		Addr:        *addr,
		Port:        *port,
		Lang:        *lang,
		OpenBrowser: !*noOpen,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "coact ui: %v\n", err)
		return 1
	}
	return 0
}
