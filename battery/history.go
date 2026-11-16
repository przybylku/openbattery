package battery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// HistoryEntry represents a single battery measurement snapshot stored in history.
type HistoryEntry struct {
	Time    time.Time `json:"time"`
	Percent float64   `json:"percent"`
	Watts   float64   `json:"watts"`
	Status  string    `json:"status"`
}

const maxHistoryEntries = 500

func historyDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".openbattery"), nil
}

func historyPath() (string, error) {
	dir, err := historyDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "history.json"), nil
}

func ensureDir() error {
	dir, err := historyDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}

// LoadHistory reads the history file from disk. Returns an empty slice
// if the file does not exist or cannot be parsed.
func LoadHistory() ([]HistoryEntry, error) {
	path, err := historyPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []HistoryEntry{}, nil
		}
		return nil, fmt.Errorf("read history: %w", err)
	}
	var entries []HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse history: %w", err)
	}
	// Clamp values to sane ranges on load
	for i := range entries {
		if entries[i].Percent < 0 {
			entries[i].Percent = 0
		}
		if entries[i].Percent > 1 {
			entries[i].Percent = 1
		}
	}
	return entries, nil
}

// AppendHistory adds an entry and saves, trimming to maxHistoryEntries.
func AppendHistory(entry HistoryEntry) error {
	entries, err := LoadHistory()
	if err != nil {
		return fmt.Errorf("load history: %w", err)
	}
	entries = append(entries, entry)
	if len(entries) > maxHistoryEntries {
		entries = entries[len(entries)-maxHistoryEntries:]
	}
	return SaveHistory(entries)
}

// SaveHistory writes the history slice to disk as pretty-printed JSON.
// The write is atomic (temp file + rename) so an interrupted write can never
// leave a truncated history file behind.
func SaveHistory(entries []HistoryEntry) error {
	if err := ensureDir(); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal history: %w", err)
	}
	path, err := historyPath()
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".history-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp history: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write history: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close history: %w", err)
	}
	if err := os.Chmod(tmp.Name(), 0644); err != nil {
		return fmt.Errorf("chmod history: %w", err)
	}
	return os.Rename(tmp.Name(), path)
}

// AvgDischargeWatts returns the average power draw of discharging samples
// taken at or after the given time. Charging samples are excluded because
// their wattage reflects charge input, not consumption. Entries without a
// positive wattage are skipped. Returns 0 if no valid entries exist.
func AvgDischargeWatts(history []HistoryEntry, since time.Time) float64 {
	var sum float64
	count := 0
	for _, e := range history {
		if e.Time.Before(since) {
			continue
		}
		if e.Status != "discharging" || e.Watts <= 0 {
			continue
		}
		sum += e.Watts
		count++
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// StartOfToday returns midnight of the current local day.
func StartOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// EstimatedTimeFromAvg returns estimated remaining hours given current capacity (mAh)
// and average amperage (mA). Returns 0 if inputs are invalid.
func EstimatedTimeFromAvg(currentCapacity int64, avgAmps float64) float64 {
	if avgAmps <= 0 || currentCapacity <= 0 {
		return 0
	}
	return float64(currentCapacity) / avgAmps
}
