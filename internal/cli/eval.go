package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tianyi-zhang-02/coact/internal/evaluation"
	"github.com/tianyi-zhang-02/coact/internal/journal"
	"github.com/tianyi-zhang-02/coact/internal/platform"
)

func cmdEval(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: coact eval [rate|report]")
		return 2
	}
	switch args[0] {
	case "rate":
		return cmdEvalRate(args[1:])
	case "report":
		return cmdEvalReport(args[1:])
	default:
		fmt.Fprintln(os.Stderr, "usage: coact eval [rate|report]")
		return 2
	}
}

func cmdEvalRate(args []string) int {
	fs := flag.NewFlagSet("eval rate", flag.ContinueOnError)
	reviewerFlag := fs.String("agent", "", "reviewer id (default: current participant)")
	peerFlag := fs.String("peer", "", "agent being rated")
	runFlag := fs.String("run", "", "planning run id (default: latest)")
	model := fs.String("model", "", "peer model label")
	overall := fs.Int("score", 0, "overall score, 1-5")
	cooperation := fs.Int("cooperation", 0, "cooperation score, 1-5 (default: --score)")
	codeQuality := fs.Int("code-quality", 0, "code quality score, 1-5 (default: --score)")
	responsiveness := fs.Int("responsiveness", 0, "responsiveness score, 1-5 (default: --score)")
	alignment := fs.Int("alignment", 0, "discrepancy/alignment handling score, 1-5 (default: --score)")
	note := fs.String("note", "", "short evidence note (max 1000 characters)")
	if _, err := parseInterspersed(fs, args); err != nil {
		return 2
	}
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	if strings.TrimSpace(*peerFlag) == "" || *overall == 0 {
		fmt.Fprintln(os.Stderr, "usage: coact eval rate --peer <agent> --score <1-5> [--run id] [dimension scores]")
		return 2
	}
	run := safeRunID(*runFlag)
	if run == "" {
		run = latestRunID(p)
	}
	if run == "" {
		run = "workspace"
	}
	fillScore := func(value int) int {
		if value == 0 {
			return *overall
		}
		return value
	}
	reviewer := agentID(*reviewerFlag)
	peer := sanitizeAgent(*peerFlag)
	rating := evaluation.Rating{
		Run: run, Reviewer: reviewer, Peer: peer, Model: strings.TrimSpace(*model), Note: strings.TrimSpace(*note),
		Scores: evaluation.Scores{
			Overall: *overall, Cooperation: fillScore(*cooperation), CodeQuality: fillScore(*codeQuality),
			Responsiveness: fillScore(*responsiveness), Alignment: fillScore(*alignment),
		},
	}
	if err := evaluation.SaveRating(p.EvaluationDir(), rating, time.Now()); err != nil {
		fmt.Fprintf(os.Stderr, "coact eval: %v\n", err)
		return 1
	}
	_ = journal.Append(p.JournalDir(), reviewer, "eval.rate", map[string]string{"run": run, "peer": peer, "score": fmt.Sprint(*overall)})
	fmt.Printf("saved %s -> %s rating for %s; report: coact eval report %s\n", reviewer, peer, run, run)
	return 0
}

func cmdEvalReport(args []string) int {
	fs := flag.NewFlagSet("eval report", flag.ContinueOnError)
	watch := fs.Bool("watch", false, "refresh continuously until Ctrl-C")
	interval := fs.Int("interval", 5, "refresh interval in seconds")
	positionals, err := parseInterspersed(fs, args)
	if err != nil || len(positionals) > 1 {
		return 2
	}
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	run := ""
	if len(positionals) == 1 {
		run = safeRunID(positionals[0])
	}
	if run == "" {
		run = latestRunID(p)
	}
	if run == "" {
		run = "workspace"
	}
	render := func(persist bool) int {
		records, err := journal.ReadRecent(p.JournalDir(), 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "coact eval: %v\n", err)
			return 1
		}
		ratings, err := evaluation.LoadRatings(p.EvaluationDir(), run)
		if err != nil {
			fmt.Fprintf(os.Stderr, "coact eval: %v\n", err)
			return 1
		}
		report := evaluation.BuildReport(run, records, ratings, time.Now())
		markdown := evaluation.Markdown(report)
		fmt.Print(markdown)
		if persist {
			path := filepath.Join(p.EvaluationDir(), run, "report.md")
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				fmt.Fprintf(os.Stderr, "coact eval: %v\n", err)
				return 1
			}
			if err := platform.AtomicWrite(path, []byte(markdown), 0o600); err != nil {
				fmt.Fprintf(os.Stderr, "coact eval: %v\n", err)
				return 1
			}
			_ = journal.Append(p.JournalDir(), agentID(""), "eval.report", map[string]string{"run": run, "path": homeRel(p.Root, path)})
			fmt.Printf("\nsaved: %s\n", homeRel(p.Root, path))
		}
		return 0
	}
	if !*watch {
		return render(true)
	}
	if *interval < 1 {
		*interval = 1
	}
	for {
		fmt.Print("\033[2J\033[H")
		if code := render(false); code != 0 {
			return code
		}
		time.Sleep(time.Duration(*interval) * time.Second)
	}
}
