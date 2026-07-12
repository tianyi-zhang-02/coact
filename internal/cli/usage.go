package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tianyi-zhang-02/coact/internal/config"
	"github.com/tianyi-zhang-02/coact/internal/inbox"
	"github.com/tianyi-zhang-02/coact/internal/journal"
	"github.com/tianyi-zhang-02/coact/internal/usage"
)

func cmdUsage(args []string) int {
	if len(args) == 0 {
		return cmdUsageReport(nil)
	}
	switch args[0] {
	case "set":
		return cmdUsageSet(args[1:])
	case "report":
		return cmdUsageReport(args[1:])
	case "alerts":
		return cmdUsageAlerts(args[1:])
	default:
		fmt.Fprintln(os.Stderr, "usage: coact usage [set|report|alerts]")
		return 2
	}
}

func cmdUsageSet(args []string) int {
	fs := flag.NewFlagSet("usage set", flag.ContinueOnError)
	agentFlag := fs.String("agent", "", "agent/provider id (default: current participant)")
	model := fs.String("model", "", "model label")
	period := fs.String("period", "weekly", "quota period label")
	percent := fs.Float64("percent", -1, "used percentage, 0-100")
	used := fs.Float64("used", -1, "used quota units")
	limit := fs.Float64("limit", -1, "quota limit units")
	refresh := fs.String("refresh", "", "next refresh time (RFC3339)")
	refreshIn := fs.String("refresh-in", "", "time until refresh, e.g. 6h or 7d")
	step := fs.Int("step", usage.DefaultThresholdStep, "alert threshold step, 1-100")
	if _, err := parseInterspersed(fs, args); err != nil {
		return 2
	}
	p, cfg, ok := loadProject()
	if !ok {
		return 1
	}
	now := time.Now().UTC()
	refreshAt, err := parseRefresh(*refresh, *refreshIn, now)
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact usage: %v\n", err)
		return 2
	}
	usedValue, limitValue, err := usageValues(*percent, *used, *limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact usage: %v\n", err)
		return 2
	}
	agent := agentID(*agentFlag)
	snapshot, alerts, err := usage.Set(p.UsageDir(), usage.Snapshot{
		Agent: agent, Model: strings.TrimSpace(*model), Period: strings.TrimSpace(*period),
		Used: usedValue, Limit: limitValue, RefreshAt: refreshAt.Format(time.RFC3339), ThresholdStep: *step,
	}, now)
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact usage: %v\n", err)
		return 1
	}
	_ = journal.Append(p.JournalDir(), agentID(""), "usage.update", map[string]string{
		"provider": agent, "percent": fmt.Sprintf("%.1f", snapshot.Percent()), "refresh_at": snapshot.RefreshAt,
	})
	if len(alerts) > 0 {
		highest := alerts[len(alerts)-1]
		message := fmt.Sprintf("Usage alert: %s%s reached %.1f%% (threshold %d%%). Refresh: %s. Human can review with `coact usage report` and `coact usage alerts`.",
			agent, modelSuffix(snapshot.Model), snapshot.Percent(), highest.Threshold, snapshot.RefreshAt)
		for _, recipient := range usageRecipients(cfg, agentID("")) {
			if err := inbox.Send(p.InboxDir(), "coact", recipient, message); err != nil {
				fmt.Fprintf(os.Stderr, "coact usage: warning: alert delivery to %s failed: %v\n", recipient, err)
			}
		}
		for _, alert := range alerts {
			_ = journal.Append(p.JournalDir(), "coact", "usage.threshold", map[string]string{
				"provider": alert.Agent, "threshold": strconv.Itoa(alert.Threshold), "percent": fmt.Sprintf("%.1f", alert.Percent),
			})
		}
		fmt.Printf("alerted at %d%% threshold; review: coact usage report\n", highest.Threshold)
	}
	fmt.Printf("%s%s: %.1f%% used; refresh %s\n", snapshot.Agent, modelSuffix(snapshot.Model), snapshot.Percent(), snapshot.RefreshAt)
	return 0
}

func cmdUsageReport(args []string) int {
	fs := flag.NewFlagSet("usage report", flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "print machine-readable JSON")
	watch := fs.Bool("watch", false, "refresh continuously until Ctrl-C")
	interval := fs.Int("interval", 5, "refresh interval in seconds")
	if _, err := parseInterspersed(fs, args); err != nil {
		return 2
	}
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	render := func() int {
		snapshots, err := usage.List(p.UsageDir())
		if err != nil {
			fmt.Fprintf(os.Stderr, "coact usage: %v\n", err)
			return 1
		}
		if *jsonOutput {
			data, _ := json.MarshalIndent(snapshots, "", "  ")
			fmt.Println(string(data))
			return 0
		}
		if len(snapshots) == 0 {
			fmt.Println("no usage snapshots — add one with: coact usage set --agent claude --percent 20 --refresh-in 7d")
			return 0
		}
		now := time.Now().UTC()
		fmt.Println("provider/model                 used      refresh")
		for _, snapshot := range snapshots {
			state := snapshot.RefreshAt
			if snapshot.RefreshDue(now) {
				state = "DUE — update with `coact usage set`"
			}
			fmt.Printf("%-30s %6.1f%%   %s\n", snapshot.Agent+modelSuffix(snapshot.Model), snapshot.Percent(), state)
		}
		return 0
	}
	if !*watch {
		return render()
	}
	if *interval < 1 {
		*interval = 1
	}
	for {
		fmt.Print("\033[2J\033[H")
		if code := render(); code != 0 {
			return code
		}
		time.Sleep(time.Duration(*interval) * time.Second)
	}
}

func cmdUsageAlerts(args []string) int {
	fs := flag.NewFlagSet("usage alerts", flag.ContinueOnError)
	n := fs.Int("n", 20, "number of recent alerts")
	if _, err := parseInterspersed(fs, args); err != nil {
		return 2
	}
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	alerts, err := usage.ReadAlerts(p.UsageDir(), *n)
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact usage: %v\n", err)
		return 1
	}
	if len(alerts) == 0 {
		fmt.Println("no usage threshold alerts yet")
		return 0
	}
	for _, alert := range alerts {
		fmt.Printf("%s  %-10s %5.1f%% crossed %d%%  refresh=%s\n", alert.At, alert.Agent, alert.Percent, alert.Threshold, alert.RefreshAt)
	}
	return 0
}

func parseRefresh(absolute, relative string, now time.Time) (time.Time, error) {
	if absolute != "" && relative != "" {
		return time.Time{}, fmt.Errorf("use only one of --refresh or --refresh-in")
	}
	if absolute != "" {
		parsed, err := time.Parse(time.RFC3339, absolute)
		if err != nil {
			return time.Time{}, fmt.Errorf("--refresh must be RFC3339: %w", err)
		}
		return parsed.UTC(), nil
	}
	if relative == "" {
		return time.Time{}, fmt.Errorf("--refresh or --refresh-in is required")
	}
	duration, err := parseFriendlyDuration(relative)
	if err != nil || duration <= 0 {
		return time.Time{}, fmt.Errorf("invalid --refresh-in %q (try 6h or 7d)", relative)
	}
	return now.Add(duration), nil
}

func parseFriendlyDuration(value string) (time.Duration, error) {
	if strings.HasSuffix(value, "d") {
		days, err := strconv.ParseFloat(strings.TrimSuffix(value, "d"), 64)
		return time.Duration(days * float64(24*time.Hour)), err
	}
	return time.ParseDuration(value)
}

func usageValues(percent, used, limit float64) (float64, float64, error) {
	if percent >= 0 {
		if used >= 0 || limit >= 0 || percent > 100 {
			return 0, 0, fmt.Errorf("use --percent alone (0-100), or use --used and --limit together")
		}
		return percent, 100, nil
	}
	if used < 0 || limit <= 0 {
		return 0, 0, fmt.Errorf("provide --percent, or both --used and --limit")
	}
	return used, limit, nil
}

func usageRecipients(cfg *config.Config, sender string) []string {
	seen := map[string]bool{}
	var recipients []string
	add := func(agent string) {
		agent = sanitizeAgent(agent)
		if agent == "" || agent == sender || seen[agent] {
			return
		}
		seen[agent] = true
		recipients = append(recipients, agent)
	}
	add("human")
	for _, agent := range cfg.Agents {
		add(agent.ID)
	}
	return recipients
}

func modelSuffix(model string) string {
	if strings.TrimSpace(model) == "" {
		return ""
	}
	return "/" + strings.TrimSpace(model)
}
