package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// bigFont is a 3×5 block font for the oversized battery-percentage gauge.
var bigFont = map[rune][5]string{
	'0': {"███", "█ █", "█ █", "█ █", "███"},
	'1': {" █ ", "██ ", " █ ", " █ ", "███"},
	'2': {"███", "  █", "███", "█  ", "███"},
	'3': {"███", "  █", "███", "  █", "███"},
	'4': {"█ █", "█ █", "███", "  █", "  █"},
	'5': {"███", "█  ", "███", "  █", "███"},
	'6': {"███", "█  ", "███", "█ █", "███"},
	'7': {"███", "  █", "  █", "  █", "  █"},
	'8': {"███", "█ █", "███", "█ █", "███"},
	'9': {"███", "█ █", "███", "  █", "███"},
	'%': {"█ █", "  █", " █ ", "█  ", "█ █"},
}

// BigDigits renders digits and '%' as a five-line block gauge, one space
// between glyphs, in a single color. Unknown runes are skipped.
func BigDigits(s string, color lipgloss.TerminalColor) string {
	var rows [5]strings.Builder
	first := true
	for _, r := range s {
		glyph, ok := bigFont[r]
		if !ok {
			continue
		}
		for i := 0; i < 5; i++ {
			if !first {
				rows[i].WriteString(" ")
			}
			rows[i].WriteString(glyph[i])
		}
		first = false
	}
	style := lipgloss.NewStyle().Foreground(color)
	lines := make([]string, 5)
	for i := range rows {
		lines[i] = style.Render(rows[i].String())
	}
	return strings.Join(lines, "\n")
}

// BigDigitsWidth returns the rendered width of BigDigits for the same input.
func BigDigitsWidth(s string) int {
	n := 0
	for _, r := range s {
		if _, ok := bigFont[r]; ok {
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return n*4 - 1
}
