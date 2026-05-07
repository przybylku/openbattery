package ui

import "strings"

var sparkChars = []rune{'\u2581', '\u2582', '\u2583', '\u2584', '\u2585', '\u2586', '\u2587', '\u2588'}

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
		values = compressed
	} else if len(values) > width {
		// width == 1: just take the last value
		values = []float64{values[len(values)-1]}
	}

	var sb strings.Builder
	sb.Grow(width * 4)

	padLeft := width - len(values)
	for i := 0; i < padLeft; i++ {
		sb.WriteRune(' ')
	}

	for _, v := range values {
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
		sb.WriteRune(sparkChars[idx])
	}

	return sb.String()
}