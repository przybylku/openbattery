// Package ui implements the terminal user interface for openbattery
// using the Bubble Tea framework.
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/przybylku/openbattery/battery"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	green  = lipgloss.Color("#00FF87")
	yellow = lipgloss.Color("#FFD700")
	red    = lipgloss.Color("#FF4040")
	cyan   = lipgloss.Color("#00CFFF")
	white  = lipgloss.Color("#FFFFFF")
	gray   = lipgloss.Color("#808080")
	dimmed = lipgloss.Color("#666666")
)

type tickMsg struct{}

type batteryData struct {
	ioreg    battery.IOReg
	pmset    battery.Pmset
	errIoreg error
	errPmset error
}

// Model is the top-level Bubble Tea model for the openbattery TUI.
type model struct {
	activeTab int
	battery   battery.IOReg
	pmset     battery.Pmset
	history   []battery.HistoryEntry
	progress  progress.Model
	loading   bool
	width     int
	height    int
	errIoreg  error
	errPmset  error
}

// NewModel creates and returns an initial model with loaded history.
func NewModel() model {
	p := progress.New(
		progress.WithScaledGradient("#FF4040", "#00FF87"),
		progress.WithWidth(44),
		progress.WithoutPercentage(),
	)

	history, _ := battery.LoadHistory()

	return model{
		activeTab: 0,
		progress:  p,
		loading:   true,
		width:     80,
		height:    24,
		history:   history,
	}
}

func fetchBatteryData() tea.Msg {
	ireg, err1 := battery.ParseIOReg()
	pset, err2 := battery.ParsePmset()
	return batteryData{
		ioreg:    ireg,
		pmset:    pset,
		errIoreg: err1,
		errPmset: err2,
	}
}

func doTick() tea.Msg {
	time.Sleep(10 * time.Second)
	return tickMsg{}
}

func (m *model) Init() tea.Cmd {
	return fetchBatteryData
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "1":
			m.activeTab = 0
		case "2":
			m.activeTab = 1
		case "r":
			return m, fetchBatteryData
		}

	case tickMsg:
		return m, fetchBatteryData

	case batteryData:
		if msg.errPmset == nil {
			m.pmset = msg.pmset
			m.errPmset = nil
			m.progress.SetPercent(m.pmset.Percent)
		} else {
			m.pmset = battery.Pmset{Status: "error"}
			m.errPmset = msg.errPmset
		}

		if msg.errIoreg == nil {
			m.battery = msg.ioreg
			m.errIoreg = nil
		} else {
			m.battery = battery.IOReg{}
			m.errIoreg = msg.errIoreg
		}

		m.loading = false

		if msg.errPmset == nil {
			watts := 0.0
			if msg.errIoreg == nil {
				watts = m.battery.WattsNow()
			}
			entry := battery.HistoryEntry{
				Time:    time.Now(),
				Percent: m.pmset.Percent,
				Watts:   watts,
				Status:  m.pmset.Status,
			}
			_ = battery.AppendHistory(entry)
			m.history, _ = battery.LoadHistory()
		}

		return m, doTick

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = progressBarWidth(m.width)
	}

	return m, nil
}

// View renders the entire TUI.
func (m *model) View() string {
	w := m.width
	if w < 20 {
		w = 20
	}

	var s strings.Builder

	s.WriteString(headerView(m.activeTab, w))
	s.WriteString("\n")
	s.WriteString(styledSep(w))
	s.WriteString("\n")

	if m.activeTab == 0 {
		s.WriteString(dashboardView(m, w))
	} else {
		s.WriteString(historyView(m, w))
	}

	s.WriteString("\n")
	s.WriteString(styledSep(w))
	s.WriteString("\n")

	footer := "Refreshes every 10s   \u2022   r: refresh now   \u2022   q: quit"
	if w < 40 {
		footer = "10s  \u2022 r: refresh  \u2022 q: quit"
	}
	s.WriteString(lipgloss.NewStyle().Foreground(dimmed).Width(w).Align(lipgloss.Center).Render(footer))

	return s.String()
}

func styledSep(w int) string {
	return lipgloss.NewStyle().Foreground(dimmed).Render(strings.Repeat("\u2500", w))
}

func progressBarWidth(w int) int {
	switch {
	case w >= 100:
		return 60
	case w >= 80:
		return 44
	case w >= 60:
		return 32
	case w >= 40:
		return 24
	default:
		if w-8 < 4 {
			return 4
		}
		return w - 8
	}
}

func layoutBreaks(w int) (labelW, gap int, compact bool) {
	switch {
	case w >= 100:
		return 24, 2, false
	case w >= 80:
		return 20, 2, false
	case w >= 60:
		return 18, 2, false
	default:
		return 14, 0, true
	}
}

func headerView(activeTab int, w int) string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(white)

	tabActive := lipgloss.NewStyle().Bold(true).Foreground(cyan)
	tabInactive := lipgloss.NewStyle().Foreground(gray)

	tab1 := tabActive.Render("[1: Dashboard]")
	tab2 := tabInactive.Render("[2: History]")
	if activeTab != 0 {
		tab1 = tabInactive.Render("[1: Dashboard]")
		tab2 = tabActive.Render("[2: History]")
	}

	header := headerStyle.Render("\u26A1 openbattery")

	if w < 40 {
		line1 := headerStyle.Width(w).Render("\u26A1 openbattery")
		line2 := lipgloss.JoinHorizontal(lipgloss.Left, tab1, "  ", tab2)
		return line1 + "\n" + line2
	}

	full := lipgloss.JoinHorizontal(lipgloss.Center, header, "    ", tab1, "  ", tab2)
	if lipgloss.Width(full) <= w {
		return full
	}

	return headerStyle.Width(w).Render("\u26A1 openbattery")
}

func sectionTitle(title string) string {
	return lipgloss.NewStyle().Foreground(cyan).Bold(true).Render(title)
}

func dashboardView(m *model, w int) string {
	var s strings.Builder

	pw := progressBarWidth(w)
	m.progress.Width = pw
	s.WriteString(m.progress.View())
	pctStr := fmt.Sprintf("  %.0f%%", m.pmset.Percent*100)
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(percentColor(m.pmset.Percent)).Render(pctStr))
	s.WriteString("\n")

	if m.loading && m.pmset.Percent == 0 && m.battery.WattsNow() == 0 {
		s.WriteString(lipgloss.NewStyle().Foreground(dimmed).Render("Loading battery data..."))
		return s.String()
	}

	labelW, gap, compact := layoutBreaks(w)
	labelStyle := lipgloss.NewStyle().Foreground(gray).Width(labelW)
	valueStyle := lipgloss.NewStyle().Foreground(white)

	gapStr := ""
	if gap > 0 {
		gapStr = lipgloss.NewStyle().Width(gap).Render("")
	}

	s.WriteString("\n")
	s.WriteString(sectionTitle("Power Status"))
	s.WriteString("\n")

	sc := statusColor(m.pmset.Status)
	statusDot := lipgloss.NewStyle().Foreground(sc).Render("\u25CF")
	statusStr := lipgloss.NewStyle().Foreground(sc).Render(fmt.Sprintf("%s %s", statusDot, m.pmset.Status))
	wattsStr := fmt.Sprintf("%.1f W", m.battery.WattsNow())

	if compact {
		s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render("\u26A1 Power:"),
			valueStyle.Render(wattsStr),
		) + "\n")
		s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render("   Status:"),
			statusStr,
		) + "\n")
	} else {
		s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render("\u26A1  Power draw:"),
			gapStr,
			valueStyle.Render(wattsStr),
			lipgloss.NewStyle().Width(2).Render(""),
			statusStr,
		) + "\n")
	}

	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("\u23F1  Remaining:"),
		gapStr,
		valueStyle.Render(m.pmset.RemainingTime),
	) + "\n")

	source := "Battery"
	if m.pmset.Status == "charging" || m.pmset.Status == "charged" {
		source = "AC Power"
	}
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("\U0001F50C  Source:"),
		gapStr,
		valueStyle.Render(source),
	) + "\n")

	chargeStr := fmt.Sprintf("%d / %d mAh", m.battery.CurrentCapacity, m.battery.AppleRawMaxCapacity)
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("\U0001F50B  Charge:"),
		gapStr,
		valueStyle.Render(chargeStr),
	) + "\n")

	if m.pmset.Status == "discharging" && m.battery.Voltage > 0 {
		s.WriteString("\n")
		s.WriteString(sectionTitle("Smart Estimates"))
		s.WriteString("\n")

		avg1h := battery.AvgWatts(m.history, 360)
		if avg1h > 0 {
			avgAmps1h := avg1h * 1_000_000 / float64(m.battery.Voltage)
			est1h := battery.EstimatedTimeFromAvg(m.battery.CurrentCapacity, avgAmps1h)
			estStr := fmt.Sprintf("%.1f W \u2192 %s", avg1h, formatDuration(est1h))
			s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
				labelStyle.Render("\U0001F4CA  Avg (1h):"),
				gapStr,
				valueStyle.Render(estStr),
			) + "\n")
		}

		avgToday := battery.AvgWatts(m.history, 500)
		if avgToday > 0 {
			avgAmpsToday := avgToday * 1_000_000 / float64(m.battery.Voltage)
			estToday := battery.EstimatedTimeFromAvg(m.battery.CurrentCapacity, avgAmpsToday)
			estStr := fmt.Sprintf("%.1f W \u2192 %s", avgToday, formatDuration(estToday))
			s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
				labelStyle.Render("\U0001F4CA  Avg (today):"),
				gapStr,
				valueStyle.Render(estStr),
			) + "\n")
		}
	}

	s.WriteString("\n")
	s.WriteString(sectionTitle("Battery Health"))
	s.WriteString("\n")

	bh := m.battery.BatteryHealth()
	healthColor := percentColor(bh / 100.0)
	healthStyled := lipgloss.NewStyle().Foreground(healthColor).Render(fmt.Sprintf("%.1f%%", bh))
	cycleStyled := valueStyle.Render(fmt.Sprintf("%d", m.battery.CycleCount))

	if compact {
		s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render("\U0001F52C Health:"),
			healthStyled,
		) + "\n")
		s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render("\U0001F504 Cycles:"),
			cycleStyled,
		) + "\n")
	} else {
		s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render("\U0001F52C  Health:"),
			gapStr,
			healthStyled,
			lipgloss.NewStyle().Width(4).Render(""),
			lipgloss.NewStyle().Foreground(gray).Render("\U0001F504  Cycles:"),
			lipgloss.NewStyle().Width(2).Render(""),
			cycleStyled,
		) + "\n")
	}

	if m.errPmset != nil {
		s.WriteString("\n" + lipgloss.NewStyle().Foreground(red).Render("pmset: "+m.errPmset.Error()))
	}
	if m.errIoreg != nil {
		s.WriteString("\n" + lipgloss.NewStyle().Foreground(red).Render("ioreg: "+m.errIoreg.Error()))
	}

	return s.String()
}

func percentColor(p float64) lipgloss.Color {
	switch {
	case p >= 0.5:
		return green
	case p >= 0.2:
		return yellow
	default:
		return red
	}
}

func statusColor(status string) lipgloss.Color {
	switch status {
	case "charging":
		return cyan
	case "charged":
		return green
	case "discharging":
		return yellow
	default:
		return red
	}
}

func formatDuration(hours float64) string {
	if hours <= 0 {
		return "\u2014"
	}
	totalMinutes := int(hours * 60)
	h := totalMinutes / 60
	m := totalMinutes % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dmin", h, m)
	}
	return fmt.Sprintf("%dmin", m)
}