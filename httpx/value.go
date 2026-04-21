package httpx

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// Value 表示一个 HTTP 请求参数值（来自 URI 路径参数、Query、Header 等）。
//
// 其底层类型为 string，可直接用字符串字面量比较（如 v == ""），
// 也可以通过各类方法做类型转换、默认值回退和存在性判断。
//
// 约定：零值（空字符串）视为"未提供"。
type Value string

// String 返回原始字符串表示。
func (v Value) String() string { return string(v) }

// Exists 判断参数是否存在（非空字符串）。
func (v Value) Exists() bool { return v != "" }

// Or 参数不存在时返回 def，否则返回原始字符串。
//
// 示例：
//
//	name := c.Query("name").Or("guest")
func (v Value) Or(def string) string {
	if v == "" {
		return def
	}
	return string(v)
}

// Int 解析为 int。
//
// 返回值：
//   - int: 解析后的整数
//   - error: 解析失败时返回错误
func (v Value) Int() (int, error) { return strconv.Atoi(string(v)) }

// IntOr 解析为 int，失败或空值时返回 def。
//
// 示例：
//
//	page := c.Query("page").IntOr(1)
func (v Value) IntOr(def int) int {
	n, err := strconv.Atoi(string(v))
	if err != nil {
		return def
	}
	return n
}

// Int64 解析为 int64。
func (v Value) Int64() (int64, error) { return strconv.ParseInt(string(v), 10, 64) }

// Int64Or 解析为 int64，失败或空值时返回 def。
//
// 示例：
//
//	id := c.Param("id").Int64Or(0)
func (v Value) Int64Or(def int64) int64 {
	n, err := strconv.ParseInt(string(v), 10, 64)
	if err != nil {
		return def
	}
	return n
}

// Uint64 解析为 uint64。
func (v Value) Uint64() (uint64, error) { return strconv.ParseUint(string(v), 10, 64) }

// Uint64Or 解析为 uint64，失败或空值时返回 def。
func (v Value) Uint64Or(def uint64) uint64 {
	n, err := strconv.ParseUint(string(v), 10, 64)
	if err != nil {
		return def
	}
	return n
}

// Float64 解析为 float64。
func (v Value) Float64() (float64, error) { return strconv.ParseFloat(string(v), 64) }

// Float64Or 解析为 float64，失败或空值时返回 def。
func (v Value) Float64Or(def float64) float64 {
	f, err := strconv.ParseFloat(string(v), 64)
	if err != nil {
		return def
	}
	return f
}

// Bool 解析为 bool，支持 1/0/true/false/yes/no/on/off（不区分大小写）。
func (v Value) Bool() (bool, error) {
	switch strings.ToLower(string(v)) {
	case "1", "true", "yes", "on", "t":
		return true, nil
	case "0", "false", "no", "off", "f":
		return false, nil
	}
	return false, errors.New("httpx: invalid bool value: " + string(v))
}

// BoolOr 解析为 bool，失败或空值时返回 def。
//
// 示例：
//
//	enabled := c.Query("enabled").BoolOr(false)
func (v Value) BoolOr(def bool) bool {
	b, err := v.Bool()
	if err != nil {
		return def
	}
	return b
}

// Time 按指定 layout 解析为 time.Time。
//
// 示例：
//
//	t, err := c.Query("since").Time(time.RFC3339)
func (v Value) Time(layout string) (time.Time, error) {
	return time.Parse(layout, string(v))
}

// TimeOr 按指定 layout 解析为 time.Time，失败或空值时返回 def。
func (v Value) TimeOr(layout string, def time.Time) time.Time {
	t, err := time.Parse(layout, string(v))
	if err != nil {
		return def
	}
	return t
}

// Duration 解析为 time.Duration，如 "5s"、"1h30m"。
func (v Value) Duration() (time.Duration, error) {
	return time.ParseDuration(string(v))
}

// DurationOr 解析为 time.Duration，失败或空值时返回 def。
func (v Value) DurationOr(def time.Duration) time.Duration {
	d, err := time.ParseDuration(string(v))
	if err != nil {
		return def
	}
	return d
}

// Split 按分隔符切分为字符串切片；空值返回 nil。
//
// 示例：
//
//	tags := c.Query("tags").Split(",")  // "a,b,c" -> ["a","b","c"]
func (v Value) Split(sep string) []string {
	if v == "" {
		return nil
	}
	return strings.Split(string(v), sep)
}
