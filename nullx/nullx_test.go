package nullx

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"
)

func TestAliases(t *testing.T) {
	var _ NullableString = sql.NullString{}
	var _ NullableInt64 = sql.NullInt64{}
	var _ NullableTime = sql.NullTime{}
	var _ NullableBool = sql.NullBool{}
}

func TestString(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  sql.NullString
	}{
		{name: "empty", value: "", want: sql.NullString{}},
		{name: "value", value: "alice", want: sql.NullString{String: "alice", Valid: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := String(tt.value); got != tt.want {
				t.Fatalf("String(%q) = %#v, want %#v", tt.value, got, tt.want)
			}
			if got := NullString(tt.value); got != tt.want {
				t.Fatalf("NullString(%q) = %#v, want %#v", tt.value, got, tt.want)
			}
		})
	}
}

func TestStringFromPtr(t *testing.T) {
	empty := ""
	value := "alice"

	tests := []struct {
		name  string
		value *string
		want  sql.NullString
	}{
		{name: "nil", value: nil, want: sql.NullString{}},
		{name: "empty", value: &empty, want: sql.NullString{String: "", Valid: true}},
		{name: "value", value: &value, want: sql.NullString{String: "alice", Valid: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringFromPtr(tt.value); got != tt.want {
				t.Fatalf("StringFromPtr() = %#v, want %#v", got, tt.want)
			}
			if got := NullStringFromPtr(tt.value); got != tt.want {
				t.Fatalf("NullStringFromPtr() = %#v, want %#v", got, tt.want)
			}
			if got := StringNull(tt.value); got != tt.want {
				t.Fatalf("StringNull() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestStringValue(t *testing.T) {
	tests := []struct {
		name  string
		value sql.NullString
		want  string
	}{
		{name: "null", value: sql.NullString{}, want: ""},
		{name: "value", value: sql.NullString{String: "alice", Valid: true}, want: "alice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringValue(tt.value); got != tt.want {
				t.Fatalf("StringValue(%#v) = %q, want %q", tt.value, got, tt.want)
			}
			if got := NullStringValue(tt.value); got != tt.want {
				t.Fatalf("NullStringValue(%#v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestPtrStringValue(t *testing.T) {
	value := "alice"
	if got := PtrStringValue(nil); got != "" {
		t.Fatalf("PtrStringValue(nil) = %q, want empty", got)
	}
	if got := PtrStringValue(&value); got != "alice" {
		t.Fatalf("PtrStringValue(value) = %q, want alice", got)
	}
}

func TestStringPtr(t *testing.T) {
	if got := StringPtr(sql.NullString{}); got != nil {
		t.Fatalf("StringPtr(null) = %#v, want nil", got)
	}
	if got := PtrNullString(sql.NullString{}); got != nil {
		t.Fatalf("PtrNullString(null) = %#v, want nil", got)
	}

	got := StringPtr(sql.NullString{String: "", Valid: true})
	if got == nil || *got != "" {
		t.Fatalf("StringPtr(empty valid) = %#v, want empty string ptr", got)
	}

	got = StringPtr(sql.NullString{String: "alice", Valid: true})
	if got == nil || *got != "alice" {
		t.Fatalf("StringPtr(value) = %#v, want alice ptr", got)
	}

	if got := PtrString(""); got != nil {
		t.Fatalf("PtrString(empty) = %#v, want nil", got)
	}
	if got := PtrString("alice"); got == nil || *got != "alice" {
		t.Fatalf("PtrString(value) = %#v, want alice ptr", got)
	}
}

func TestPatchNullString(t *testing.T) {
	current := sql.NullString{String: "old", Valid: true}
	empty := ""
	value := "new"

	if got := PatchNullString(current, nil); got != current {
		t.Fatalf("PatchNullString(nil) = %#v, want current", got)
	}
	if got := PatchNullString(current, &empty); got.Valid {
		t.Fatalf("PatchNullString(empty) = %#v, want invalid", got)
	}
	if got := PatchNullString(current, &value); got != (sql.NullString{String: "new", Valid: true}) {
		t.Fatalf("PatchNullString(value) = %#v, want new", got)
	}

	if got := PatchStringPtrAsNull(nil, nil); got.Valid {
		t.Fatalf("PatchStringPtrAsNull(nil, nil) = %#v, want invalid", got)
	}
	if got := PatchStringPtrAsNull(&empty, nil); got != (sql.NullString{String: "", Valid: true}) {
		t.Fatalf("PatchStringPtrAsNull(empty current) = %#v, want valid empty", got)
	}
}

func TestStringChanged(t *testing.T) {
	empty := ""
	old := "old"
	next := "new"

	if StringPtrChanged(&old, nil) {
		t.Fatal("StringPtrChanged(nil patch) = true, want false")
	}
	if StringPtrChanged(nil, &empty) {
		t.Fatal("StringPtrChanged(nil current, empty patch) = true, want false")
	}
	if !StringPtrChanged(&old, &next) {
		t.Fatal("StringPtrChanged(changed) = false, want true")
	}
	if StringChanged(nil, "") {
		t.Fatal("StringChanged(nil, empty) = true, want false")
	}
	if !StringChanged(&old, next) {
		t.Fatal("StringChanged(changed) = false, want true")
	}
	if !StringNullChanged(&old, sql.NullString{}) {
		t.Fatal("StringNullChanged(valid current, null) = false, want true")
	}
	if StringNullChanged(&old, sql.NullString{String: "old", Valid: true}) {
		t.Fatal("StringNullChanged(same) = true, want false")
	}
}

func TestBool(t *testing.T) {
	tests := []struct {
		name  string
		value bool
		want  sql.NullBool
	}{
		{name: "false", value: false, want: sql.NullBool{Bool: false, Valid: true}},
		{name: "true", value: true, want: sql.NullBool{Bool: true, Valid: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Bool(tt.value); got != tt.want {
				t.Fatalf("Bool(%v) = %#v, want %#v", tt.value, got, tt.want)
			}
		})
	}
}

func TestInt64(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  int64
		valid bool
	}{
		{name: "negative", value: -1, want: -1, valid: false},
		{name: "zero", value: 0, want: 0, valid: false},
		{name: "positive", value: 1, want: 1, valid: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Int64(tt.value)
			if got.Valid != tt.valid {
				t.Fatalf("Int64(%d).Valid = %v, want %v", tt.value, got.Valid, tt.valid)
			}
			if got := NullInt64(tt.value); got.Valid != tt.valid {
				t.Fatalf("NullInt64(%d).Valid = %v, want %v", tt.value, got.Valid, tt.valid)
			}
			if got.Int64 != tt.want {
				t.Fatalf("Int64(%d).Int64 = %d, want %d", tt.value, got.Int64, tt.want)
			}
		})
	}
}

func TestInt64FromPtr(t *testing.T) {
	zero := int64(0)
	value := int64(42)

	tests := []struct {
		name  string
		value *int64
		want  sql.NullInt64
	}{
		{name: "nil", value: nil, want: sql.NullInt64{}},
		{name: "zero", value: &zero, want: sql.NullInt64{Int64: 0, Valid: true}},
		{name: "value", value: &value, want: sql.NullInt64{Int64: 42, Valid: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Int64FromPtr(tt.value); got != tt.want {
				t.Fatalf("Int64FromPtr() = %#v, want %#v", got, tt.want)
			}
			if got := NullInt64FromPtr(tt.value); got != tt.want {
				t.Fatalf("NullInt64FromPtr() = %#v, want %#v", got, tt.want)
			}
			if got := NullInt64Ptr(tt.value); got != tt.want {
				t.Fatalf("NullInt64Ptr() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestInt64FromIntPtr(t *testing.T) {
	zero := 0
	value := 7

	if got := Int64FromIntPtr(nil); got.Valid {
		t.Fatalf("Int64FromIntPtr(nil) = %#v, want invalid", got)
	}
	if got := Int64FromIntPtr(&zero); got != (sql.NullInt64{Int64: 0, Valid: true}) {
		t.Fatalf("Int64FromIntPtr(zero) = %#v, want valid zero", got)
	}
	if got := Int64FromIntPtr(&value); got != (sql.NullInt64{Int64: 7, Valid: true}) {
		t.Fatalf("Int64FromIntPtr(value) = %#v, want valid 7", got)
	}
}

func TestInt64Value(t *testing.T) {
	tests := []struct {
		name  string
		value sql.NullInt64
		want  int64
	}{
		{name: "null", value: sql.NullInt64{}, want: 0},
		{name: "value", value: sql.NullInt64{Int64: 42, Valid: true}, want: 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Int64Value(tt.value); got != tt.want {
				t.Fatalf("Int64Value(%#v) = %d, want %d", tt.value, got, tt.want)
			}
			if got := NullInt64Value(tt.value); got != tt.want {
				t.Fatalf("NullInt64Value(%#v) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

func TestPtrInt64Value(t *testing.T) {
	value := int64(42)
	if got := PtrInt64Value(nil); got != 0 {
		t.Fatalf("PtrInt64Value(nil) = %d, want 0", got)
	}
	if got := PtrInt64Value(&value); got != 42 {
		t.Fatalf("PtrInt64Value(value) = %d, want 42", got)
	}
}

func TestInt64Ptr(t *testing.T) {
	if got := Int64Ptr(sql.NullInt64{}); got != nil {
		t.Fatalf("Int64Ptr(null) = %#v, want nil", got)
	}
	if got := PtrNullInt64(sql.NullInt64{}); got != nil {
		t.Fatalf("PtrNullInt64(null) = %#v, want nil", got)
	}

	got := Int64Ptr(sql.NullInt64{Int64: 0, Valid: true})
	if got == nil || *got != 0 {
		t.Fatalf("Int64Ptr(zero valid) = %#v, want zero ptr", got)
	}

	got = Int64Ptr(sql.NullInt64{Int64: 42, Valid: true})
	if got == nil || *got != 42 {
		t.Fatalf("Int64Ptr(value) = %#v, want 42 ptr", got)
	}

	if got := PtrInt64(0); got != nil {
		t.Fatalf("PtrInt64(0) = %#v, want nil", got)
	}
	if got := PtrInt64(42); got == nil || *got != 42 {
		t.Fatalf("PtrInt64(42) = %#v, want 42 ptr", got)
	}
}

func TestPatchNullInt64(t *testing.T) {
	current := sql.NullInt64{Int64: 10, Valid: true}
	zero := int64(0)
	value := int64(42)

	if got := PatchNullInt64(current, nil); got != current {
		t.Fatalf("PatchNullInt64(nil) = %#v, want current", got)
	}
	if got := PatchNullInt64(current, &zero); got.Valid {
		t.Fatalf("PatchNullInt64(zero) = %#v, want invalid", got)
	}
	if got := PatchNullInt64(current, &value); got != (sql.NullInt64{Int64: 42, Valid: true}) {
		t.Fatalf("PatchNullInt64(value) = %#v, want valid 42", got)
	}
}

func TestInt64Changed(t *testing.T) {
	zero := int64(0)
	old := int64(10)
	next := int64(42)

	if Int64PtrChanged(&old, nil) {
		t.Fatal("Int64PtrChanged(nil patch) = true, want false")
	}
	if Int64PtrChanged(nil, &zero) {
		t.Fatal("Int64PtrChanged(nil current, zero patch) = true, want false")
	}
	if !Int64PtrChanged(&old, &next) {
		t.Fatal("Int64PtrChanged(changed) = false, want true")
	}
	if !Int64NullChanged(&old, sql.NullInt64{}) {
		t.Fatal("Int64NullChanged(valid current, null) = false, want true")
	}
	if Int64NullChanged(&old, sql.NullInt64{Int64: 10, Valid: true}) {
		t.Fatal("Int64NullChanged(same) = true, want false")
	}
}

func TestTime(t *testing.T) {
	now := time.Date(2026, 6, 28, 10, 11, 12, 0, time.UTC)

	if got := TimeFromPtr(nil); got.Valid {
		t.Fatalf("TimeFromPtr(nil) = %#v, want invalid", got)
	}

	gotNull := TimeFromPtr(&now)
	if !gotNull.Valid || !gotNull.Time.Equal(now) {
		t.Fatalf("TimeFromPtr(value) = %#v, want valid %v", gotNull, now)
	}
	if got := NullTimeFromPtr(&now); !got.Valid || !got.Time.Equal(now) {
		t.Fatalf("NullTimeFromPtr(value) = %#v, want valid %v", got, now)
	}
	if got := NullTime(&now); !got.Valid || !got.Time.Equal(now) {
		t.Fatalf("NullTime(value) = %#v, want valid %v", got, now)
	}

	if got := TimePtr(sql.NullTime{}); got != nil {
		t.Fatalf("TimePtr(null) = %#v, want nil", got)
	}
	if got := PtrNullTime(sql.NullTime{}); got != nil {
		t.Fatalf("PtrNullTime(null) = %#v, want nil", got)
	}

	gotPtr := TimePtr(sql.NullTime{Time: now, Valid: true})
	if gotPtr == nil || !gotPtr.Equal(now) {
		t.Fatalf("TimePtr(value) = %#v, want %v", gotPtr, now)
	}

	if got := PtrTimeValue(nil); !got.IsZero() {
		t.Fatalf("PtrTimeValue(nil) = %v, want zero", got)
	}
	if got := PtrTimeValue(&now); !got.Equal(now) {
		t.Fatalf("PtrTimeValue(value) = %v, want %v", got, now)
	}
	if got := PtrTime(time.Time{}); got != nil {
		t.Fatalf("PtrTime(zero) = %#v, want nil", got)
	}
	if got := PtrTime(now); got == nil || !got.Equal(now) {
		t.Fatalf("PtrTime(value) = %#v, want %v", got, now)
	}
}

func TestTimeString(t *testing.T) {
	value := sql.NullTime{Time: time.Date(2026, 6, 28, 10, 11, 12, 0, time.UTC), Valid: true}

	if got := TimeString(sql.NullTime{}); got != "" {
		t.Fatalf("TimeString(null) = %q, want empty", got)
	}
	if got := TimeString(value); got != "2026-06-28 10:11:12" {
		t.Fatalf("TimeString(value) = %q, want %q", got, "2026-06-28 10:11:12")
	}
	if got := NullTimeString(value); got != "2026-06-28 10:11:12" {
		t.Fatalf("NullTimeString(value) = %q, want %q", got, "2026-06-28 10:11:12")
	}
}

func TestTimeStringPtr(t *testing.T) {
	value := time.Date(2026, 6, 28, 10, 11, 12, 0, time.UTC)

	if got := TimeStringPtr(nil); got != nil {
		t.Fatalf("TimeStringPtr(null) = %#v, want nil", got)
	}

	got := TimeStringPtr(&value)
	if got == nil || *got != "2026-06-28 10:11:12" {
		t.Fatalf("TimeStringPtr(value) = %#v, want datetime ptr", got)
	}

	if got := PtrTimeString(nil); got != "" {
		t.Fatalf("PtrTimeString(nil) = %q, want empty", got)
	}
	if got := PtrTimeString(&value); got != "2026-06-28 10:11:12" {
		t.Fatalf("PtrTimeString(value) = %q, want datetime", got)
	}
}

func TestNullTimeStringPtr(t *testing.T) {
	value := sql.NullTime{Time: time.Date(2026, 6, 28, 10, 11, 12, 0, time.UTC), Valid: true}

	if got := NullTimeStringPtr(sql.NullTime{}); got != nil {
		t.Fatalf("NullTimeStringPtr(null) = %#v, want nil", got)
	}

	got := NullTimeStringPtr(value)
	if got == nil || *got != "2026-06-28 10:11:12" {
		t.Fatalf("NullTimeStringPtr(value) = %#v, want datetime ptr", got)
	}
}

func TestBytesHelpers(t *testing.T) {
	value := []byte("hello")

	if got := PtrBytesValue(nil); got != nil {
		t.Fatalf("PtrBytesValue(nil) = %#v, want nil", got)
	}
	if got := PtrBytesValue(&value); string(got) != "hello" {
		t.Fatalf("PtrBytesValue(value) = %q, want hello", string(got))
	}
	if got := BytesStringArg(nil); got != nil {
		t.Fatalf("BytesStringArg(nil) = %#v, want nil", got)
	}
	if got := BytesStringArg(value); got != "hello" {
		t.Fatalf("BytesStringArg(value) = %#v, want hello", got)
	}
	if got := BytesStringPtr(nil); got != nil {
		t.Fatalf("BytesStringPtr(nil) = %#v, want nil", got)
	}
	if got := BytesStringPtr(value); got == nil || *got != "hello" {
		t.Fatalf("BytesStringPtr(value) = %#v, want hello ptr", got)
	}
}

func TestJSON(t *testing.T) {
	if got, err := JSON(nil); err != nil || got != nil {
		t.Fatalf("JSON(nil) = (%#v, %v), want nil, nil", got, err)
	}

	got, err := JSON(map[string]any{"name": "alice"})
	if err != nil {
		t.Fatalf("JSON(value) error = %v", err)
	}
	var decoded map[string]string
	if err := json.Unmarshal([]byte(got.(string)), &decoded); err != nil {
		t.Fatalf("JSON(value) invalid JSON: %v", err)
	}
	if decoded["name"] != "alice" {
		t.Fatalf("JSON(value) decoded = %#v, want name alice", decoded)
	}

	got, err = NullableJSON(map[string]any{"name": "alice"})
	if err != nil || got == nil {
		t.Fatalf("NullableJSON(value) = (%#v, %v), want value, nil", got, err)
	}
}

func TestScanners(t *testing.T) {
	var text string
	if err := StringScanner(&text).Scan(nil); err != nil {
		t.Fatalf("StringScanner(nil) error = %v", err)
	}
	if text != "" {
		t.Fatalf("StringScanner(nil) = %q, want empty", text)
	}
	if err := StringScanner(&text).Scan("alice"); err != nil {
		t.Fatalf("StringScanner(value) error = %v", err)
	}
	if text != "alice" {
		t.Fatalf("StringScanner(value) = %q, want alice", text)
	}

	var number int64
	if err := Int64Scanner(&number).Scan(nil); err != nil {
		t.Fatalf("Int64Scanner(nil) error = %v", err)
	}
	if number != 0 {
		t.Fatalf("Int64Scanner(nil) = %d, want 0", number)
	}
	if err := Int64Scanner(&number).Scan(int64(42)); err != nil {
		t.Fatalf("Int64Scanner(value) error = %v", err)
	}
	if number != 42 {
		t.Fatalf("Int64Scanner(value) = %d, want 42", number)
	}
}
