package battery

import (
	"testing"
)

func TestParsePmsetOutput(t *testing.T) {
	tests := []struct {
		name        string
		out         string
		wantPercent float64
		wantStatus  string
		wantTime    string
	}{
		{
			name:        "discharging with estimate",
			out:         "Now drawing from 'Battery Power'\n -InternalBattery-0 (id=22347875)\t94%; discharging; 9:53 remaining present: true\n",
			wantPercent: 0.94,
			wantStatus:  "discharging",
			wantTime:    "9h 53min",
		},
		{
			name:        "charging no estimate",
			out:         "Now drawing from 'AC Power'\n -InternalBattery-0 (id=22347875)\t80%; charging; (no estimate) present: true\n",
			wantPercent: 0.80,
			wantStatus:  "charging",
			wantTime:    "\u2014",
		},
		{
			name:        "charged",
			out:         "Now drawing from 'AC Power'\n -InternalBattery-0 (id=22347875)\t100%; charged; 0:00 remaining present: true\n",
			wantPercent: 1.0,
			wantStatus:  "charged",
			wantTime:    "0min",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parsePmsetOutput(tt.out)
			if err != nil {
				t.Fatalf("parsePmsetOutput error: %v", err)
			}
			if p.Percent != tt.wantPercent {
				t.Errorf("Percent = %f, want %f", p.Percent, tt.wantPercent)
			}
			if p.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", p.Status, tt.wantStatus)
			}
			if p.RemainingTime != tt.wantTime {
				t.Errorf("RemainingTime = %q, want %q", p.RemainingTime, tt.wantTime)
			}
		})
	}

	if _, err := parsePmsetOutput("garbage output"); err == nil {
		t.Error("parsePmsetOutput(garbage) should return an error")
	}
}

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
