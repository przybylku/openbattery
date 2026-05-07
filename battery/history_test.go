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
	defer os.Setenv("HOME", home)
	os.Setenv("HOME", tmpDir)

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

func TestAvgWatts(t *testing.T) {
	now := time.Now()
	history := []HistoryEntry{
		{Time: now, Watts: 5.0},
		{Time: now.Add(time.Second), Watts: 0},
		{Time: now.Add(2 * time.Second), Watts: 10.0},
		{Time: now.Add(3 * time.Second), Watts: 15.0},
	}

	avg := AvgWatts(history, 3)
	if avg != 12.5 {
		t.Errorf("AvgWatts(last 3) = %f, want 12.5 (avg of 10 and 15, skipping 0)", avg)
	}

	avgAll := AvgWatts(history, 10)
	if avgAll != 10.0 {
		t.Errorf("AvgWatts(all) = %f, want 10.0 (avg of 5, 10, 15, skipping 0)", avgAll)
	}

	avgNone := AvgWatts([]HistoryEntry{}, 5)
	if avgNone != 0 {
		t.Errorf("AvgWatts(empty) = %f, want 0", avgNone)
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
