package timex

import (
	"time"
)

// Format 使用给定的布局格式化时间
func Format(t time.Time, layout string) string {
	return t.Format(layout)
}

// FormatUnix 使用给定的布局格式化 Unix 时间戳
func FormatUnix(unix int64, layout string) string {
	return time.Unix(unix, 0).UTC().Format(layout)
}

// 常见时间格式列表，用于 Parse 函数
var commonFormats = []string{
	DateTime,
	Date,
	time.RFC3339,
	time.RFC3339Nano,
	ISO8601,
	RFC3339Milli,
	DateTimeZone,
	time.RFC1123,
	time.RFC1123Z,
}

// Parse 尝试使用常见格式解析时间字符串
// 它会尝试多种格式并返回第一个成功解析的结果
func Parse(s string) (time.Time, error) {
	formats := commonFormats

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, &time.ParseError{
		Layout:  "multiple formats",
		Value:   s,
		Message: "no matching format found",
	}
}

// FromUnix 将 Unix 时间戳（秒）转换为 time.Time
func FromUnix(sec int64) time.Time {
	return time.Unix(sec, 0)
}

// FromUnixMilli 将 Unix 时间戳（毫秒）转换为 time.Time
func FromUnixMilli(msec int64) time.Time {
	return time.UnixMilli(msec)
}
