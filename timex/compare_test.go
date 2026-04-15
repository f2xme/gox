package timex

import (
	"testing"
	"time"
)

func TestIsBefore(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC)

	if !IsBefore(t1, t2) {
		t.Error("expected t1 to be before t2")
	}
	if IsBefore(t2, t1) {
		t.Error("expected t2 not to be before t1")
	}
}

func TestIsAfter(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC)

	if !IsAfter(t2, t1) {
		t.Error("expected t2 to be after t1")
	}
	if IsAfter(t1, t2) {
		t.Error("expected t1 not to be after t2")
	}
}

func TestIsBetween(t *testing.T) {
	start := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
	middle := time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC)
	before := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	after := time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC)

	if !IsBetween(middle, start, end) {
		t.Error("expected middle to be between start and end")
	}
	if IsBetween(before, start, end) {
		t.Error("expected before not to be between start and end")
	}
	if IsBetween(after, start, end) {
		t.Error("expected after not to be between start and end")
	}
}

func TestIsWeekend(t *testing.T) {
	saturday := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)
	sunday := time.Date(2024, 1, 21, 10, 0, 0, 0, time.UTC)
	monday := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	if !IsWeekend(saturday) {
		t.Error("expected Saturday to be weekend")
	}
	if !IsWeekend(sunday) {
		t.Error("expected Sunday to be weekend")
	}
	if IsWeekend(monday) {
		t.Error("expected Monday not to be weekend")
	}
}

func TestIsWorkday(t *testing.T) {
	monday := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	saturday := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)

	if !IsWorkday(monday) {
		t.Error("expected Monday to be workday")
	}
	if IsWorkday(saturday) {
		t.Error("expected Saturday not to be workday")
	}
}

func TestIsSameDay(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	t3 := time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC)

	if !IsSameDay(t1, t2) {
		t.Error("expected t1 and t2 to be same day")
	}
	if IsSameDay(t1, t3) {
		t.Error("expected t1 and t3 not to be same day")
	}
}

func TestIsSameWeek(t *testing.T) {
	// Monday and Friday of same week
	monday := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	friday := time.Date(2024, 1, 19, 10, 0, 0, 0, time.UTC)
	nextMonday := time.Date(2024, 1, 22, 10, 0, 0, 0, time.UTC)

	if !IsSameWeek(monday, friday) {
		t.Error("expected monday and friday to be same week")
	}
	if IsSameWeek(monday, nextMonday) {
		t.Error("expected monday and nextMonday not to be same week")
	}
}

func TestIsSameMonth(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 25, 10, 0, 0, 0, time.UTC)
	t3 := time.Date(2024, 2, 15, 10, 0, 0, 0, time.UTC)

	if !IsSameMonth(t1, t2) {
		t.Error("expected t1 and t2 to be same month")
	}
	if IsSameMonth(t1, t3) {
		t.Error("expected t1 and t3 not to be same month")
	}
}
