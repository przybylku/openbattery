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

type tickMsg struct{}

type batteryData struct {
	ioreg    battery.IOReg
	pmset    battery.Pmset
	errIoreg error
	errPmset error
}

// model is the top-level Bubble Tea model for the openbattery TUI.
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
		progress.WithScaledGradient("#FF6B6B", "#69DB7C"),
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

// scheduleTick arms a single refresh timer. Only tickMsg reschedules it,
// so manual refreshes ("r") never spawn additional timer loops.
func scheduleTick() tea.Cmd {
	return tea.Tick(10*time.Second, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(fetchBatteryData, scheduleTick())
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
		return m, tea.Batch(fetchBatteryData, scheduleTick())

	case batteryData:
		var cmds []tea.Cmd
		if msg.errPmset == nil {
			m.pmset = msg.pmset
			m.errPmset = nil
			cmds = append(cmds, m.progress.SetPercent(m.pmset.Percent))
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
			if updated, err := battery.LoadHistory(); err == nil {
				m.history = updated
			}
		}

		return m, tea.Batch(cmds...)

	case progress.FrameMsg:
		pm, cmd := m.progress.Update(msg)
		m.progress = pm.(progress.Model)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the entire TUI.
func (m *model) View() string {
	w := m.width
	if w < 24 {
		w = 24
	}

	var s strings.Builder
	s.WriteString(headerView(m.activeTab, w))
	s.WriteString("\n\n")

	if m.activeTab == 0 {
		s.WriteString(dashboardView(m, w))
	} else {
		s.WriteString(historyView(m, w))
	}

	s.WriteString("\n\n")
	s.WriteString(footerView(w))
	return s.String()
}

func headerView(activeTab int, w int) string {
	logo := lipgloss.NewStyle().Bold(true).Foreground(white).Render("⚡ openbattery")

	active := lipgloss.NewStyle().Bold(true).
		Background(cyan).Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#101418"}).
		Padding(0, 1)
	inactive := lipgloss.NewStyle().Foreground(gray).Padding(0, 1)

	t1, t2 := inactive, active
	if activeTab == 0 {
		t1, t2 = active, inactive
	}
	tabs := lipgloss.JoinHorizontal(lipgloss.Center,
		t1.Render("1 dashboard"), " ", t2.Render("2 history"))

	gap := w - lipgloss.Width(logo) - lipgloss.Width(tabs)
	if gap >= 2 {
		return logo + strings.Repeat(" ", gap) + tabs
	}
	if w >= 40 {
		return logo + "\n" + tabs
	}
	return logo
}

func footerView(w int) string {
	var footer string
	switch {
	case w >= 56:
		footer = lipgloss.JoinHorizontal(lipgloss.Center,
			keyHint("1·2", "tabs"), "   ",
			keyHint("r", "refresh"), "   ",
			keyHint("q", "quit"),
			lipgloss.NewStyle().Foreground(dimmed).Render("   · auto 10s"))
	case w >= 36:
		footer = lipgloss.JoinHorizontal(lipgloss.Center,
			keyHint("1·2", "tabs"), "  ",
			keyHint("r", "refresh"), "  ",
			keyHint("q", "quit"))
	default:
		footer = lipgloss.JoinHorizontal(lipgloss.Center,
			keyHint("r", "refresh"), "  ", keyHint("q", "quit"))
	}
	return lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Render(footer)
}

func dashboardView(m *model, w int) string {
	if m.loading && m.pmset.Percent == 0 && m.battery.WattsNow() == 0 {
		return card("battery", lipgloss.NewStyle().Foreground(dimmed).Render("Loading battery data…"), w, cyan)
	}

	var s strings.Builder
	s.WriteString(batteryCard(m, w))

	if w >= 88 {
		leftW := (w - 2) / 2
		rightW := w - 2 - leftW
		s.WriteString("\n")
		s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
			powerCard(m, leftW), "  ", healthCard(m, rightW)))
	} else {
		s.WriteString("\n")
		s.WriteString(powerCard(m, w))
		s.WriteString("\n")
		s.WriteString(healthCard(m, w))
	}

	if est := estimatesCard(m, w); est != "" {
		s.WriteString("\n")
		s.WriteString(est)
	}

	if m.errPmset != nil {
		s.WriteString("\n" + lipgloss.NewStyle().Foreground(red).Render("pmset: "+m.errPmset.Error()))
	}
	if m.errIoreg != nil {
		s.WriteString("\n" + lipgloss.NewStyle().Foreground(red).Render("ioreg: "+m.errIoreg.Error()))
	}
	return s.String()
}

// batteryCard is the signature gauge: oversized block digits next to the
// gradient bar, status and time estimates. Falls back to a compact layout
// on narrow terminals.
func batteryCard(m *model, w int) string {
	inner := w - 4
	pct := m.pmset.Percent
	pctText := fmt.Sprintf("%.0f%%", pct*100)

	statusLine := lipgloss.NewStyle().Foreground(statusColor(m.pmset.Status)).
		Render("● " + m.pmset.Status)

	timeLine := lipgloss.NewStyle().Foreground(gray).Render("left ") +
		lipgloss.NewStyle().Foreground(white).Render(m.pmset.RemainingTime)
	if m.pmset.Status == "charging" {
		if toFull := m.battery.TimeToFullHours(); toFull > 0 {
			timeLine = lipgloss.NewStyle().Foreground(gray).Render("to full ") +
				lipgloss.NewStyle().Foreground(white).Render(formatDuration(toFull))
		}
	}

	source := "battery"
	if m.pmset.Status == "charging" || m.pmset.Status == "charged" {
		source = "ac power"
	}
	sourceLine := lipgloss.NewStyle().Foreground(gray).Render("source ") +
		lipgloss.NewStyle().Foreground(white).Render(source)

	digitsW := BigDigitsWidth(pctText)
	if inner >= digitsW+3+16 {
		big := BigDigits(pctText, percentColor(pct))
		rightW := inner - digitsW - 3
		if rightW > 60 {
			rightW = 60
		}
		m.progress.Width = rightW
		right := lipgloss.JoinVertical(lipgloss.Left,
			statusLine,
			"",
			m.progress.View(),
			timeLine,
			sourceLine,
		)
		content := lipgloss.JoinHorizontal(lipgloss.Center, big, "   ", right)
		return card("battery", content, w, cyan)
	}

	barW := inner - 6
	if barW < 4 {
		barW = 4
	}
	m.progress.Width = barW
	barLine := m.progress.View() + " " +
		lipgloss.NewStyle().Bold(true).Foreground(percentColor(pct)).Render(pctText)
	lines := []string{
		barLine,
		statusLine + "  " + lipgloss.NewStyle().Foreground(white).Render(fmt.Sprintf("%.1f W", m.battery.WattsNow())),
	}
	if lipgloss.Width(timeLine)+2+lipgloss.Width(sourceLine) <= inner {
		lines = append(lines, timeLine+"  "+sourceLine)
	} else {
		lines = append(lines, timeLine, sourceLine)
	}
	return card("battery", lipgloss.JoinVertical(lipgloss.Left, lines...), w, cyan)
}

// narrowLabels reports whether cards should use compact labels/values.
func narrowLabels(w int) bool { return w-4 < 26 }

func powerCard(m *model, w int) string {
	vs := lipgloss.NewStyle().Foreground(white)
	labelW, chargeFmt := 10, "%d / %d mAh"
	if narrowLabels(w) {
		chargeFmt = "%d/%d"
	}
	content := lipgloss.JoinVertical(lipgloss.Left,
		row("draw", labelW, vs.Render(fmt.Sprintf("%.1f W", m.battery.WattsNow()))),
		row("charge", labelW, vs.Render(fmt.Sprintf(chargeFmt, m.battery.CurrentCapacity, m.battery.AppleRawMaxCapacity))),
		row("voltage", labelW, vs.Render(fmt.Sprintf("%.2f V", float64(m.battery.Voltage)/1000))),
	)
	return card("power", content, w, cyan)
}

func healthCard(m *model, w int) string {
	vs := lipgloss.NewStyle().Foreground(white)
	bh := m.battery.BatteryHealth()
	healthVal := lipgloss.NewStyle().Foreground(percentColor(bh / 100)).Render(fmt.Sprintf("%.1f%%", bh))
	labelW, capFmt := 10, "%d / %d mAh"
	if narrowLabels(w) {
		capFmt = "%d/%d"
	}
	content := lipgloss.JoinVertical(lipgloss.Left,
		row("health", labelW, healthVal),
		row("cycles", labelW, vs.Render(fmt.Sprintf("%d", m.battery.CycleCount))),
		row("capacity", labelW, vs.Render(fmt.Sprintf(capFmt, m.battery.AppleRawMaxCapacity, m.battery.DesignCapacity))),
	)
	return card("health", content, w, cyan)
}

func estimatesCard(m *model, w int) string {
	if m.pmset.Status != "discharging" || m.battery.Voltage <= 0 {
		return ""
	}

	arrow := lipgloss.NewStyle().Foreground(dimmed).Render(" → ")
	vs := lipgloss.NewStyle().Foreground(white)
	labelW := 10
	narrow := narrowLabels(w)
	estVal := func(avg, est float64) string {
		if narrow {
			return vs.Render(formatDuration(est))
		}
		return vs.Render(fmt.Sprintf("%.1f W", avg)) + arrow + vs.Render(formatDuration(est))
	}
	var lines []string

	if avg := battery.AvgDischargeWatts(m.history, time.Now().Add(-time.Hour)); avg > 0 {
		amps := avg * 1_000_000 / float64(m.battery.Voltage)
		est := battery.EstimatedTimeFromAvg(m.battery.CurrentCapacity, amps)
		lines = append(lines, row("last hour", labelW, estVal(avg, est)))
	}
	if avg := battery.AvgDischargeWatts(m.history, battery.StartOfToday()); avg > 0 {
		amps := avg * 1_000_000 / float64(m.battery.Voltage)
		est := battery.EstimatedTimeFromAvg(m.battery.CurrentCapacity, amps)
		lines = append(lines, row("today", labelW, estVal(avg, est)))
	}

	if len(lines) == 0 {
		return ""
	}
	return card("estimates", lipgloss.JoinVertical(lipgloss.Left, lines...), w, cyan)
}

func percentColor(p float64) lipgloss.TerminalColor {
	switch {
	case p >= 0.5:
		return green
	case p >= 0.2:
		return yellow
	default:
		return red
	}
}

func statusColor(status string) lipgloss.TerminalColor {
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
		return "—"
	}
	totalMinutes := int(hours * 60)
	h := totalMinutes / 60
	m := totalMinutes % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dmin", h, m)
	}
	return fmt.Sprintf("%dmin", m)
}
