package battery

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadHistoryEmpty(t *testing.T) {
	h, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory() error: %v", err)
	}
	if h == nil {
		t.Fatal("LoadHistory() returned nil slice")
	}
}

func TestAppendAndLoadHistory(t *testing.T) {
	origDir, _ := historyDir()
	tmpDir := t.TempDir()
	home := os.Getenv("HOME")
	defer func() { _ = os.Setenv("HOME", home) }()
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("setenv HOME: %v", err)
	}

	_ = os.MkdirAll(filepath.Join(tmpDir, ".openbattery"), 0755)

	entry1 := HistoryEntry{
		Time:    time.Now(),
		Percent: 0.72,
		Watts:   8.4,
		Status:  "discharging",
	}
	if err := AppendHistory(entry1); err != nil {
		t.Fatalf("AppendHistory: %v", err)
	}

	entry2 := HistoryEntry{
		Time:    time.Now().Add(10 * time.Second),
		Percent: 0.71,
		Watts:   8.5,
		Status:  "discharging",
	}
	if err := AppendHistory(entry2); err != nil {
		t.Fatalf("AppendHistory: %v", err)
	}

	loaded, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(loaded))
	}
	if loaded[0].Percent != 0.72 {
		t.Errorf("first entry percent = %f, want 0.72", loaded[0].Percent)
	}
	if loaded[1].Watts != 8.5 {
		t.Errorf("second entry watts = %f, want 8.5", loaded[1].Watts)
	}

	_ = origDir
}

func TestAvgDischargeWatts(t *testing.T) {
	now := time.Now()
	history := []HistoryEntry{
		{Time: now.Add(-2 * time.Hour), Watts: 50.0, Status: "discharging"}, // outside window
		{Time: now.Add(-30 * time.Minute), Watts: 5.0, Status: "discharging"},
		{Time: now.Add(-20 * time.Minute), Watts: 0, Status: "discharging"}, // no wattage
		{Time: now.Add(-10 * time.Minute), Watts: 40.0, Status: "charging"}, // charging excluded
		{Time: now.Add(-5 * time.Minute), Watts: 15.0, Status: "discharging"},
	}

	avg := AvgDischargeWatts(history, now.Add(-time.Hour))
	if avg != 10.0 {
		t.Errorf("AvgDischargeWatts(1h) = %f, want 10.0 (avg of 5 and 15)", avg)
	}

	avgAll := AvgDischargeWatts(history, now.Add(-3*time.Hour))
	if want := 70.0 / 3.0; avgAll != want {
		t.Errorf("AvgDischargeWatts(3h) = %f, want %f (avg of 50, 5, 15)", avgAll, want)
	}

	avgNone := AvgDischargeWatts([]HistoryEntry{}, now)
	if avgNone != 0 {
		t.Errorf("AvgDischargeWatts(empty) = %f, want 0", avgNone)
	}

	avgChargingOnly := AvgDischargeWatts([]HistoryEntry{
		{Time: now, Watts: 30.0, Status: "charging"},
	}, now.Add(-time.Hour))
	if avgChargingOnly != 0 {
		t.Errorf("AvgDischargeWatts(charging only) = %f, want 0", avgChargingOnly)
	}
}

func TestStartOfToday(t *testing.T) {
	sot := StartOfToday()
	now := time.Now()
	if sot.After(now) {
		t.Errorf("StartOfToday() = %v, is in the future", sot)
	}
	if now.Sub(sot) >= 24*time.Hour {
		t.Errorf("StartOfToday() = %v, more than 24h ago", sot)
	}
	if sot.Hour() != 0 || sot.Minute() != 0 || sot.Second() != 0 {
		t.Errorf("StartOfToday() = %v, want midnight", sot)
	}
}

func TestEstimatedTimeFromAvg(t *testing.T) {
	est := EstimatedTimeFromAvg(323, 661)
	if est < 0.48 || est > 0.50 {
		t.Errorf("EstimatedTimeFromAvg(323, 661) = %f, want ~0.489", est)
	}

	estZero := EstimatedTimeFromAvg(0, 100)
	if estZero != 0 {
		t.Errorf("EstimatedTimeFromAvg(0, 100) = %f, want 0", estZero)
	}

	estNegative := EstimatedTimeFromAvg(100, -5)
	if estNegative != 0 {
		t.Errorf("EstimatedTimeFromAvg negative amps = %f, want 0", estNegative)
	}
}
