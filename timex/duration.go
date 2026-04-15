package timex

import (
	"fmt"
	"strconv"
	"time"
)

const (
	day   = 24 * time.Hour
	week  = 7 * day
	month = 30 * day
	year  = 365 * day
)

// pluralize formats a count with singular or plural unit name.
func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", singular)
	}
	return fmt.Sprintf("%d %s", n, plural)
}

// HumanDuration 返回人类可读的时长字符串
func HumanDuration(d time.Duration) string {
	if d < time.Minute {
		return pluralize(int(d.Seconds()), "second", "seconds")
	}
	if d < time.Hour {
		return pluralize(int(d.Minutes()), "minute", "minutes")
	}
	if d < day {
		return pluralize(int(d.Hours()), "hour", "hours")
	}
	if d < week {
		return pluralize(int(d/day), "day", "days")
	}
	if d < month {
		return pluralize(int(d/week), "week", "weeks")
	}
	if d < year {
		return pluralize(int(d/month), "month", "months")
	}
	return pluralize(int(d/year), "year", "years")
}

// ParseDuration 解析支持天（d）和周（w）的时长字符串
// 它扩展了 time.ParseDuration 以支持额外的单位
func ParseDuration(s string) (time.Duration, error) {
	// Try standard parsing first
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	// Handle custom units: d (days), w (weeks)
	var total time.Duration
	remaining := s

	for remaining != "" {
		// Find the next number
		i := 0
		for i < len(remaining) && (remaining[i] >= '0' && remaining[i] <= '9' || remaining[i] == '.') {
			i++
		}
		if i == 0 {
			return 0, fmt.Errorf("invalid duration: %s", s)
		}

		numStr := remaining[:i]
		num, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid duration: %s", s)
		}

		if i >= len(remaining) {
			return 0, fmt.Errorf("missing unit in duration: %s", s)
		}

		unit := remaining[i]
		var duration time.Duration

		switch unit {
		case 'd':
			duration = time.Duration(num * float64(day))
		case 'w':
			duration = time.Duration(num * float64(week))
		default:
			// Try to parse the rest with standard parser
			if d, err := time.ParseDuration(remaining); err == nil {
				return total + d, nil
			}
			return 0, fmt.Errorf("unknown unit in duration: %c", unit)
		}

		total += duration
		remaining = remaining[i+1:]
	}

	return total, nil
}

// DaysBetween 返回两个时间之间的天数
func DaysBetween(t1, t2 time.Time) int {
	duration := t2.Sub(t1)
	return int(duration.Hours() / 24)
}

// HoursBetween 返回两个时间之间的小时数
func HoursBetween(t1, t2 time.Time) int {
	duration := t2.Sub(t1)
	return int(duration.Hours())
}

// MinutesBetween 返回两个时间之间的分钟数
func MinutesBetween(t1, t2 time.Time) int {
	duration := t2.Sub(t1)
	return int(duration.Minutes())
}
