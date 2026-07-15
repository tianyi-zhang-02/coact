package cli

import (
	"fmt"

	"github.com/tianyi-zhang-02/coact/internal/adapter"
)

// cmdAdapters lists the agents CoAct can coordinate (the adapter registry).
func cmdAdapters(args []string) int {
	fmt.Println("agents CoAct can coordinate:")
	for _, a := range adapter.All() {
		fmt.Printf("  %-12s  launch: coact %-12s  contract: %-15s  %s\n",
			a.ID, a.ID, a.ContractFile, a.Enforcement())
	}
	fmt.Println("\nreal-time push (mid-turn) is Claude-only; others coordinate turn-based.")
	return 0
}
