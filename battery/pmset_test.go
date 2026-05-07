package battery

import (
	"testing"
)

func TestNormalizeStatus(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{"discharging", "discharging"},
		{"Discharging", "discharging"},
		{"  DISCHARGING  ", "discharging"},
		{"charging", "charging"},
		{"finishing charge", "charging"},
		{"charged", "charged"},
		{"AC attached", "charged"},
		{"ac attached", "charged"},
		{"unknown", "error"},
		{"", "error"},
	}

	for _, tt := range tests {
		got := normalizeStatus(tt.raw)
		if got != tt.want {
			t.Errorf("normalizeStatus(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}

func TestFormatPmsetTime(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{"4:32 remaining", "4h 32min"},
		{"0:45 remaining", "45min"},
		{"0:05", "5min"},
		{"", "\u2014"},
		{" (no estimate)", "\u2014"},
		{"no estimate available", "\u2014"},
		{"some other text", "some other text"},
	}

	for _, tt := range tests {
		got := formatPmsetTime(tt.raw)
		if got != tt.want {
			t.Errorf("formatPmsetTime(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}
