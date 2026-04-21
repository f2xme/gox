package httpx

import (
	"testing"
	"time"
)

func TestValue_String(t *testing.T) {
	if Value("hello").String() != "hello" {
		t.Fatal("String 应返回原始字符串")
	}
	if Value("").String() != "" {
		t.Fatal("空值 String 应返回空字符串")
	}
}

func TestValue_Exists(t *testing.T) {
	if !Value("x").Exists() {
		t.Fatal("非空值 Exists 应为 true")
	}
	if Value("").Exists() {
		t.Fatal("空值 Exists 应为 false")
	}
}

func TestValue_Or(t *testing.T) {
	if got := Value("").Or("def"); got != "def" {
		t.Fatalf("空值应回退默认值，got %q", got)
	}
	if got := Value("x").Or("def"); got != "x" {
		t.Fatalf("非空值应返回原值，got %q", got)
	}
}

func TestValue_Int(t *testing.T) {
	tests := []struct {
		name    string
		in      Value
		want    int
		wantErr bool
	}{
		{"正数", "42", 42, false},
		{"零", "0", 0, false},
		{"负数", "-7", -7, false},
		{"非法", "abc", 0, true},
		{"空值", "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.in.Int()
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("got %d want %d", got, tt.want)
			}
		})
	}
}

func TestValue_IntOr(t *testing.T) {
	if Value("42").IntOr(1) != 42 {
		t.Fatal("合法值应返回解析结果")
	}
	if Value("").IntOr(1) != 1 {
		t.Fatal("空值应回退默认值")
	}
	if Value("abc").IntOr(99) != 99 {
		t.Fatal("非法值应回退默认值")
	}
}

func TestValue_Int64(t *testing.T) {
	got, err := Value("9223372036854775807").Int64()
	if err != nil {
		t.Fatal(err)
	}
	if got != 9223372036854775807 {
		t.Fatalf("got %d", got)
	}
	if Value("abc").Int64Or(-1) != -1 {
		t.Fatal("非法值应回退默认值")
	}
}

func TestValue_Uint64(t *testing.T) {
	got, err := Value("100").Uint64()
	if err != nil || got != 100 {
		t.Fatalf("got %d err %v", got, err)
	}
	if Value("-1").Uint64Or(0) != 0 {
		t.Fatal("负数应被视为非法")
	}
}

func TestValue_Float64(t *testing.T) {
	got, err := Value("3.14").Float64()
	if err != nil || got != 3.14 {
		t.Fatalf("got %v err %v", got, err)
	}
	if Value("xx").Float64Or(0.5) != 0.5 {
		t.Fatal("非法值应回退默认值")
	}
}

func TestValue_Bool(t *testing.T) {
	trueVals := []Value{"1", "true", "TRUE", "yes", "YES", "on", "t"}
	for _, v := range trueVals {
		b, err := v.Bool()
		if err != nil || !b {
			t.Fatalf("%q 应解析为 true, got=%v err=%v", v, b, err)
		}
	}
	falseVals := []Value{"0", "false", "FALSE", "no", "off", "f"}
	for _, v := range falseVals {
		b, err := v.Bool()
		if err != nil || b {
			t.Fatalf("%q 应解析为 false, got=%v err=%v", v, b, err)
		}
	}
	if _, err := Value("abc").Bool(); err == nil {
		t.Fatal("非法布尔应返回 error")
	}
	if !Value("").BoolOr(true) {
		t.Fatal("空值应回退默认值 true")
	}
}

func TestValue_Time(t *testing.T) {
	const layout = time.RFC3339
	ts := "2025-01-02T15:04:05Z"
	got, err := Value(ts).Time(layout)
	if err != nil {
		t.Fatal(err)
	}
	if got.Year() != 2025 {
		t.Fatalf("got %v", got)
	}
	def := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	if !Value("bad").TimeOr(layout, def).Equal(def) {
		t.Fatal("非法时间应回退默认值")
	}
}

func TestValue_Duration(t *testing.T) {
	d, err := Value("1h30m").Duration()
	if err != nil {
		t.Fatal(err)
	}
	if d != 90*time.Minute {
		t.Fatalf("got %v", d)
	}
	if Value("bad").DurationOr(5*time.Second) != 5*time.Second {
		t.Fatal("非法 duration 应回退默认值")
	}
}

func TestValue_Split(t *testing.T) {
	got := Value("a,b,c").Split(",")
	if len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("got %v", got)
	}
	if Value("").Split(",") != nil {
		t.Fatal("空值应返回 nil")
	}
}

func TestValue_UntypedStringCompare(t *testing.T) {
	// 回归：Value 应能直接与字符串字面量比较（底层类型为 string）
	v := Value("x")
	if v == "" {
		t.Fatal("非空 Value 不应等于空字符串")
	}
	if v != "x" {
		t.Fatal("Value 应能与字符串字面量比较")
	}
}
