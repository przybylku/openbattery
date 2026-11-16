package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestBigDigitsWidth(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"93%", 11},
		{"100%", 15},
		{"7%", 7},
		{"", 0},
		{"abc", 0}, // unsupported runes are skipped
	}
	for _, tt := range tests {
		if got := BigDigitsWidth(tt.in); got != tt.want {
			t.Errorf("BigDigitsWidth(%q) = %d, want %d", tt.in, got, tt.want)
		}
		rendered := BigDigits(tt.in, green)
		lines := strings.Split(rendered, "\n")
		if len(lines) != 5 {
			t.Fatalf("BigDigits(%q) has %d lines, want 5", tt.in, len(lines))
		}
		for i, line := range lines {
			if w := lipgloss.Width(line); w != tt.want {
				t.Errorf("BigDigits(%q) line %d width = %d, want %d", tt.in, i, w, tt.want)
			}
		}
	}
}

func TestCardWidth(t *testing.T) {
	for _, w := range []int{8, 24, 40, 80} {
		c := card("title", "a\nlonger line\nx", w, cyan)
		for i, line := range strings.Split(c, "\n") {
			if got := lipgloss.Width(line); got != w {
				t.Errorf("card(width=%d) line %d width = %d, want %d", w, i, got, w)
			}
		}
	}
}

func TestColoredSparklineWidth(t *testing.T) {
	values := make([]float64, 100)
	for i := range values {
		values[i] = float64(i%100) / 100.0
	}
	for _, w := range []int{2, 10, 60} {
		got := ColoredSparkline(values, w)
		if rw := lipgloss.Width(got); rw != w {
			t.Errorf("ColoredSparkline width = %d, want %d", rw, w)
		}
	}
	if got := ColoredSparkline(nil, 10); lipgloss.Width(got) != 10 {
		t.Errorf("ColoredSparkline(empty) width = %d, want 10", lipgloss.Width(got))
	}
	if got := ColoredSparkline(values, 0); got != "" {
		t.Errorf("ColoredSparkline(width=0) = %q, want empty", got)
	}
}

func TestShortStatus(t *testing.T) {
	tests := map[string]string{
		"discharging": "dischg",
		"charging":    "chrg",
		"charged":     "full",
		"error":       "err",
	}
	for in, want := range tests {
		if got := shortStatus(in); got != want {
			t.Errorf("shortStatus(%q) = %q, want %q", in, got, want)
		}
		if len(shortStatus(in)) > 9 {
			t.Errorf("shortStatus(%q) too long for narrow column", in)
		}
	}
}
