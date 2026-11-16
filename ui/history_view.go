package ui

import (
	"fmt"
	"strings"

	"github.com/przybylku/openbattery/battery"

	"github.com/charmbracelet/lipgloss"
)

func historyView(m *model, w int) string {
	if len(m.history) == 0 {
		return card("history",
			lipgloss.NewStyle().Foreground(dimmed).Render("No samples yet — the first one lands within 10 seconds."),
			w, cyan)
	}

	percents := make([]float64, len(m.history))
	for i, entry := range m.history {
		percents[i] = entry.Percent
	}

	inner := w - 4
	sparkWidth := inner - 10
	if sparkWidth < 2 {
		sparkWidth = 2
	}

	hiLabel := lipgloss.NewStyle().Foreground(dimmed).Width(4).Align(lipgloss.Right).Render("100%")
	loLabel := lipgloss.NewStyle().Foreground(dimmed).Width(4).Align(lipgloss.Left).Render("0%")
	sparkLine := hiLabel + " " + ColoredSparkline(percents, sparkWidth) + " " + loLabel

	firstT := m.history[0].Time.Format("15:04")
	lastT := m.history[len(m.history)-1].Time.Format("15:04")
	rangeLine := lipgloss.NewStyle().Foreground(dimmed).Width(inner).Align(lipgloss.Center).
		Render(firstT + " — " + lastT)

	levelCard := card(fmt.Sprintf("level · %d samples", len(m.history)),
		sparkLine+"\n"+rangeLine, w, cyan)

	start := len(m.history) - 10
	if start < 0 {
		start = 0
	}
	recentCard := card("recent", historyTable(m.history[start:], inner), w, cyan)

	return levelCard + "\n" + recentCard
}

// historyTable renders a clean column table: muted header, hairline rule,
// right-aligned numbers, semantic colors. No heavy grid.
func historyTable(entries []battery.HistoryEntry, inner int) string {
	timeW, pctW, wattsW, statusW := tableColumnWidths(inner)

	header := lipgloss.NewStyle().Foreground(gray)
	muted := lipgloss.NewStyle().Foreground(dimmed)
	vs := lipgloss.NewStyle().Foreground(white)

	col := func(text string, width int, style lipgloss.Style, align lipgloss.Position) string {
		return style.Width(width).Align(align).Render(text)
	}

	var head, rule strings.Builder
	head.WriteString(col("time", timeW, header, lipgloss.Left))
	head.WriteString(col("%", pctW, header, lipgloss.Right))
	rule.WriteString(strings.Repeat("─", timeW+pctW))
	if wattsW > 0 {
		head.WriteString(col("watts", wattsW, header, lipgloss.Right))
		rule.WriteString(strings.Repeat("─", wattsW))
	}
	head.WriteString("  ")
	head.WriteString(col("status", statusW, header, lipgloss.Left))
	rule.WriteString("──")
	rule.WriteString(strings.Repeat("─", statusW))

	var sb strings.Builder
	sb.WriteString(head.String())
	sb.WriteString("\n")
	sb.WriteString(muted.Render(rule.String()))
	sb.WriteString("\n")

	for _, entry := range entries {
		timeStr := entry.Time.Format("15:04:05")
		if timeW < 8 {
			timeStr = entry.Time.Format("15:04")
		}
		status := entry.Status
		if statusW < 11 {
			status = shortStatus(status)
		}
		sb.WriteString(col(timeStr, timeW, muted, lipgloss.Left))
		sb.WriteString(col(fmt.Sprintf("%.0f%%", entry.Percent*100), pctW,
			lipgloss.NewStyle().Foreground(percentColor(entry.Percent)), lipgloss.Right))
		if wattsW > 0 {
			sb.WriteString(col(fmt.Sprintf("%.1f", entry.Watts), wattsW, vs, lipgloss.Right))
		}
		sb.WriteString("  ")
		sb.WriteString(col(status, statusW,
			lipgloss.NewStyle().Foreground(statusColor(entry.Status)), lipgloss.Left))
		sb.WriteString("\n")
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// shortStatus abbreviates statuses so narrow tables never wrap.
func shortStatus(status string) string {
	switch status {
	case "discharging":
		return "dischg"
	case "charging":
		return "chrg"
	case "charged":
		return "full"
	default:
		return "err"
	}
}

func tableColumnWidths(w int) (timeW, pctW, wattsW, statusW int) {
	switch {
	case w >= 60:
		return 12, 8, 10, 14
	case w >= 40:
		return 10, 6, 8, 12
	case w >= 28:
		return 8, 5, 0, 12
	default:
		return 6, 5, 0, 9
	}
}
