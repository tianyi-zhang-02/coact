package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/expression"
)

func cmdZH(args []string) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		printZHUsage()
		return 0
	}
	switch args[0] {
	case "check":
		return cmdZHCheck(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "coact zh: unknown command %q\n\n", args[0])
		printZHUsage()
		return 1
	}
}

func cmdZHCheck(args []string) int {
	fs := flag.NewFlagSet("zh check", flag.ContinueOnError)
	userMessage := fs.String("user", "", "original user message for trigger detection")
	diagnostics := fs.Bool("diagnostics", false, "print safe diagnostics")
	off := fs.Bool("off", false, "disable the adapter for this check")
	if _, err := parseInterspersed(fs, args); err != nil {
		return 2
	}
	positionals := fs.Args()
	text, err := readZHInput(positionals)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	cfg := expression.DefaultConfig()
	if *off {
		cfg.Enabled = false
	}
	cfg.Diagnostics = *diagnostics
	detect := expression.ShouldPolishChineseOutput(*userMessage, text, cfg)
	protected, spans := expression.ProtectSpans(text, cfg)

	fmt.Printf("chinese-expression: should_run=%v reason=%s\n", detect.ShouldRun, detect.Reason)
	fmt.Printf("ratios: chinese=%.2f code=%.2f\n", detect.ChineseRatio, detect.CodeRatio)
	fmt.Printf("protected_spans: %d\n", len(spans))
	if *diagnostics {
		for _, span := range spans {
			fmt.Printf("  %s %s len=%d\n", span.ID, span.Kind, len([]rune(span.Value)))
		}
		fmt.Println()
		fmt.Println("protected preview:")
		fmt.Println(strings.TrimSpace(protected))
	}
	if detect.ShouldRun {
		fmt.Println()
		fmt.Println("next: pass this text through a caller-supplied polish model, then validate and restore protected spans.")
	}
	return 0
}

func readZHInput(paths []string) (string, error) {
	if len(paths) > 1 {
		return "", fmt.Errorf("usage: coact zh check [flags] [file]")
	}
	if len(paths) == 1 {
		data, err := os.ReadFile(paths[0])
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(data)) == "" {
		return "", fmt.Errorf("usage: echo '中文文本' | coact zh check")
	}
	return string(data), nil
}

func printZHUsage() {
	fmt.Print(`coact zh — Chinese expression adapter diagnostics.

Usage:
  coact zh check [--user "message"] [--diagnostics] [--off] [file]

The adapter is model-agnostic and enabled by default. Use --off to verify the
explicit disable path. This command only reports whether a response should be
polished and which technical spans would be protected; actual polish is supplied
by a caller's response pipeline.
`)
}
