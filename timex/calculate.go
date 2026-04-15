package timex

import (
	"time"
)

// AddWorkdays 向给定时间添加指定数量的工作日
// 工作日是周一到周五，跳过周末
func AddWorkdays(t time.Time, days int) time.Time {
	if days == 0 {
		return t
	}

	direction := 1
	if days < 0 {
		direction = -1
		days = -days
	}

	result := t
	for i := 0; i < days; {
		result = result.AddDate(0, 0, direction)
		if IsWorkday(result) {
			i++
		}
	}

	return result
}

// WorkdaysBetween 返回两个时间之间的工作日数量
// 计算包含开始日期但不包含结束日期
func WorkdaysBetween(start, end time.Time) int {
	if end.Before(start) {
		return -WorkdaysBetween(end, start)
	}

	// Calculate total days
	totalDays := int(end.Sub(start).Hours() / 24)
	if totalDays == 0 {
		return 0
	}

	// Calculate full weeks and remaining days
	fullWeeks := totalDays / 7
	remainingDays := totalDays % 7

	// Each full week has 5 workdays
	workdays := fullWeeks * 5

	// Count workdays in remaining days
	current := start.AddDate(0, 0, fullWeeks*7)
	for range remainingDays {
		if IsWorkday(current) {
			workdays++
		}
		current = current.AddDate(0, 0, 1)
	}

	return workdays
}

// Today 返回今天的开始和结束时间
func Today() (time.Time, time.Time) {
	now := time.Now()
	start := StartOfDay(now)
	end := EndOfDay(now)
	return start, end
}

// ThisWeek 返回本周的开始和结束时间（周一到周日）
func ThisWeek() (time.Time, time.Time) {
	now := time.Now()
	start := StartOfWeek(now)
	end := EndOfWeek(now)
	return start, end
}

// ThisMonth 返回本月的开始和结束时间
func ThisMonth() (time.Time, time.Time) {
	now := time.Now()
	start := StartOfMonth(now)
	end := EndOfMonth(now)
	return start, end
}

// ThisYear 返回今年的开始和结束时间
func ThisYear() (time.Time, time.Time) {
	now := time.Now()
	start := StartOfDay(time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()))
	end := EndOfDay(time.Date(now.Year(), 12, 31, 0, 0, 0, 0, now.Location()))
	return start, end
}

// StartOfDay 返回一天的开始时间（00:00:00）
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay 返回一天的结束时间（23:59:59.999999999）
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek 返回一周的开始时间（周一 00:00:00）
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday
	}
	daysToMonday := weekday - 1
	monday := t.AddDate(0, 0, -daysToMonday)
	return StartOfDay(monday)
}

// EndOfWeek 返回一周的结束时间（周日 23:59:59.999999999）
func EndOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		return EndOfDay(t)
	}
	daysToSunday := 7 - weekday
	sunday := t.AddDate(0, 0, daysToSunday)
	return EndOfDay(sunday)
}

// StartOfMonth 返回一个月的开始时间（1号 00:00:00）
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth 返回一个月的结束时间（最后一天 23:59:59.999999999）
func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}
