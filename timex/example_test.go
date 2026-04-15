package timex_test

import (
	"fmt"
	"time"

	"github.com/f2xme/gox/timex"
)

func ExampleFormat() {
	t := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	fmt.Println(timex.Format(t, timex.DateTime))
	fmt.Println(timex.Format(t, timex.Date))
	// Output:
	// 2024-01-15 14:30:45
	// 2024-01-15
}

func ExampleParse() {
	t, _ := timex.Parse("2024-01-15 14:30:45")
	fmt.Println(t.Year(), t.Month(), t.Day())
	// Output:
	// 2024 January 15
}

func ExampleToTimezone() {
	utc := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	shanghai, _ := timex.ToTimezone(utc, "Asia/Shanghai")
	fmt.Println(shanghai.Hour())
	// Output:
	// 22
}

func ExampleAddWorkdays() {
	// Monday
	monday := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	// Add 5 workdays -> next Monday (skip weekend)
	result := timex.AddWorkdays(monday, 5)
	fmt.Println(result.Weekday(), result.Day())
	// Output:
	// Monday 22
}

func ExampleStartOfDay() {
	t := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	start := timex.StartOfDay(t)
	fmt.Printf("%02d:%02d:%02d\n", start.Hour(), start.Minute(), start.Second())
	// Output:
	// 00:00:00
}

func ExampleIsWeekend() {
	saturday := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)
	monday := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	fmt.Println(timex.IsWeekend(saturday))
	fmt.Println(timex.IsWeekend(monday))
	// Output:
	// true
	// false
}

func ExampleHumanDuration() {
	fmt.Println(timex.HumanDuration(30 * time.Second))
	fmt.Println(timex.HumanDuration(2 * time.Hour))
	fmt.Println(timex.HumanDuration(48 * time.Hour))
	// Output:
	// 30 seconds
	// 2 hours
	// 2 days
}

func ExampleParseDuration() {
	d, _ := timex.ParseDuration("2d")
	fmt.Println(d)
	d, _ = timex.ParseDuration("1w")
	fmt.Println(d)
	// Output:
	// 48h0m0s
	// 168h0m0s
}
