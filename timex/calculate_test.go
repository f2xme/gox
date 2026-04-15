package timex

import (
	"testing"
	"time"
)

func TestAddWorkdays(t *testing.T) {
	// Monday
	monday := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	// Add 0 workdays -> same day
	result := AddWorkdays(monday, 0)
	if !result.Equal(monday) {
		t.Errorf("expected same time, got %v", result)
	}

	// Add 1 workday -> Tuesday
	result = AddWorkdays(monday, 1)
	if result.Weekday() != time.Tuesday {
		t.Errorf("expected Tuesday, got %v", result.Weekday())
	}

	// Add 5 workdays -> next Monday (skip weekend)
	result = AddWorkdays(monday, 5)
	if result.Weekday() != time.Monday {
		t.Errorf("expected Monday, got %v", result.Weekday())
	}
	if result.Day() != 22 {
		t.Errorf("expected day 22, got %d", result.Day())
	}

	// Subtract workdays (negative)
	result = AddWorkdays(monday, -1)
	if result.Weekday() != time.Friday {
		t.Errorf("expected Friday, got %v", result.Weekday())
	}
}

func TestWorkdaysBetween(t *testing.T) {
	// Monday to Friday (same week)
	monday := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	friday := time.Date(2024, 1, 19, 0, 0, 0, 0, time.UTC)

	days := WorkdaysBetween(monday, friday)
	if days != 4 {
		t.Errorf("expected 4 workdays, got %d", days)
	}

	// Monday to next Monday (includes weekend)
	nextMonday := time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC)
	days = WorkdaysBetween(monday, nextMonday)
	if days != 5 {
		t.Errorf("expected 5 workdays, got %d", days)
	}

	// Reversed order (negative result)
	days = WorkdaysBetween(friday, monday)
	if days != -4 {
		t.Errorf("expected -4 workdays, got %d", days)
	}

	// Same day
	days = WorkdaysBetween(monday, monday)
	if days != 0 {
		t.Errorf("expected 0 workdays, got %d", days)
	}
}

func TestToday(t *testing.T) {
	start, end := Today()

	now := time.Now()
	if start.Year() != now.Year() || start.Month() != now.Month() || start.Day() != now.Day() {
		t.Error("start date mismatch")
	}
	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Error("start should be at 00:00:00")
	}
	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Error("end should be at 23:59:59")
	}
}

func TestStartOfDay(t *testing.T) {
	tm := time.Date(2024, 1, 15, 14, 30, 45, 123456789, time.UTC)
	result := StartOfDay(tm)

	if result.Hour() != 0 || result.Minute() != 0 || result.Second() != 0 || result.Nanosecond() != 0 {
		t.Errorf("expected 00:00:00.000000000, got %02d:%02d:%02d.%09d",
			result.Hour(), result.Minute(), result.Second(), result.Nanosecond())
	}
}

func TestEndOfDay(t *testing.T) {
	tm := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	result := EndOfDay(tm)

	if result.Hour() != 23 || result.Minute() != 59 || result.Second() != 59 {
		t.Errorf("expected 23:59:59, got %02d:%02d:%02d",
			result.Hour(), result.Minute(), result.Second())
	}
}

func TestStartOfWeek(t *testing.T) {
	// Wednesday
	wed := time.Date(2024, 1, 17, 14, 30, 0, 0, time.UTC)
	result := StartOfWeek(wed)

	// Should be Monday
	if result.Weekday() != time.Monday {
		t.Errorf("expected Monday, got %v", result.Weekday())
	}
	if result.Day() != 15 {
		t.Errorf("expected day 15, got %d", result.Day())
	}

	// Sunday edge case
	sunday := time.Date(2024, 1, 21, 14, 30, 0, 0, time.UTC)
	result = StartOfWeek(sunday)
	if result.Weekday() != time.Monday {
		t.Errorf("expected Monday, got %v", result.Weekday())
	}
	if result.Day() != 15 {
		t.Errorf("expected day 15, got %d", result.Day())
	}
}

func TestStartOfMonth(t *testing.T) {
	tm := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	result := StartOfMonth(tm)

	if result.Day() != 1 {
		t.Errorf("expected day 1, got %d", result.Day())
	}
	if result.Hour() != 0 {
		t.Errorf("expected hour 0, got %d", result.Hour())
	}
}

func TestThisWeek(t *testing.T) {
	start, end := ThisWeek()

	if start.Weekday() != time.Monday {
		t.Errorf("expected start to be Monday, got %v", start.Weekday())
	}
	if end.Weekday() != time.Sunday {
		t.Errorf("expected end to be Sunday, got %v", end.Weekday())
	}
	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Error("start should be at 00:00:00")
	}
	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Error("end should be at 23:59:59")
	}
}

func TestThisMonth(t *testing.T) {
	start, end := ThisMonth()

	now := time.Now()
	if start.Year() != now.Year() || start.Month() != now.Month() {
		t.Error("start month mismatch")
	}
	if start.Day() != 1 {
		t.Errorf("expected start day 1, got %d", start.Day())
	}
	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Error("start should be at 00:00:00")
	}
	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Error("end should be at 23:59:59")
	}
}

func TestThisYear(t *testing.T) {
	start, end := ThisYear()

	now := time.Now()
	if start.Year() != now.Year() {
		t.Error("start year mismatch")
	}
	if start.Month() != time.January || start.Day() != 1 {
		t.Error("start should be January 1st")
	}
	if end.Month() != time.December || end.Day() != 31 {
		t.Error("end should be December 31st")
	}
	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Error("start should be at 00:00:00")
	}
	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Error("end should be at 23:59:59")
	}
}

func TestEndOfWeek(t *testing.T) {
	// Wednesday
	wed := time.Date(2024, 1, 17, 14, 30, 0, 0, time.UTC)
	result := EndOfWeek(wed)

	if result.Weekday() != time.Sunday {
		t.Errorf("expected Sunday, got %v", result.Weekday())
	}
	if result.Day() != 21 {
		t.Errorf("expected day 21, got %d", result.Day())
	}
	if result.Hour() != 23 || result.Minute() != 59 || result.Second() != 59 {
		t.Error("end should be at 23:59:59")
	}

	// Sunday edge case
	sunday := time.Date(2024, 1, 21, 14, 30, 0, 0, time.UTC)
	result = EndOfWeek(sunday)
	if result.Weekday() != time.Sunday {
		t.Errorf("expected Sunday, got %v", result.Weekday())
	}
	if result.Day() != 21 {
		t.Errorf("expected day 21, got %d", result.Day())
	}
}

func TestEndOfMonth(t *testing.T) {
	// January (31 days)
	jan := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	result := EndOfMonth(jan)

	if result.Day() != 31 {
		t.Errorf("expected day 31, got %d", result.Day())
	}
	if result.Hour() != 23 || result.Minute() != 59 || result.Second() != 59 {
		t.Error("end should be at 23:59:59")
	}

	// February (29 days in 2024, leap year)
	feb := time.Date(2024, 2, 15, 14, 30, 0, 0, time.UTC)
	result = EndOfMonth(feb)
	if result.Day() != 29 {
		t.Errorf("expected day 29, got %d", result.Day())
	}
}
