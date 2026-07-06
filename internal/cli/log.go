package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// cmdLog prints recent journal events — the human's oversight view of what the
// agents have done.
func cmdLog(args []string) int {
	fs := flag.NewFlagSet("log", flag.ContinueOnError)
	n := fs.Int("n", 20, "number of recent events to show")
	if _, err := parseInterspersed(fs, args); err != nil {
		return 2
	}
	p, _, ok := loadProject()
	if !ok {
		return 1
	}

	entries, err := os.ReadDir(p.JournalDir())
	if err != nil {
		fmt.Println("(no events yet)")
		return 0
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonl") {
			files = append(files, filepath.Join(p.JournalDir(), e.Name()))
		}
	}
	sort.Strings(files) // date-named files sort chronologically

	var lines []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		for _, ln := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
			if strings.TrimSpace(ln) != "" {
				lines = append(lines, ln)
			}
		}
	}
	if len(lines) == 0 {
		fmt.Println("(no events yet)")
		return 0
	}
	if len(lines) > *n {
		lines = lines[len(lines)-*n:]
	}

	for _, ln := range lines {
		var rec map[string]string
		if json.Unmarshal([]byte(ln), &rec) != nil {
			continue
		}
		var details []string
		for k, v := range rec {
			if k == "ts" || k == "agent" || k == "event" {
				continue
			}
			details = append(details, k+"="+v)
		}
		sort.Strings(details)
		fmt.Printf("%s  %-8s %-14s %s\n", rec["ts"], rec["agent"], rec["event"], strings.Join(details, " "))
	}
	return 0
}
