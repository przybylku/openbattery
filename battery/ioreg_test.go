package battery

import (
	"testing"
)

func TestFindIntField(t *testing.T) {
	text := `"Amperage" = -1103
"Voltage" = 10825
"CurrentCapacity" = 10
"AppleRawCurrentCapacity" = 323
"DesignCapacity" = 4563
"CycleCount" = 108
"NominalChargeCapacity" = 4148`

	tests := []struct {
		key  string
		want int64
	}{
		{"Amperage", -1103},
		{"Voltage", 10825},
		{"CurrentCapacity", 10},
		{"AppleRawCurrentCapacity", 323},
		{"DesignCapacity", 4563},
		{"CycleCount", 108},
		{"NominalChargeCapacity", 4148},
		{"NonExistentKey", 0},
	}

	for _, tt := range tests {
		got := findIntField(text, tt.key)
		if got != tt.want {
			t.Errorf("findIntField(%q) = %d, want %d", tt.key, got, tt.want)
		}
	}
}

func TestFindIntFieldUnsignedNegative(t *testing.T) {
	text := `"Amperage" = 18446744073709550510`

	got := findIntField(text, "Amperage")
	want := int64(-1106)

	if got != want {
		t.Errorf("findIntField unsigned negative = %d, want %d", got, want)
	}
}

func TestIORegComputed(t *testing.T) {
	b := IOReg{
		Amperage:            -661,
		Voltage:             10825,
		CurrentCapacity:     323,
		AppleRawMaxCapacity: 4148,
		DesignCapacity:      4563,
		CycleCount:          108,
	}

	watts := b.WattsNow()
	if watts < 7.1 || watts > 7.2 {
		t.Errorf("WattsNow = %f, want ~7.16", watts)
	}

	hours := b.EstimatedHours()
	if hours < 0.48 || hours > 0.50 {
		t.Errorf("EstimatedHours = %f, want ~0.489", hours)
	}

	health := b.BatteryHealth()
	if health < 90.0 || health > 91.0 {
		t.Errorf("BatteryHealth = %f, want ~90.9", health)
	}
}

func TestIORegBatteryHealthZeroDesign(t *testing.T) {
	b := IOReg{
		AppleRawMaxCapacity: 100,
		DesignCapacity:      0,
	}
	if got := b.BatteryHealth(); got != 0 {
		t.Errorf("BatteryHealth with zero DesignCapacity = %f, want 0", got)
	}
}

func TestIORegEstimatedHoursCharging(t *testing.T) {
	b := IOReg{
		Amperage:        661,
		CurrentCapacity: 323,
	}
	if got := b.EstimatedHours(); got != 0 {
		t.Errorf("EstimatedHours while charging = %f, want 0", got)
	}
}