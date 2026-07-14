package geo

import (
	"errors"
	"testing"
)

func TestNormalizeIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{name: "ipv4", input: "1.2.3.4", want: "1.2.3.4"},
		{name: "ipv4 with spaces", input: "  8.8.8.8  ", want: "8.8.8.8"},
		{name: "ipv6", input: "2001:db8::1", want: "2001:db8::1"},
		{name: "empty", input: "", wantErr: ErrCodeInvalidIP},
		{name: "invalid", input: "not-an-ip", wantErr: ErrCodeInvalidIP},
		{name: "partial", input: "1.2.3", wantErr: ErrCodeInvalidIP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NormalizeIP(tt.input)
			if tt.wantErr != "" {
				if !IsCode(err, tt.wantErr) {
					t.Fatalf("NormalizeIP(%q) error = %v, want code %s", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeIP(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeIP(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLocationStringAndEmpty(t *testing.T) {
	t.Parallel()

	if !(*Location)(nil).Empty() {
		t.Fatal("nil location should be empty")
	}
	if (*Location)(nil).String() != "" {
		t.Fatal("nil location String should be empty")
	}

	onlyIP := &Location{IP: "1.1.1.1"}
	if !onlyIP.Empty() {
		t.Fatal("location with only IP should be empty")
	}
	if onlyIP.String() != "1.1.1.1" {
		t.Fatalf("String() = %q, want IP", onlyIP.String())
	}

	full := &Location{
		IP:       "1.2.3.4",
		Country:  "中国",
		Province: "广东省",
		City:     "深圳市",
		ISP:      "电信",
	}
	if full.Empty() {
		t.Fatal("full location should not be empty")
	}
	if got := full.String(); got != "中国 广东省 深圳市 电信" {
		t.Fatalf("String() = %q", got)
	}
}

func TestLocationClone(t *testing.T) {
	t.Parallel()

	if Clone := (*Location)(nil).Clone(); Clone != nil {
		t.Fatal("nil clone should be nil")
	}

	src := &Location{
		IP:      "1.1.1.1",
		Country: "中国",
		Extra:   map[string]string{"source": "test"},
	}
	cloned := src.Clone()
	if cloned == src {
		t.Fatal("clone should be a different pointer")
	}
	cloned.Country = "美国"
	cloned.Extra["source"] = "changed"
	if src.Country != "中国" {
		t.Fatal("clone should not mutate source country")
	}
	if src.Extra["source"] != "test" {
		t.Fatal("clone should not mutate source extra")
	}
}

func TestErrorHelpers(t *testing.T) {
	t.Parallel()

	err := NewError(ErrCodeInvalidIP, "bad", "x.x.x.x")
	if err.Error() == "" {
		t.Fatal("error string should not be empty")
	}
	if !IsInvalidIP(err) {
		t.Fatal("expected invalid ip")
	}
	if IsNotFound(err) {
		t.Fatal("should not be not found")
	}

	wrapped := WrapError(ErrCodeUnavailable, "down", errors.New("timeout"), "1.1.1.1")
	if !IsUnavailable(wrapped) {
		t.Fatal("expected unavailable")
	}
	if !errors.Is(wrapped, wrapped.Err) {
		t.Fatal("unwrap should work")
	}
	if ErrorCode(errors.New("plain")) != "" {
		t.Fatal("plain error should have empty code")
	}
}
