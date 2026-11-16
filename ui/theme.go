package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Palette: soft semantic colors (open-color inspired) with adaptive
// neutrals so the UI stays legible on light and dark terminals alike.
var (
	green  = lipgloss.AdaptiveColor{Light: "#2B8A3E", Dark: "#69DB7C"}
	yellow = lipgloss.AdaptiveColor{Light: "#E67700", Dark: "#FFD43B"}
	red    = lipgloss.AdaptiveColor{Light: "#C92A2A", Dark: "#FF6B6B"}
	cyan   = lipgloss.AdaptiveColor{Light: "#1098AD", Dark: "#66D9EF"}
	white  = lipgloss.AdaptiveColor{Light: "#212529", Dark: "#E9ECEF"}
	gray   = lipgloss.AdaptiveColor{Light: "#868E96", Dark: "#8A9199"}
	dimmed = lipgloss.AdaptiveColor{Light: "#ADB5BD", Dark: "#5C636B"}
	border = lipgloss.AdaptiveColor{Light: "#CED4DA", Dark: "#3A4149"}
)

// card renders content inside a rounded box whose title is embedded in the
// top border: ╭─ title ────╮. Content is padded one space on each side and
// every line is normalized to the inner width.
func card(title, content string, width int, titleColor lipgloss.TerminalColor) string {
	if width < 8 {
		width = 8
	}
	inner := width - 4

	bs := lipgloss.NewStyle().Foreground(border)
	ts := lipgloss.NewStyle().Foreground(titleColor).Bold(true)

	titleW := lipgloss.Width(title)
	dashes := width - 5 - titleW
	var top string
	if dashes >= 1 {
		top = bs.Render("╭─ ") + ts.Render(title) + bs.Render(" "+strings.Repeat("─", dashes)+"╮")
	} else {
		top = bs.Render("╭" + strings.Repeat("─", width-2) + "╮")
	}

	trunc := lipgloss.NewStyle().MaxWidth(inner)
	var s strings.Builder
	s.WriteString(top)
	s.WriteString("\n")
	for _, line := range strings.Split(content, "\n") {
		l := trunc.Render(line)
		pad := inner - lipgloss.Width(l)
		if pad < 0 {
			pad = 0
		}
		s.WriteString(bs.Render("│") + " " + l + strings.Repeat(" ", pad) + " " + bs.Render("│"))
		s.WriteString("\n")
	}
	s.WriteString(bs.Render("╰" + strings.Repeat("─", width-2) + "╯"))
	return s.String()
}

// row renders a muted label followed by a bright value, aligned in columns.
func row(label string, labelW int, value string) string {
	ls := lipgloss.NewStyle().Foreground(gray).Width(labelW)
	return ls.Render(label) + value
}

// keyHint renders a keyboard shortcut: bold accent key, muted description.
func keyHint(key, desc string) string {
	k := lipgloss.NewStyle().Foreground(cyan).Bold(true).Render(key)
	d := lipgloss.NewStyle().Foreground(gray).Render(" " + desc)
	return k + d
}
