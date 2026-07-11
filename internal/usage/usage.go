// Package usage stores provider-independent quota snapshots and threshold alerts.
package usage

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

const DefaultThresholdStep = 20

type Snapshot struct {
	Agent         string  `json:"agent"`
	Model         string  `json:"model,omitempty"`
	Period        string  `json:"period,omitempty"`
	Used          float64 `json:"used"`
	Limit         float64 `json:"limit"`
	RefreshAt     string  `json:"refresh_at"`
	UpdatedAt     string  `json:"updated_at"`
	ThresholdStep int     `json:"threshold_step"`
	Triggered     []int   `json:"triggered,omitempty"`
}

type Alert struct {
	At        string  `json:"at"`
	Agent     string  `json:"agent"`
	Model     string  `json:"model,omitempty"`
	Percent   float64 `json:"percent"`
	Threshold int     `json:"threshold"`
	RefreshAt string  `json:"refresh_at"`
}

func (s Snapshot) Percent() float64 {
	if s.Limit <= 0 {
		return 0
	}
	return s.Used / s.Limit * 100
}

func (s Snapshot) RefreshTime() (time.Time, error) {
	return time.Parse(time.RFC3339, s.RefreshAt)
}

func (s Snapshot) RefreshDue(now time.Time) bool {
	refresh, err := s.RefreshTime()
	return err == nil && !now.Before(refresh)
}

func ValidateAgent(agent string) error {
	if agent == "" {
		return errors.New("agent is required")
	}
	for _, r := range agent {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' && r != '-' {
			return fmt.Errorf("unsafe agent id %q", agent)
		}
	}
	return nil
}

func Set(dir string, next Snapshot, now time.Time) (Snapshot, []Alert, error) {
	if err := ValidateAgent(next.Agent); err != nil {
		return Snapshot{}, nil, err
	}
	if len(next.Model) > 120 || len(next.Period) > 64 {
		return Snapshot{}, nil, errors.New("model must be <= 120 characters and period <= 64 characters")
	}
	if next.Used < 0 || next.Limit <= 0 || next.Used > next.Limit {
		return Snapshot{}, nil, errors.New("usage requires 0 <= used <= limit and limit > 0")
	}
	if _, err := next.RefreshTime(); err != nil {
		return Snapshot{}, nil, fmt.Errorf("refresh must be RFC3339: %w", err)
	}
	if next.ThresholdStep <= 0 || next.ThresholdStep > 100 {
		next.ThresholdStep = DefaultThresholdStep
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return Snapshot{}, nil, err
	}
	lockPath := filepath.Join(dir, next.Agent+".lock")
	if err := metalock.Acquire(lockPath, 5*time.Second, 10*time.Second); err != nil {
		return Snapshot{}, nil, err
	}
	defer metalock.Release(lockPath)

	previous, _ := Load(dir, next.Agent)
	if previous.RefreshAt == next.RefreshAt && previous.ThresholdStep == next.ThresholdStep && next.Used >= previous.Used {
		next.Triggered = append([]int(nil), previous.Triggered...)
	}
	next.UpdatedAt = now.UTC().Format(time.RFC3339)

	seen := map[int]bool{}
	for _, threshold := range next.Triggered {
		seen[threshold] = true
	}
	var alerts []Alert
	percent := next.Percent()
	for threshold := next.ThresholdStep; threshold <= 100; threshold += next.ThresholdStep {
		if percent+1e-9 < float64(threshold) || seen[threshold] {
			continue
		}
		next.Triggered = append(next.Triggered, threshold)
		alerts = append(alerts, Alert{
			At: now.UTC().Format(time.RFC3339), Agent: next.Agent, Model: next.Model,
			Percent: percent, Threshold: threshold, RefreshAt: next.RefreshAt,
		})
	}
	sort.Ints(next.Triggered)
	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return Snapshot{}, nil, err
	}
	data = append(data, '\n')
	if err := platform.AtomicWrite(snapshotPath(dir, next.Agent), data, 0o600); err != nil {
		return Snapshot{}, nil, err
	}
	for _, alert := range alerts {
		if err := appendAlert(dir, alert); err != nil {
			return Snapshot{}, nil, err
		}
	}
	return next, alerts, nil
}

func Load(dir, agent string) (Snapshot, error) {
	if err := ValidateAgent(agent); err != nil {
		return Snapshot{}, err
	}
	data, err := os.ReadFile(snapshotPath(dir, agent))
	if err != nil {
		return Snapshot{}, err
	}
	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return Snapshot{}, err
	}
	return snapshot, nil
}

func List(dir string) ([]Snapshot, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var snapshots []Snapshot
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") || entry.Name() == "alerts.json" {
			continue
		}
		agent := strings.TrimSuffix(entry.Name(), ".json")
		snapshot, err := Load(dir, agent)
		if err == nil {
			snapshots = append(snapshots, snapshot)
		}
	}
	sort.Slice(snapshots, func(i, j int) bool { return snapshots[i].Agent < snapshots[j].Agent })
	return snapshots, nil
}

func ReadAlerts(dir string, n int) ([]Alert, error) {
	data, err := os.ReadFile(filepath.Join(dir, "alerts.jsonl"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var alerts []Alert
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var alert Alert
		if json.Unmarshal([]byte(line), &alert) == nil {
			alerts = append(alerts, alert)
		}
	}
	if n > 0 && len(alerts) > n {
		alerts = alerts[len(alerts)-n:]
	}
	return alerts, nil
}

func snapshotPath(dir, agent string) string { return filepath.Join(dir, agent+".json") }

func appendAlert(dir string, alert Alert) error {
	lockPath := filepath.Join(dir, "alerts.lock")
	if err := metalock.Acquire(lockPath, 5*time.Second, 10*time.Second); err != nil {
		return err
	}
	defer metalock.Release(lockPath)
	data, err := json.Marshal(alert)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, "alerts.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(data, '\n'))
	return err
}
