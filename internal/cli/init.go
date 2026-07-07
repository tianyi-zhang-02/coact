package cli

import (
	"fmt"
	"os"

	"github.com/tianyi-zhang-02/coact/internal/project"
	"github.com/tianyi-zhang-02/coact/internal/setup"
)

func cmdInit(args []string) int {
	root, err := project.Locate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}
	res, err := setup.Initialize(root, agentID(""))
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}

	fmt.Printf("coact initialized at %s\n", root)
	if len(res.Wired) > 0 {
		fmt.Println("wired:")
		for _, c := range res.Wired {
			fmt.Printf("  %s\n", c)
		}
	} else {
		fmt.Println("(already initialized — nothing to change)")
	}
	fmt.Print(`
verify it works (no second agent needed):
  coact doctor           checks the wiring and runs an enforcement self-test

then launch each agent in its own terminal — one command each:
  coact claude          # or:  coact codex   /   coact gemini

coact adds a gate — it does not disable your permissions, and the hook fails
open, so if coact ever errors your editing still works. Run "coact adapters" to
see the agents it can coordinate, and "coact deinit" to undo everything.
`)
	return 0
}
