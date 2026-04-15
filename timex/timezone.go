package timex

import (
	"time"
)

// ToTimezone 将时间转换为指定时区
func ToTimezone(t time.Time, tz string) (time.Time, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Time{}, err
	}
	return t.In(loc), nil
}

// ToUTC 将时间转换为 UTC
func ToUTC(t time.Time) time.Time {
	return t.UTC()
}

// ToLocal 将时间转换为本地时区
func ToLocal(t time.Time) time.Time {
	return t.Local()
}
