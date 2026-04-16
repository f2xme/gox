package random

import (
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		charset string
		wantLen int
		wantErr bool
	}{
		{"正常生成", 10, Alphanumeric, 10, false},
		{"长度为0", 0, Alphanumeric, 0, false},
		{"负数长度", -1, Alphanumeric, 0, false},
		{"空字符集", 10, "", 0, false},
		{"单字符集", 5, "a", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := String(tt.length, tt.charset)
			if (err != nil) != tt.wantErr {
				t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("String() length = %v, want %v", len(got), tt.wantLen)
			}
			// 验证所有字符都在字符集中
			for _, c := range got {
				if !strings.ContainsRune(tt.charset, c) {
					t.Errorf("String() contains invalid char %c", c)
				}
			}
		})
	}
}

func TestNumeric(t *testing.T) {
	testCharsetFunc(t, "Numeric", Numeric, Digits)
}

func TestAlpha(t *testing.T) {
	testCharsetFunc(t, "Alpha", Alpha, Letters)
}

func TestAlphaLower(t *testing.T) {
	testCharsetFunc(t, "AlphaLower", AlphaLower, LowerLetters)
}

func TestAlphaUpper(t *testing.T) {
	testCharsetFunc(t, "AlphaUpper", AlphaUpper, UpperLetters)
}

func TestAlphaNumeric(t *testing.T) {
	testCharsetFunc(t, "AlphaNumeric", AlphaNumeric, Alphanumeric)
}

// testCharsetFunc 测试字符集函数的通用逻辑
func testCharsetFunc(t *testing.T, name string, fn func(int) (string, error), charset string) {
	t.Helper()
	result, err := fn(10)
	if err != nil {
		t.Fatalf("%s() error = %v", name, err)
	}
	if len(result) != 10 {
		t.Errorf("%s() length = %v, want 10", name, len(result))
	}
	for _, c := range result {
		if !strings.ContainsRune(charset, c) {
			t.Errorf("%s() contains invalid char %c", name, c)
		}
	}
}

func TestRandomness(t *testing.T) {
	// 生成多个字符串，验证它们不完全相同（概率极低）
	results := make(map[string]bool)
	for range 100 {
		s, err := AlphaNumeric(20)
		if err != nil {
			t.Fatalf("AlphaNumeric() error = %v", err)
		}
		results[s] = true
	}
	// 100 次生成应该有至少 95 个不同的结果
	if len(results) < 95 {
		t.Errorf("Randomness test failed: only %d unique strings out of 100", len(results))
	}
}

func BenchmarkString(b *testing.B) {
	for b.Loop() {
		_, _ = String(32, Alphanumeric)
	}
}

func BenchmarkNumeric(b *testing.B) {
	for b.Loop() {
		_, _ = Numeric(32)
	}
}

func BenchmarkAlphaNumeric(b *testing.B) {
	for b.Loop() {
		_, _ = AlphaNumeric(32)
	}
}
