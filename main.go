package main

import (
	"fmt"
	"os"

	"github.com/przybylku/openbattery/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := ui.NewModel()
	p := tea.NewProgram(&m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "openbattery: %v\n", err)
		os.Exit(1)
	}
}