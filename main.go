package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/przybylku/openbattery/battery"
	"github.com/przybylku/openbattery/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// Set via ldflags by goreleaser.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// snapshot is the machine-readable battery state printed by --json.
type snapshot struct {
	Time               time.Time `json:"time"`
	Percent            float64   `json:"percent"` // 0–100
	Status             string    `json:"status"`
	Watts              float64   `json:"watts"`
	RemainingTime      string    `json:"remaining_time"`
	EstimatedHoursLeft float64   `json:"estimated_hours_left,omitempty"`
	TimeToFullHours    float64   `json:"time_to_full_hours,omitempty"`
	CurrentCapacitymAh int64     `json:"current_capacity_mah"`
	MaxCapacitymAh     int64     `json:"max_capacity_mah"`
	DesignCapacitymAh  int64     `json:"design_capacity_mah"`
	CycleCount         int64     `json:"cycle_count"`
	HealthPercent      float64   `json:"health_percent"`
}

func main() {
	jsonFlag := flag.Bool("json", false, "print a one-shot battery snapshot as JSON and exit")
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("openbattery %s (commit %s, built %s)\n", version, commit, date)
		return
	}

	if *jsonFlag {
		if err := printJSON(); err != nil {
			fmt.Fprintf(os.Stderr, "openbattery: %v\n", err)
			os.Exit(1)
		}
		return
	}

	m := ui.NewModel()
	p := tea.NewProgram(&m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "openbattery: %v\n", err)
		os.Exit(1)
	}
}

func printJSON() error {
	pset, err := battery.ParsePmset()
	if err != nil {
		return err
	}
	ireg, err := battery.ParseIOReg()
	if err != nil {
		return err
	}

	s := snapshot{
		Time:               time.Now(),
		Percent:            pset.Percent * 100,
		Status:             pset.Status,
		Watts:              ireg.WattsNow(),
		RemainingTime:      pset.RemainingTime,
		EstimatedHoursLeft: ireg.EstimatedHours(),
		TimeToFullHours:    ireg.TimeToFullHours(),
		CurrentCapacitymAh: ireg.CurrentCapacity,
		MaxCapacitymAh:     ireg.AppleRawMaxCapacity,
		DesignCapacitymAh:  ireg.DesignCapacity,
		CycleCount:         ireg.CycleCount,
		HealthPercent:      ireg.BatteryHealth(),
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}
