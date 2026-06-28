package nullx

import (
	"database/sql"
	"encoding/json"
	"time"
)

// NullableString 是 sql.NullString 的别名。
type NullableString = sql.NullString

// NullableInt64 是 sql.NullInt64 的别名。
type NullableInt64 = sql.NullInt64

// NullableTime 是 sql.NullTime 的别名。
type NullableTime = sql.NullTime

// NullableBool 是 sql.NullBool 的别名。
type NullableBool = sql.NullBool

// String 将字符串转换为 sql.NullString，空字符串会转换为 NULL。
func String(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

// NullString 将字符串转换为 sql.NullString，空字符串会转换为 NULL。
func NullString(value string) sql.NullString {
	return String(value)
}

// StringFromPtr 将字符串指针转换为 sql.NullString，nil 会转换为 NULL。
func StringFromPtr(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

// NullStringFromPtr 将字符串指针转换为 sql.NullString，nil 会转换为 NULL。
func NullStringFromPtr(value *string) sql.NullString {
	return StringFromPtr(value)
}

// StringNull 将字符串指针转换为 sql.NullString，nil 会转换为 NULL。
func StringNull(value *string) sql.NullString {
	return StringFromPtr(value)
}

// StringValue 从 sql.NullString 取出字符串，NULL 返回空字符串。
func StringValue(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

// NullStringValue 从 sql.NullString 取出字符串，NULL 返回空字符串。
func NullStringValue(value sql.NullString) string {
	return StringValue(value)
}

// PtrStringValue 从字符串指针取出字符串，nil 返回空字符串。
func PtrStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// StringPtr 将 sql.NullString 转换为字符串指针，NULL 返回 nil。
func StringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

// PtrNullString 将 sql.NullString 转换为字符串指针，NULL 返回 nil。
func PtrNullString(value sql.NullString) *string {
	return StringPtr(value)
}

// PtrString 将字符串转换为字符串指针，空字符串返回 nil。
func PtrString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// PatchNullString 按可选字符串补丁更新 sql.NullString，nil 保持 current。
func PatchNullString(current sql.NullString, value *string) sql.NullString {
	if value == nil {
		return current
	}
	return String(*value)
}

// PatchStringPtrAsNull 按可选字符串补丁更新 sql.NullString，nil 时使用 current。
func PatchStringPtrAsNull(current *string, value *string) sql.NullString {
	if value != nil {
		return String(*value)
	}
	return StringFromPtr(current)
}

// StringPtrChanged 判断可选字符串指针补丁是否改变当前值，nil 补丁表示不变。
func StringPtrChanged(current *string, value *string) bool {
	if value == nil {
		return false
	}
	if current == nil {
		return *value != ""
	}
	return *current != *value
}

// StringChanged 判断字符串值是否改变当前可选字符串。
func StringChanged(current *string, value string) bool {
	if current == nil {
		return value != ""
	}
	return *current != value
}

// StringNullChanged 判断 sql.NullString 是否改变当前可选字符串。
func StringNullChanged(current *string, value sql.NullString) bool {
	if current == nil {
		return value.Valid
	}
	return !value.Valid || *current != value.String
}

// Bool 将布尔值转换为 sql.NullBool，false 也是有效值。
func Bool(value bool) sql.NullBool {
	return sql.NullBool{Bool: value, Valid: true}
}

// Int64 将 int64 转换为 sql.NullInt64，非正整数会转换为 NULL。
func Int64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value > 0}
}

// NullInt64 将 int64 转换为 sql.NullInt64，非正整数会转换为 NULL。
func NullInt64(value int64) sql.NullInt64 {
	return Int64(value)
}

// Int64FromPtr 将 int64 指针转换为 sql.NullInt64，nil 会转换为 NULL。
func Int64FromPtr(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *value, Valid: true}
}

// NullInt64FromPtr 将 int64 指针转换为 sql.NullInt64，nil 会转换为 NULL。
func NullInt64FromPtr(value *int64) sql.NullInt64 {
	return Int64FromPtr(value)
}

// NullInt64Ptr 将 int64 指针转换为 sql.NullInt64，nil 会转换为 NULL。
func NullInt64Ptr(value *int64) sql.NullInt64 {
	return Int64FromPtr(value)
}

// Int64FromIntPtr 将 int 指针转换为 sql.NullInt64，nil 会转换为 NULL。
func Int64FromIntPtr(value *int) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*value), Valid: true}
}

// Int64Value 从 sql.NullInt64 取出 int64，NULL 返回 0。
func Int64Value(value sql.NullInt64) int64 {
	if !value.Valid {
		return 0
	}
	return value.Int64
}

// NullInt64Value 从 sql.NullInt64 取出 int64，NULL 返回 0。
func NullInt64Value(value sql.NullInt64) int64 {
	return Int64Value(value)
}

// PtrInt64Value 从 int64 指针取出 int64，nil 返回 0。
func PtrInt64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

// Int64Ptr 将 sql.NullInt64 转换为 int64 指针，NULL 返回 nil。
func Int64Ptr(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}

// PtrNullInt64 将 sql.NullInt64 转换为 int64 指针，NULL 返回 nil。
func PtrNullInt64(value sql.NullInt64) *int64 {
	return Int64Ptr(value)
}

// PtrInt64 将 int64 转换为 int64 指针，非正整数返回 nil。
func PtrInt64(value int64) *int64 {
	if value <= 0 {
		return nil
	}
	return &value
}

// PatchNullInt64 按可选 int64 补丁更新 sql.NullInt64，nil 保持 current。
func PatchNullInt64(current sql.NullInt64, value *int64) sql.NullInt64 {
	if value == nil {
		return current
	}
	return Int64(*value)
}

// Int64NullChanged 判断 sql.NullInt64 是否改变当前可选 int64。
func Int64NullChanged(current *int64, value sql.NullInt64) bool {
	if current == nil {
		return value.Valid
	}
	return !value.Valid || *current != value.Int64
}

// Int64PtrChanged 判断可选 int64 指针补丁是否改变当前值，nil 补丁表示不变。
func Int64PtrChanged(current *int64, value *int64) bool {
	if value == nil {
		return false
	}
	if current == nil {
		return *value != 0
	}
	return *current != *value
}

// TimeFromPtr 将时间指针转换为 sql.NullTime，nil 会转换为 NULL。
func TimeFromPtr(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *value, Valid: true}
}

// NullTimeFromPtr 将时间指针转换为 sql.NullTime，nil 会转换为 NULL。
func NullTimeFromPtr(value *time.Time) sql.NullTime {
	return TimeFromPtr(value)
}

// NullTime 将时间指针转换为 sql.NullTime，nil 会转换为 NULL。
func NullTime(value *time.Time) sql.NullTime {
	return TimeFromPtr(value)
}

// TimePtr 将 sql.NullTime 转换为时间指针，NULL 返回 nil。
func TimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}

// PtrNullTime 将 sql.NullTime 转换为时间指针，NULL 返回 nil。
func PtrNullTime(value sql.NullTime) *time.Time {
	return TimePtr(value)
}

// PtrTimeValue 从时间指针取出时间，nil 返回零值时间。
func PtrTimeValue(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

// PtrTime 将时间转换为时间指针，零值时间返回 nil。
func PtrTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

// TimeString 将 sql.NullTime 格式化为日期时间字符串，NULL 返回空字符串。
func TimeString(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format(time.DateTime)
}

// NullTimeString 将 sql.NullTime 格式化为日期时间字符串，NULL 返回空字符串。
func NullTimeString(value sql.NullTime) string {
	return TimeString(value)
}

// TimeStringPtr 将时间指针格式化为日期时间字符串指针，nil 返回 nil。
func TimeStringPtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format(time.DateTime)
	return &formatted
}

// PtrTimeString 将时间指针格式化为日期时间字符串，nil 返回空字符串。
func PtrTimeString(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(time.DateTime)
}

// NullTimeStringPtr 将 sql.NullTime 格式化为日期时间字符串指针，NULL 返回 nil。
func NullTimeStringPtr(value sql.NullTime) *string {
	if !value.Valid {
		return nil
	}
	formatted := value.Time.Format(time.DateTime)
	return &formatted
}

// PtrBytesValue 从字节切片指针取出字节切片，nil 返回 nil。
func PtrBytesValue(value *[]byte) []byte {
	if value == nil {
		return nil
	}
	return *value
}

// BytesStringArg 将字节切片转换为数据库字符串参数，空切片返回 nil。
func BytesStringArg(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return string(value)
}

// BytesStringPtr 将字节切片转换为字符串指针，空切片返回 nil。
func BytesStringPtr(value []byte) *string {
	if len(value) == 0 {
		return nil
	}
	text := string(value)
	return &text
}

// JSON 将 map 转换为数据库 JSON 字符串参数，空 map 返回 nil。
func JSON(value map[string]any) (any, error) {
	if len(value) == 0 {
		return nil, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

// NullableJSON 将 map 转换为数据库 JSON 字符串参数，空 map 返回 nil。
func NullableJSON(value map[string]any) (any, error) {
	return JSON(value)
}

// StringScanner 返回把 SQL NULL 扫描为空字符串的 scanner。
func StringScanner(dest *string) sql.Scanner {
	return stringScanner{dest: dest}
}

type stringScanner struct {
	dest *string
}

func (s stringScanner) Scan(src any) error {
	var value sql.NullString
	if err := value.Scan(src); err != nil {
		return err
	}
	if value.Valid {
		*s.dest = value.String
		return nil
	}
	*s.dest = ""
	return nil
}

// Int64Scanner 返回把 SQL NULL 扫描为 0 的 scanner。
func Int64Scanner(dest *int64) sql.Scanner {
	return int64Scanner{dest: dest}
}

type int64Scanner struct {
	dest *int64
}

func (s int64Scanner) Scan(src any) error {
	var value sql.NullInt64
	if err := value.Scan(src); err != nil {
		return err
	}
	if value.Valid {
		*s.dest = value.Int64
		return nil
	}
	*s.dest = 0
	return nil
}
