// Package timex 提供便捷的时间操作工具函数。
//
// timex 包提供一组简单易用的时间处理函数，涵盖格式化、解析、时区转换、时间计算等常用操作。
// 所有函数都是线程安全的，可以在多个 goroutine 中并发使用。
//
// # 功能特性
//
//   - 时间格式化和解析：支持常用时间格式的快速转换
//   - 时区转换：在不同时区之间转换时间
//   - 时间计算：计算工作日、时间范围、时间段的开始和结束
//   - 时间比较：判断时间关系、是否同一天/周/月、是否工作日/周末
//   - 人性化时长：将时长转换为易读的字符串格式
//   - 扩展的时长解析：支持天（d）和周（w）单位
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import (
//		"fmt"
//		"time"
//
//		"github.com/f2xme/gox/timex"
//	)
//
//	func main() {
//		// 格式化时间
//		now := time.Now()
//		str := timex.Format(now, timex.DateTime)
//		fmt.Println(str) // "2024-01-01 12:00:00"
//
//		// 解析时间字符串
//		t, err := timex.Parse("2024-01-01 12:00:00")
//		if err != nil {
//			panic(err)
//		}
//
//		// 时区转换
//		utc := timex.ToUTC(now)
//		ny, _ := timex.ToTimezone(now, "America/New_York")
//
//		// 时间计算
//		start := timex.StartOfDay(now)   // 今天 00:00:00
//		end := timex.EndOfDay(now)       // 今天 23:59:59
//
//		// 时间比较
//		isToday := timex.IsSameDay(t, now)
//		isWorkday := timex.IsWorkday(now)
//
//		// 人性化时长
//		duration := timex.HumanDuration(5 * time.Minute)
//		fmt.Println(duration) // "5 minutes"
//	}
//
// # 常用时间格式
//
// 包定义了常用的时间格式常量：
//
//	DateTime     = "2006-01-02 15:04:05"
//	Date         = "2006-01-02"
//	Time         = "15:04:05"
//	DateTimeZone = "2006-01-02 15:04:05 -0700"
//	ISO8601      = "2006-01-02T15:04:05Z07:00"
//	RFC3339Milli = "2006-01-02T15:04:05.000Z07:00"
//
// # 格式化和解析
//
// 使用预定义格式：
//
//	// 格式化
//	str := timex.Format(time.Now(), timex.DateTime)
//	// "2024-01-01 12:00:00"
//
//	// 解析（自动尝试多种格式）
//	t, err := timex.Parse("2024-01-01 12:00:00")
//
//	// Unix 时间戳
//	t := timex.FromUnix(1704067200)
//	t := timex.FromUnixMilli(1704067200000)
//	str := timex.FormatUnix(1704067200, timex.DateTime)
//
// # 时区转换
//
// 在不同时区之间转换：
//
//	// 转换到指定时区
//	t, err := timex.ToTimezone(time.Now(), "America/New_York")
//
//	// 转换到 UTC
//	utc := timex.ToUTC(time.Now())
//
//	// 转换到本地时区
//	local := timex.ToLocal(utcTime)
//
// 常用时区：
//
//	"Asia/Shanghai"       // 中国
//	"America/New_York"    // 美国东部
//	"America/Los_Angeles" // 美国西部
//	"Europe/London"       // 英国
//	"Asia/Tokyo"          // 日本
//
// # 时间计算
//
// 获取时间段的开始和结束：
//
//	// 今天
//	start := timex.StartOfDay(time.Now())   // 2024-01-01 00:00:00
//	end := timex.EndOfDay(time.Now())       // 2024-01-01 23:59:59
//	start, end := timex.Today()
//
//	// 本周（周一到周日）
//	start := timex.StartOfWeek(time.Now())
//	end := timex.EndOfWeek(time.Now())
//	start, end := timex.ThisWeek()
//
//	// 本月
//	start := timex.StartOfMonth(time.Now())
//	end := timex.EndOfMonth(time.Now())
//	start, end := timex.ThisMonth()
//
//	// 今年
//	start, end := timex.ThisYear()
//
// 工作日计算：
//
//	// 添加工作日
//	future := timex.AddWorkdays(time.Now(), 5)  // 5 个工作日后
//	past := timex.AddWorkdays(time.Now(), -3)   // 3 个工作日前
//
//	// 计算工作日数量
//	count := timex.WorkdaysBetween(start, end)
//
// 时间间隔计算：
//
//	days := timex.DaysBetween(t1, t2)
//	hours := timex.HoursBetween(t1, t2)
//	minutes := timex.MinutesBetween(t1, t2)
//
// # 时间比较
//
// 基本比较：
//
//	// 前后关系
//	isBefore := timex.IsBefore(t1, t2)
//	isAfter := timex.IsAfter(t1, t2)
//	isBetween := timex.IsBetween(t, start, end)
//
//	// 同一时间段
//	sameDay := timex.IsSameDay(t1, t2)
//	sameWeek := timex.IsSameWeek(t1, t2)
//	sameMonth := timex.IsSameMonth(t1, t2)
//
//	// 工作日/周末
//	isWorkday := timex.IsWorkday(t)
//	isWeekend := timex.IsWeekend(t)
//
// # 人性化时长
//
// 将时长转换为易读格式：
//
//	duration := timex.HumanDuration(5 * time.Minute)
//	// "5 minutes"
//
//	duration := timex.HumanDuration(2 * time.Hour)
//	// "2 hours"
//
//	duration := timex.HumanDuration(3 * 24 * time.Hour)
//	// "3 days"
//
// 解析扩展的时长格式：
//
//	d, err := timex.ParseDuration("2d")      // 2 天
//	d, err := timex.ParseDuration("1w")      // 1 周
//	d, err := timex.ParseDuration("1w2d3h")  // 1 周 2 天 3 小时
//
// # 最佳实践
//
// 1. 使用预定义格式常量
//
//	// 推荐：使用常量
//	str := timex.Format(t, timex.DateTime)
//
//	// 不推荐：硬编码格式
//	str := t.Format("2006-01-02 15:04:05")
//
// 2. 统一使用 UTC 存储时间
//
//	// 存储时转换为 UTC
//	utc := timex.ToUTC(time.Now())
//	db.Save(utc)
//
//	// 显示时转换为本地时区
//	local := timex.ToLocal(utc)
//	fmt.Println(timex.Format(local, timex.DateTime))
//
// 3. 使用时间范围函数
//
//	// 推荐：使用工具函数
//	start := timex.StartOfDay(time.Now())
//	end := timex.EndOfDay(time.Now())
//
//	// 不推荐：手动计算
//	now := time.Now()
//	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
//
// 4. 处理时区
//
//	// 明确指定时区
//	t, _ := timex.ToTimezone(time.Now(), "Asia/Shanghai")
//
//	// 避免使用固定偏移量
//	// 不推荐：time.FixedZone("CST", 8*3600)
//
// # 性能考虑
//
//   - 时间格式化和解析是相对昂贵的操作
//   - 如果需要频繁格式化，考虑缓存结果
//   - 时区转换涉及时区数据库查询，避免在循环中重复加载时区
package timex
