package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUsageSetWritesSnapshotAndAlertsWorkmates(t *testing.T) {
	dir := chdirInitializedProject(t)
	code := Run([]string{"usage", "set", "--agent", "claude", "--model", "opus", "--percent", "42", "--refresh-in", "7d"})
	if code != 0 {
		t.Fatalf("Run(usage set) = %d", code)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".coact", "usage", "claude.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"threshold_step": 20`) || !strings.Contains(string(data), "40") {
		t.Fatalf("unexpected usage snapshot:\n%s", data)
	}
	for _, recipient := range []string{"codex", "gemini"} {
		message, err := os.ReadFile(filepath.Join(dir, ".coact", "inbox", recipient+".md"))
		if err != nil || !strings.Contains(string(message), "Usage alert") {
			t.Fatalf("%s alert missing: %v\n%s", recipient, err, message)
		}
	}
}

func TestEvalRateAndReportWriteLocalDecisionSupport(t *testing.T) {
	dir := chdirInitializedProject(t)
	if code := Run([]string{"plan", "--id", "review-run", "--with", "codex,claude", "Review work"}); code != 0 {
		t.Fatalf("Run(plan) = %d", code)
	}
	if code := Run([]string{"eval", "rate", "--agent", "codex", "--peer", "claude", "--run", "review-run", "--score", "4", "--code-quality", "5", "--note", "solid review"}); code != 0 {
		t.Fatalf("Run(eval rate) = %d", code)
	}
	if code := Run([]string{"eval", "report", "review-run"}); code != 0 {
		t.Fatalf("Run(eval report) = %d", code)
	}
	report, err := os.ReadFile(filepath.Join(dir, ".coact", "evaluations", "review-run", "report.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Audit events analyzed", "codex rated claude", "code quality 5/5", "subjective"} {
		if !strings.Contains(string(report), want) {
			t.Fatalf("report missing %q:\n%s", want, report)
		}
	}
}
