package timex

import (
	"testing"
	"time"
)

func TestToTimezone(t *testing.T) {
	utc := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)

	shanghai, err := ToTimezone(utc, "Asia/Shanghai")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Shanghai is UTC+8
	if shanghai.Hour() != 22 {
		t.Errorf("expected hour 22, got %d", shanghai.Hour())
	}

	// Test error case
	_, err = ToTimezone(utc, "Invalid/Timezone")
	if err == nil {
		t.Error("ToTimezone(invalid): expected error, got nil")
	}
}

func TestToUTC(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	shanghai := time.Date(2024, 1, 15, 22, 30, 0, 0, loc)

	utc := ToUTC(shanghai)

	if utc.Hour() != 14 {
		t.Errorf("expected hour 14, got %d", utc.Hour())
	}
	if utc.Location() != time.UTC {
		t.Error("expected UTC location")
	}
}

func TestToLocal(t *testing.T) {
	utc := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	local := ToLocal(utc)

	if local.Location() != time.Local {
		t.Error("expected Local location")
	}
}
