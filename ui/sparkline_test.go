package ui

import (
	"strings"
	"testing"
)

func TestSparklineEmpty(t *testing.T) {
	got := Sparkline(nil, 10)
	if len([]rune(got)) != 10 {
		t.Errorf("Sparkline(empty, 10) length = %d, want 10", len([]rune(got)))
	}
	if !strings.Contains(got, string(sparkChars[0])) {
		t.Errorf("Sparkline(empty) should contain lowest char")
	}
}

func TestSparklineWidthExact(t *testing.T) {
	values := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
	got := Sparkline(values, 5)
	if len([]rune(got)) != 5 {
		t.Errorf("Sparkline width = %d, want 5", len([]rune(got)))
	}
}

func TestSparklineCompress(t *testing.T) {
	values := make([]float64, 100)
	for i := range values {
		values[i] = float64(i) / 100.0
	}
	got := Sparkline(values, 10)
	if len([]rune(got)) != 10 {
		t.Errorf("Sparkline compressed width = %d, want 10", len([]rune(got)))
	}
}

func TestSparklinePad(t *testing.T) {
	values := []float64{0.5, 0.8}
	got := Sparkline(values, 10)
	runes := []rune(got)
	if len(runes) != 10 {
		t.Errorf("Sparkline padded width = %d, want 10", len(runes))
	}
	paddedCount := 0
	for _, r := range runes {
		if r == ' ' {
			paddedCount++
		}
	}
	if paddedCount != 8 {
		t.Errorf("Sparkline padding = %d spaces, want 8", paddedCount)
	}
}

func TestSparklineClamp(t *testing.T) {
	values := []float64{-0.5, 0.5, 1.5}
	got := Sparkline(values, 3)
	runes := []rune(got)
	if runes[0] != sparkChars[0] {
		t.Errorf("clamped low: got %c, want %c", runes[0], sparkChars[0])
	}
	if runes[2] != sparkChars[7] {
		t.Errorf("clamped high: got %c, want %c", runes[2], sparkChars[7])
	}
}
