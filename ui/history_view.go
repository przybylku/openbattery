package ui

import (
	"fmt"
	"strings"

	"github.com/przybylku/openbattery/battery"
	"github.com/charmbracelet/lipgloss"
)

func historyView(m *model, w int) string {
	dimStyle := lipgloss.NewStyle().Foreground(dimmed)

	if len(m.history) == 0 {
		return dimStyle.Width(w).Align(lipgloss.Center).Render("No history data yet. Waiting for first sample...")
	}

	var s strings.Builder

	percents := make([]float64, len(m.history))
	for i, entry := range m.history {
		percents[i] = entry.Percent
	}

	// Sparkline width: labels (4+4) + spaces (2) = 10, leaving w-10 for the chart
	sparkWidth := w - 10
	if sparkWidth < 2 {
		sparkWidth = 2
	}

	s.WriteString(sectionTitle(fmt.Sprintf("Battery Level (%d samples)", min(len(m.history), 60))))
	s.WriteString("\n")

	hiLabel := lipgloss.NewStyle().Foreground(green).Width(4).Align(lipgloss.Right).Render("100%")
	loLabel := lipgloss.NewStyle().Foreground(red).Width(4).Align(lipgloss.Right).Render("0%")
	sparkStyled := lipgloss.NewStyle().Foreground(yellow).Render(Sparkline(percents, sparkWidth))

	sparkLine := hiLabel + " " + sparkStyled + " " + loLabel
	s.WriteString(lipgloss.NewStyle().Width(w).Render(sparkLine))
	s.WriteString("\n")

	loLabel2 := lipgloss.NewStyle().Foreground(dimmed).Width(4).Align(lipgloss.Left).Render("0%")
	hiLabel2 := lipgloss.NewStyle().Foreground(dimmed).Width(4).Align(lipgloss.Right).Render("100%")
	firstT := m.history[0].Time.Format("15:04")
	lastT := m.history[len(m.history)-1].Time.Format("15:04")
	timeLine := lipgloss.JoinHorizontal(lipgloss.Top,
		loLabel2,
		lipgloss.NewStyle().Width(sparkWidth).Align(lipgloss.Center).Render(firstT+" \u2014 "+lastT),
		hiLabel2,
	)
	s.WriteString(lipgloss.NewStyle().Width(w).Render(timeLine))
	s.WriteString("\n\n")

	start := len(m.history) - 10
	if start < 0 {
		start = 0
	}
	recent := m.history[start:]

	s.WriteString(sectionTitle("Recent Activity"))
	s.WriteString("\n")
	s.WriteString(historyTable(recent, w))
	return s.String()
}

func historyTable(entries []battery.HistoryEntry, w int) string {
	timeW, pctW, wattsW, statusW := tableColumnWidths(w)

	colTime := lipgloss.NewStyle().Foreground(dimmed).Width(timeW).Align(lipgloss.Left)
	colPct := lipgloss.NewStyle().Foreground(dimmed).Width(pctW).Align(lipgloss.Left)

	var sb strings.Builder

	if wattsW > 0 {
		colWatts := lipgloss.NewStyle().Foreground(dimmed).Width(wattsW).Align(lipgloss.Left)
		colStatus := lipgloss.NewStyle().Foreground(dimmed).Width(statusW).Align(lipgloss.Left)

		sb.WriteString(colTime.Render("Time"))
		sb.WriteString("\u2502")
		sb.WriteString(colPct.Render("%"))
		sb.WriteString("\u2502")
		sb.WriteString(colWatts.Render("Watts"))
		sb.WriteString("\u2502")
		sb.WriteString(colStatus.Render("Status"))
		sb.WriteString("\n")

		sb.WriteString(strings.Repeat("\u2500", timeW) + "\u253C")
		sb.WriteString(strings.Repeat("\u2500", pctW) + "\u253C")
		sb.WriteString(strings.Repeat("\u2500", wattsW) + "\u253C")
		sb.WriteString(strings.Repeat("\u2500", statusW))
		sb.WriteString("\n")
	} else {
		colStatus := lipgloss.NewStyle().Foreground(dimmed).Width(statusW).Align(lipgloss.Left)

		sb.WriteString(colTime.Render("Time"))
		sb.WriteString("\u2502")
		sb.WriteString(colPct.Render("%"))
		sb.WriteString("\u2502")
		sb.WriteString(colStatus.Render("Status"))
		sb.WriteString("\n")

		sb.WriteString(strings.Repeat("\u2500", timeW) + "\u253C")
		sb.WriteString(strings.Repeat("\u2500", pctW) + "\u253C")
		sb.WriteString(strings.Repeat("\u2500", statusW))
		sb.WriteString("\n")
	}

	for _, entry := range entries {
		timeStr := entry.Time.Format("15:04:05")
		if timeW < 8 {
			timeStr = entry.Time.Format("15:04")
		}
		pctStr := fmt.Sprintf("%.0f%%", entry.Percent*100)

		pctStyled := lipgloss.NewStyle().Foreground(percentColor(entry.Percent)).
			Width(pctW).Align(lipgloss.Left).Render(pctStr)

		sb.WriteString(colTime.Render(timeStr))
		sb.WriteString("\u2502")
		sb.WriteString(pctStyled)
		sb.WriteString("\u2502")

		if wattsW > 0 {
			wattsStr := fmt.Sprintf("%.1fW", entry.Watts)
			statusStyled := lipgloss.NewStyle().Foreground(statusColor(entry.Status)).
				Width(statusW).Align(lipgloss.Left).Render(entry.Status)
			sb.WriteString(lipgloss.NewStyle().Foreground(white).Width(wattsW).Align(lipgloss.Left).Render(wattsStr))
			sb.WriteString("\u2502")
			sb.WriteString(statusStyled)
		} else {
			statusStyled := lipgloss.NewStyle().Foreground(statusColor(entry.Status)).
				Width(statusW).Align(lipgloss.Left).Render(entry.Status)
			sb.WriteString(statusStyled)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func tableColumnWidths(w int) (timeW, pctW, wattsW, statusW int) {
	if w >= 80 {
		return 10, 6, 8, 12
	}
	if w >= 50 {
		return 8, 5, 6, 10
	}
	if w >= 36 {
		return 8, 4, 6, 8
	}
	return 5, 4, 0, 8
}