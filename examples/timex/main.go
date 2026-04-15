package main

import (
	"fmt"
	"time"

	"github.com/f2xme/gox/timex"
)

func main() {
	fmt.Println("=== timex 包使用示例 ===\n")

	// 示例 1: 时间格式化
	fmt.Println("示例 1: 时间格式化")
	now := time.Now()
	fmt.Printf("DateTime: %s\n", timex.Format(now, timex.DateTime))
	fmt.Printf("Date: %s\n", timex.Format(now, timex.Date))
	fmt.Printf("ISO8601: %s\n", timex.Format(now, timex.ISO8601))

	// 示例 2: 时间解析
	fmt.Println("\n示例 2: 时间解析")
	timeStr := "2024-01-15 14:30:00"
	parsed, err := timex.Parse(timeStr)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
	} else {
		fmt.Printf("解析成功: %s\n", parsed.Format(timex.DateTime))
	}

	// 示例 3: 时间范围计算
	fmt.Println("\n示例 3: 时间范围计算")
	start := timex.StartOfDay(now)
	end := timex.EndOfDay(now)
	fmt.Printf("今天开始: %s\n", start.Format(timex.DateTime))
	fmt.Printf("今天结束: %s\n", end.Format(timex.DateTime))

	weekStart := timex.StartOfWeek(now)
	weekEnd := timex.EndOfWeek(now)
	fmt.Printf("本周开始: %s\n", weekStart.Format(timex.DateTime))
	fmt.Printf("本周结束: %s\n", weekEnd.Format(timex.DateTime))

	monthStart := timex.StartOfMonth(now)
	monthEnd := timex.EndOfMonth(now)
	fmt.Printf("本月开始: %s\n", monthStart.Format(timex.DateTime))
	fmt.Printf("本月结束: %s\n", monthEnd.Format(timex.DateTime))

	// 示例 4: 时间比较
	fmt.Println("\n示例 4: 时间比较")
	today := time.Now()
	fmt.Printf("是否是同一天: %v\n", timex.IsSameDay(now, today))
	fmt.Printf("是否是工作日: %v\n", timex.IsWorkday(now))
	fmt.Printf("是否是周末: %v\n", timex.IsWeekend(now))

	// 示例 5: 人性化时间
	fmt.Println("\n示例 5: 人性化时间")
	past := now.Add(-5 * time.Minute)
	fmt.Printf("5分钟前: %s\n", timex.HumanDuration(now.Sub(past)))

	past2 := now.Add(-2 * time.Hour)
	fmt.Printf("2小时前: %s\n", timex.HumanDuration(now.Sub(past2)))

	past3 := now.Add(-3 * 24 * time.Hour)
	fmt.Printf("3天前: %s\n", timex.HumanDuration(now.Sub(past3)))

	// 示例 6: 工作日计算
	fmt.Println("\n示例 6: 工作日计算")
	futureDate := timex.AddWorkdays(now, 5)
	fmt.Printf("5个工作日后: %s\n", futureDate.Format(timex.Date))

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.Local)
	workdays := timex.WorkdaysBetween(startDate, endDate)
	fmt.Printf("2024年1月的工作日数: %d\n", workdays)

	// 示例 7: 时间间隔计算
	fmt.Println("\n示例 7: 时间间隔计算")
	t1 := time.Date(2024, 1, 1, 10, 0, 0, 0, time.Local)
	t2 := time.Date(2024, 1, 5, 14, 30, 0, 0, time.Local)
	fmt.Printf("相差天数: %d\n", timex.DaysBetween(t1, t2))
	fmt.Printf("相差小时数: %d\n", timex.HoursBetween(t1, t2))
	fmt.Printf("相差分钟数: %d\n", timex.MinutesBetween(t1, t2))

	// 示例 8: Unix 时间戳
	fmt.Println("\n示例 8: Unix 时间戳")
	timestamp := now.Unix()
	fmt.Printf("当前时间戳: %d\n", timestamp)
	fromUnix := timex.FromUnix(timestamp)
	fmt.Printf("从时间戳转换: %s\n", fromUnix.Format(timex.DateTime))

	milliTimestamp := now.UnixMilli()
	fmt.Printf("毫秒时间戳: %d\n", milliTimestamp)
	fromMilli := timex.FromUnixMilli(milliTimestamp)
	fmt.Printf("从毫秒时间戳转换: %s\n", fromMilli.Format(timex.DateTime))

	// 示例 9: 自定义时长解析
	fmt.Println("\n示例 9: 自定义时长解析")
	duration1, _ := timex.ParseDuration("2d")
	fmt.Printf("2天 = %v\n", duration1)

	duration2, _ := timex.ParseDuration("1w")
	fmt.Printf("1周 = %v\n", duration2)

	duration3, _ := timex.ParseDuration("3h30m")
	fmt.Printf("3小时30分钟 = %v\n", duration3)

	fmt.Println("\n=== 示例结束 ===")
}
