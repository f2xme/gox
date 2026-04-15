package timex

import (
	"time"
)

// IsBefore 如果 t1 在 t2 之前则返回 true
func IsBefore(t1, t2 time.Time) bool {
	return t1.Before(t2)
}

// IsAfter 如果 t1 在 t2 之后则返回 true
func IsAfter(t1, t2 time.Time) bool {
	return t1.After(t2)
}

// IsBetween 如果 t 在 start 和 end 之间（包含边界）则返回 true
func IsBetween(t, start, end time.Time) bool {
	return (t.Equal(start) || t.After(start)) && (t.Equal(end) || t.Before(end))
}

// IsWeekend 如果 t 是周六或周日则返回 true
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// IsWorkday 如果 t 是周一到周五则返回 true
func IsWorkday(t time.Time) bool {
	return !IsWeekend(t)
}

// IsSameDay 如果 t1 和 t2 在同一天则返回 true
func IsSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// IsSameWeek 如果 t1 和 t2 在同一周则返回 true
func IsSameWeek(t1, t2 time.Time) bool {
	y1, w1 := t1.ISOWeek()
	y2, w2 := t2.ISOWeek()
	return y1 == y2 && w1 == w2
}

// IsSameMonth 如果 t1 和 t2 在同一月则返回 true
func IsSameMonth(t1, t2 time.Time) bool {
	y1, m1, _ := t1.Date()
	y2, m2, _ := t2.Date()
	return y1 == y2 && m1 == m2
}
