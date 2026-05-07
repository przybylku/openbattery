// Package battery provides parsers for macOS battery data sources
// (pmset and ioreg) and persistent history storage.
package battery

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

// IOReg holds battery hardware data parsed from macOS ioreg.
type IOReg struct {
	Amperage            int64 // mA (negative when discharging)
	Voltage             int64 // mV
	CurrentCapacity     int64 // mAh
	AppleRawMaxCapacity int64 // mAh
	DesignCapacity      int64 // mAh
	CycleCount          int64
}

// ParseIOReg runs `ioreg -rn AppleSmartBattery -w0` and parses the output.
// It tries AppleRawCurrentCapacity first (accurate on Apple Silicon),
// falling back to CurrentCapacity (correct on Intel Macs).
// For max capacity it checks: AppleRawMaxCapacity → NominalChargeCapacity → MaxCapacity → DesignCapacity.
func ParseIOReg() (IOReg, error) {
	out, err := exec.Command("ioreg", "-rn", "AppleSmartBattery", "-w0").Output()
	if err != nil {
		return IOReg{}, fmt.Errorf("ioreg: %w", err)
	}

	text := string(out)

	amperage := findIntField(text, "Amperage")
	voltage := findIntField(text, "Voltage")

	currentCapacity := findIntField(text, "AppleRawCurrentCapacity")
	if currentCapacity == 0 {
		currentCapacity = findIntField(text, "CurrentCapacity")
	}

	rawMax := findIntField(text, "AppleRawMaxCapacity")
	if rawMax == 0 {
		rawMax = findIntField(text, "NominalChargeCapacity")
	}
	if rawMax == 0 {
		rawMax = findIntField(text, "MaxCapacity")
	}
	if rawMax == 0 {
		rawMax = findIntField(text, "DesignCapacity")
	}

	designCapacity := findIntField(text, "DesignCapacity")
	cycleCount := findIntField(text, "CycleCount")

	return IOReg{
		Amperage:            amperage,
		Voltage:             voltage,
		CurrentCapacity:     currentCapacity,
		AppleRawMaxCapacity: rawMax,
		DesignCapacity:      designCapacity,
		CycleCount:          cycleCount,
	}, nil
}

// findIntField parses `"Key" = Value` from ioreg output.
// Handles unsigned values that represent negative numbers (two's complement in uint64).
func findIntField(text, key string) int64 {
	pattern := fmt.Sprintf(`"%s"\s*=\s*(-?\d+)`, regexp.QuoteMeta(key))
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if matches == nil {
		return 0
	}

	// Try unsigned first (ioreg uses unsigned representation for negative amperage values)
	if val, err := strconv.ParseUint(matches[1], 10, 64); err == nil {
		return int64(val)
	}
	if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
		return val
	}
	return 0
}

// WattsNow returns current power draw in watts.
func (b IOReg) WattsNow() float64 {
	return float64(abs64(b.Amperage)*b.Voltage) / 1_000_000.0
}

// EstimatedHours returns estimated remaining hours based on current capacity and discharge rate.
// Returns 0 when charging or idle (Amperage >= 0).
func (b IOReg) EstimatedHours() float64 {
	if b.Amperage >= 0 {
		return 0
	}
	amps := float64(abs64(b.Amperage))
	if amps == 0 {
		return 0
	}
	return float64(b.CurrentCapacity) / amps
}

// BatteryHealth returns battery health as a percentage (max capacity / design capacity * 100).
// Returns 0 if DesignCapacity is not available.
func (b IOReg) BatteryHealth() float64 {
	if b.DesignCapacity == 0 {
		return 0
	}
	return float64(b.AppleRawMaxCapacity) / float64(b.DesignCapacity) * 100
}

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}