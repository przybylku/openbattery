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

// Real-world ioreg output shape: the nested BatteryData JSON blob contains
// keys like "Voltage"=..., "CycleCount"=... mid-line. The parser must only
// read the top-level `"Key" = Value` lines, not the embedded blob.
func TestParseIORegOutputIgnoresBatteryDataBlob(t *testing.T) {
	text := `+-o AppleSmartBattery  <class AppleSmartBattery, id 0x1000002c1>
    {
      "CurrentCapacity" = 94
      "Amperage" = 18446744073709551262
      "AppleRawCurrentCapacity" = 3633
      "BatteryData" = {"CycleCount"=999,"DesignCapacity"=1,"MaxCapacity"=100,"Voltage"=1,"Ra00"=324}
      "NominalChargeCapacity" = 4184
      "MaxCapacity" = 100
      "InstantAmperage" = 18446744073709551262
      "DesignCapacity" = 4563
      "Voltage" = 12700
      "CycleCount" = 158
      "AppleRawMaxCapacity" = 4057
    }`

	b := parseIORegOutput(text)

	if b.Voltage != 12700 {
		t.Errorf("Voltage = %d, want 12700 (not the BatteryData blob value)", b.Voltage)
	}
	if b.CycleCount != 158 {
		t.Errorf("CycleCount = %d, want 158 (not the BatteryData blob value)", b.CycleCount)
	}
	if b.DesignCapacity != 4563 {
		t.Errorf("DesignCapacity = %d, want 4563", b.DesignCapacity)
	}
	if b.Amperage != -354 {
		t.Errorf("Amperage = %d, want -354 (two's complement)", b.Amperage)
	}
	if b.CurrentCapacity != 3633 {
		t.Errorf("CurrentCapacity = %d, want 3633 (AppleRawCurrentCapacity)", b.CurrentCapacity)
	}
	if b.AppleRawMaxCapacity != 4057 {
		t.Errorf("AppleRawMaxCapacity = %d, want 4057", b.AppleRawMaxCapacity)
	}
}

func TestTimeToFullHours(t *testing.T) {
	b := IOReg{
		Amperage:            2000,
		CurrentCapacity:     3000,
		AppleRawMaxCapacity: 4000,
	}
	if got := b.TimeToFullHours(); got != 0.5 {
		t.Errorf("TimeToFullHours = %f, want 0.5", got)
	}

	full := IOReg{Amperage: 2000, CurrentCapacity: 4000, AppleRawMaxCapacity: 4000}
	if got := full.TimeToFullHours(); got != 0 {
		t.Errorf("TimeToFullHours when full = %f, want 0", got)
	}

	discharging := IOReg{Amperage: -500, CurrentCapacity: 3000, AppleRawMaxCapacity: 4000}
	if got := discharging.TimeToFullHours(); got != 0 {
		t.Errorf("TimeToFullHours while discharging = %f, want 0", got)
	}

	noData := IOReg{Amperage: 2000}
	if got := noData.TimeToFullHours(); got != 0 {
		t.Errorf("TimeToFullHours without capacity data = %f, want 0", got)
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