// Package evaluation builds local collaboration reports from audit events and peer ratings.
package evaluation

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tianyi-zhang-02/coact/internal/metalock"
	"github.com/tianyi-zhang-02/coact/internal/platform"
)

type Scores struct {
	Overall        int `json:"overall"`
	Cooperation    int `json:"cooperation"`
	CodeQuality    int `json:"code_quality"`
	Responsiveness int `json:"responsiveness"`
	Alignment      int `json:"alignment"`
}

type Rating struct {
	Run       string `json:"run"`
	Reviewer  string `json:"reviewer"`
	Peer      string `json:"peer"`
	Model     string `json:"model,omitempty"`
	Scores    Scores `json:"scores"`
	Note      string `json:"note,omitempty"`
	CreatedAt string `json:"created_at"`
}

type AgentMetrics struct {
	Agent             string
	MessagesSent      int
	MessagesRead      int
	TasksClaimed      int
	TasksFinished     int
	LockDenials       int
	LockViolations    int
	MergeConflicts    int
	Handoffs          int
	ResponseSamples   int
	MedianResponse    time.Duration
	AveragePeerScores map[string]float64
}

type Report struct {
	Run       string
	Generated time.Time
	Agents    []AgentMetrics
	Ratings   []Rating
	Events    int
}

func SaveRating(dir string, rating Rating, now time.Time) error {
	if err := validateID(rating.Run); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	if err := validateID(rating.Reviewer); err != nil {
		return fmt.Errorf("reviewer: %w", err)
	}
	if err := validateID(rating.Peer); err != nil {
		return fmt.Errorf("peer: %w", err)
	}
	if rating.Reviewer == rating.Peer {
		return errors.New("reviewer and peer must be different")
	}
	if len(rating.Note) > 1000 {
		return errors.New("note must be 1000 characters or fewer")
	}
	if len(rating.Model) > 120 {
		return errors.New("model must be 120 characters or fewer")
	}
	for name, score := range scoreMap(rating.Scores) {
		if score < 1 || score > 5 {
			return fmt.Errorf("%s score must be between 1 and 5", name)
		}
	}
	runDir := filepath.Join(dir, rating.Run)
	if err := os.MkdirAll(runDir, 0o700); err != nil {
		return err
	}
	lockPath := filepath.Join(runDir, rating.Reviewer+"-to-"+rating.Peer+".lock")
	if err := metalock.Acquire(lockPath, 5*time.Second, 10*time.Second); err != nil {
		return err
	}
	defer metalock.Release(lockPath)
	rating.CreatedAt = now.UTC().Format(time.RFC3339)
	data, err := json.MarshalIndent(rating, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return platform.AtomicWrite(filepath.Join(runDir, rating.Reviewer+"-to-"+rating.Peer+".json"), data, 0o600)
}

func LoadRatings(dir, run string) ([]Rating, error) {
	if err := validateID(run); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(filepath.Join(dir, run))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var ratings []Rating
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, run, entry.Name()))
		if err != nil {
			continue
		}
		var rating Rating
		if json.Unmarshal(data, &rating) == nil {
			ratings = append(ratings, rating)
		}
	}
	sort.Slice(ratings, func(i, j int) bool {
		if ratings[i].Peer == ratings[j].Peer {
			return ratings[i].Reviewer < ratings[j].Reviewer
		}
		return ratings[i].Peer < ratings[j].Peer
	})
	return ratings, nil
}

func BuildReport(run string, records []map[string]string, ratings []Rating, now time.Time) Report {
	start, end := runWindow(records, run)
	var filtered []map[string]string
	for _, record := range records {
		ts, err := time.Parse(time.RFC3339, record["ts"])
		if err != nil || (run != "workspace" && start.IsZero()) || (!start.IsZero() && ts.Before(start)) || (!end.IsZero() && !ts.Before(end)) {
			continue
		}
		filtered = append(filtered, record)
	}
	metrics := map[string]*AgentMetrics{}
	metric := func(agent string) *AgentMetrics {
		if agent == "" {
			agent = "unknown"
		}
		if metrics[agent] == nil {
			metrics[agent] = &AgentMetrics{Agent: agent, AveragePeerScores: map[string]float64{}}
		}
		return metrics[agent]
	}
	for _, record := range filtered {
		m := metric(record["agent"])
		switch record["event"] {
		case "msg.send":
			m.MessagesSent++
		case "msg.read":
			m.MessagesRead++
		case "task.claim":
			m.TasksClaimed++
		case "task.finish":
			m.TasksFinished++
		case "lock.denied":
			m.LockDenials++
		case "lock.violation":
			m.LockViolations++
		case "merge.conflict":
			m.MergeConflicts++
		case "handoff":
			m.Handoffs++
		}
	}
	responseDurations := observedResponses(filtered)
	for agent, durations := range responseDurations {
		m := metric(agent)
		m.ResponseSamples = len(durations)
		m.MedianResponse = median(durations)
	}
	type totals struct{ overall, cooperation, codeQuality, responsiveness, alignment, count int }
	peerTotals := map[string]*totals{}
	for _, rating := range ratings {
		metric(rating.Reviewer)
		metric(rating.Peer)
		if peerTotals[rating.Peer] == nil {
			peerTotals[rating.Peer] = &totals{}
		}
		t := peerTotals[rating.Peer]
		t.overall += rating.Scores.Overall
		t.cooperation += rating.Scores.Cooperation
		t.codeQuality += rating.Scores.CodeQuality
		t.responsiveness += rating.Scores.Responsiveness
		t.alignment += rating.Scores.Alignment
		t.count++
	}
	for agent, total := range peerTotals {
		if total.count == 0 {
			continue
		}
		m := metric(agent)
		div := float64(total.count)
		m.AveragePeerScores = map[string]float64{
			"overall":        float64(total.overall) / div,
			"cooperation":    float64(total.cooperation) / div,
			"code_quality":   float64(total.codeQuality) / div,
			"responsiveness": float64(total.responsiveness) / div,
			"alignment":      float64(total.alignment) / div,
		}
	}
	var agents []AgentMetrics
	for _, m := range metrics {
		agents = append(agents, *m)
	}
	sort.Slice(agents, func(i, j int) bool { return agents[i].Agent < agents[j].Agent })
	return Report{Run: run, Generated: now.UTC(), Agents: agents, Ratings: ratings, Events: len(filtered)}
}

func Markdown(report Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# CoAct collaboration report: %s\n\n", report.Run)
	fmt.Fprintf(&b, "Generated: %s\n\n", report.Generated.Format(time.RFC3339))
	fmt.Fprintf(&b, "Audit events analyzed: %d\n\n", report.Events)
	b.WriteString("## Agent summary\n\n")
	if len(report.Agents) == 0 {
		b.WriteString("No auditable activity or peer ratings yet.\n")
	} else {
		b.WriteString("| Agent | Tasks done | Messages | Lock issues | Observed response | Peer overall | Code quality |\n")
		b.WriteString("|---|---:|---:|---:|---:|---:|---:|\n")
		for _, agent := range report.Agents {
			response := "n/a"
			if agent.ResponseSamples > 0 {
				response = shortDuration(agent.MedianResponse) + fmt.Sprintf(" (%d)", agent.ResponseSamples)
			}
			fmt.Fprintf(&b, "| %s | %d/%d | %d sent, %d read | %d denied, %d violation, %d merge | %s | %s | %s |\n",
				agent.Agent, agent.TasksFinished, agent.TasksClaimed, agent.MessagesSent, agent.MessagesRead,
				agent.LockDenials, agent.LockViolations, agent.MergeConflicts, response,
				score(agent.AveragePeerScores["overall"]), score(agent.AveragePeerScores["code_quality"]))
		}
	}
	b.WriteString("\n## Peer ratings\n\n")
	if len(report.Ratings) == 0 {
		b.WriteString("No peer ratings submitted. Code quality and cooperation cannot be inferred safely from audit events alone.\n")
	} else {
		for _, rating := range report.Ratings {
			fmt.Fprintf(&b, "- %s rated %s", rating.Reviewer, rating.Peer)
			if rating.Model != "" {
				fmt.Fprintf(&b, " (%s)", rating.Model)
			}
			fmt.Fprintf(&b, ": overall %d/5, cooperation %d/5, code quality %d/5, responsiveness %d/5, alignment %d/5",
				rating.Scores.Overall, rating.Scores.Cooperation, rating.Scores.CodeQuality, rating.Scores.Responsiveness, rating.Scores.Alignment)
			if rating.Note != "" {
				fmt.Fprintf(&b, " — %s", strings.ReplaceAll(rating.Note, "\n", " "))
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("\n## Interpretation\n\n")
	b.WriteString("Observed response time is the delay from a journaled message to the recipient's next journaled action. It is a coordination signal, not provider latency. Peer scores are subjective human/agent input; audit counts are factual local events.\n")
	return b.String()
}

func validateID(value string) error {
	if value == "" {
		return errors.New("value is required")
	}
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' && r != '-' {
			return fmt.Errorf("unsafe id %q", value)
		}
	}
	return nil
}

func scoreMap(scores Scores) map[string]int {
	return map[string]int{
		"overall": scores.Overall, "cooperation": scores.Cooperation,
		"code-quality": scores.CodeQuality, "responsiveness": scores.Responsiveness,
		"alignment": scores.Alignment,
	}
}

func runWindow(records []map[string]string, run string) (time.Time, time.Time) {
	var start, end time.Time
	for _, record := range records {
		if record["event"] != "plan.start" {
			continue
		}
		timestamp, err := time.Parse(time.RFC3339, record["ts"])
		if err != nil {
			continue
		}
		if record["id"] == run {
			start = timestamp
			continue
		}
		if !start.IsZero() && timestamp.After(start) {
			end = timestamp
			break
		}
	}
	return start, end
}

func observedResponses(records []map[string]string) map[string][]time.Duration {
	out := map[string][]time.Duration{}
	for i, record := range records {
		if record["event"] != "msg.send" || record["to"] == "" {
			continue
		}
		sent, err := time.Parse(time.RFC3339, record["ts"])
		if err != nil {
			continue
		}
		for _, candidate := range records[i+1:] {
			if candidate["agent"] != record["to"] {
				continue
			}
			responded, err := time.Parse(time.RFC3339, candidate["ts"])
			if err == nil && !responded.Before(sent) && responded.Sub(sent) <= 24*time.Hour {
				out[record["to"]] = append(out[record["to"]], responded.Sub(sent))
			}
			break
		}
	}
	return out
}

func median(values []time.Duration) time.Duration {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]time.Duration(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	return sorted[len(sorted)/2]
}

func shortDuration(value time.Duration) string {
	if value < time.Minute {
		return fmt.Sprintf("%ds", int(value.Seconds()))
	}
	if value < time.Hour {
		return fmt.Sprintf("%dm", int(value.Minutes()))
	}
	return fmt.Sprintf("%.1fh", value.Hours())
}

func score(value float64) string {
	if value == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%.1f/5", value)
}
