package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// sparkValues compresses values down to at most width samples, evenly
// spaced, always including the first and last sample.
func sparkValues(values []float64, width int) []float64 {
	if len(values) > width && width > 1 {
		step := float64(len(values)-1) / float64(width-1)
		compressed := make([]float64, width)
		for i := 0; i < width; i++ {
			idx := int(float64(i) * step)
			if idx >= len(values) {
				idx = len(values) - 1
			}
			compressed[i] = values[idx]
		}
		return compressed
	}
	if len(values) > width {
		// width == 1: just take the last value
		return []float64{values[len(values)-1]}
	}
	return values
}

func sparkRune(v float64) rune {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	idx := int(v * 7)
	if idx > 7 {
		idx = 7
	}
	return sparkChars[idx]
}

// Sparkline renders a list of values (0.0–1.0) as an ASCII sparkline bar
// using block elements ▁▂▃▄▅▆▇█. Returns a string of exactly width characters.
// Values beyond 0–1 are clamped. If width <= 0, returns an empty string.
// If values is empty, returns a bar of the lowest character.
func Sparkline(values []float64, width int) string {
	if width <= 0 {
		return ""
	}
	if len(values) == 0 {
		return strings.Repeat(string(sparkChars[0]), width)
	}

	values = sparkValues(values, width)

	var sb strings.Builder
	sb.Grow(width * 4)
	for i := len(values); i < width; i++ {
		sb.WriteRune(' ')
	}
	for _, v := range values {
		sb.WriteRune(sparkRune(v))
	}
	return sb.String()
}

// ColoredSparkline renders like Sparkline but colors each block by its
// battery level (red / yellow / green), so low-charge periods stand out.
func ColoredSparkline(values []float64, width int) string {
	if width <= 0 {
		return ""
	}
	if len(values) == 0 {
		return lipgloss.NewStyle().Foreground(dimmed).
			Render(strings.Repeat(string(sparkChars[0]), width))
	}

	values = sparkValues(values, width)

	var sb strings.Builder
	for i := len(values); i < width; i++ {
		sb.WriteRune(' ')
	}
	for _, v := range values {
		sb.WriteString(lipgloss.NewStyle().Foreground(percentColor(v)).Render(string(sparkRune(v))))
	}
	return sb.String()
}
