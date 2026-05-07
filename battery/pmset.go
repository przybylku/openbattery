// Package battery provides parsers for macOS battery data sources
// (pmset and ioreg) and persistent history storage.
package battery

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Pmset holds battery status parsed from the macOS pmset command.
type Pmset struct {
	Percent       float64 // Battery level 0.0–1.0
	RemainingTime string  // Human-readable remaining time (e.g. "4h 32min")
	Status        string  // One of: "discharging", "charging", "charged", "error"
}

var pmsetRe = regexp.MustCompile(`(\d+)%;\s*([^;]*);?\s*(.*)`)

// ParsePmset runs `pmset -g batt` and parses the output into a Pmset struct.
func ParsePmset() (Pmset, error) {
	out, err := exec.Command("pmset", "-g", "batt").Output()
	if err != nil {
		return Pmset{Status: "error"}, fmt.Errorf("pmset: %w", err)
	}

	matches := pmsetRe.FindStringSubmatch(string(out))
	if matches == nil {
		return Pmset{Status: "error"}, fmt.Errorf("pmset: could not parse output")
	}

	percent, err := strconv.Atoi(matches[1])
	if err != nil {
		return Pmset{Status: "error"}, fmt.Errorf("pmset: invalid percentage: %w", err)
	}

	p := float64(percent) / 100.0
	if p > 1 {
		p = 1
	}
	if p < 0 {
		p = 0
	}

	status := strings.TrimSpace(matches[2])
	remainder := strings.TrimSpace(matches[3])

	return Pmset{
		Percent:       p,
		RemainingTime: formatPmsetTime(remainder),
		Status:        normalizeStatus(status),
	}, nil
}

func normalizeStatus(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "discharging":
		return "discharging"
	case "charging":
		return "charging"
	case "finishing charge":
		return "charging"
	case "charged", "ac attached":
		return "charged"
	default:
		return "error"
	}
}

var pmsetTimeRe = regexp.MustCompile(`(\d+):(\d+)`)

func formatPmsetTime(raw string) string {
	if raw == "" || strings.Contains(strings.ToLower(raw), "no estimate") {
		return "\u2014"
	}
	matches := pmsetTimeRe.FindStringSubmatch(raw)
	if matches == nil {
		return raw
	}
	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	if hours > 0 {
		return fmt.Sprintf("%dh %dmin", hours, minutes)
	}
	return fmt.Sprintf("%dmin", minutes)
}