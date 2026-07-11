package usage

import (
	"testing"
	"time"
)

func TestSetTriggersEachTwentyPercentOnce(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	base := Snapshot{Agent: "claude", Model: "opus", Used: 42, Limit: 100, RefreshAt: now.Add(time.Hour).Format(time.RFC3339)}
	got, alerts, err := Set(dir, base, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(alerts) != 2 || alerts[0].Threshold != 20 || alerts[1].Threshold != 40 {
		t.Fatalf("alerts = %#v", alerts)
	}
	if got.ThresholdStep != 20 {
		t.Fatalf("threshold step = %d", got.ThresholdStep)
	}
	base.Used = 45
	_, alerts, err = Set(dir, base, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(alerts) != 0 {
		t.Fatalf("duplicate alerts = %#v", alerts)
	}
}

func TestNewWindowResetsThresholds(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)
	first := Snapshot{Agent: "codex", Used: 80, Limit: 100, RefreshAt: now.Add(time.Hour).Format(time.RFC3339)}
	if _, _, err := Set(dir, first, now); err != nil {
		t.Fatal(err)
	}
	second := first
	second.Used = 20
	second.RefreshAt = now.Add(8 * 24 * time.Hour).Format(time.RFC3339)
	_, alerts, err := Set(dir, second, now.Add(2*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(alerts) != 1 || alerts[0].Threshold != 20 {
		t.Fatalf("new-window alerts = %#v", alerts)
	}
}

func TestRejectsUnsafeAgentAndInvalidUsage(t *testing.T) {
	now := time.Now().UTC()
	_, _, err := Set(t.TempDir(), Snapshot{Agent: "../x", Used: 1, Limit: 2, RefreshAt: now.Format(time.RFC3339)}, now)
	if err == nil {
		t.Fatal("unsafe agent should fail")
	}
	_, _, err = Set(t.TempDir(), Snapshot{Agent: "codex", Used: 3, Limit: 2, RefreshAt: now.Format(time.RFC3339)}, now)
	if err == nil {
		t.Fatal("usage over limit should fail")
	}
}
