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
	result, err := Numeric(10)
	if err != nil {
		t.Fatalf("Numeric() error = %v", err)
	}
	if len(result) != 10 {
		t.Errorf("Numeric() length = %v, want 10", len(result))
	}
	for _, c := range result {
		if !strings.ContainsRune(Digits, c) {
			t.Errorf("Numeric() contains non-digit char %c", c)
		}
	}
}

func TestAlpha(t *testing.T) {
	result, err := Alpha(10)
	if err != nil {
		t.Fatalf("Alpha() error = %v", err)
	}
	if len(result) != 10 {
		t.Errorf("Alpha() length = %v, want 10", len(result))
	}
	for _, c := range result {
		if !strings.ContainsRune(Letters, c) {
			t.Errorf("Alpha() contains non-letter char %c", c)
		}
	}
}

func TestAlphaLower(t *testing.T) {
	result, err := AlphaLower(10)
	if err != nil {
		t.Fatalf("AlphaLower() error = %v", err)
	}
	if len(result) != 10 {
		t.Errorf("AlphaLower() length = %v, want 10", len(result))
	}
	for _, c := range result {
		if !strings.ContainsRune(LowerLetters, c) {
			t.Errorf("AlphaLower() contains non-lowercase char %c", c)
		}
	}
}

func TestAlphaUpper(t *testing.T) {
	result, err := AlphaUpper(10)
	if err != nil {
		t.Fatalf("AlphaUpper() error = %v", err)
	}
	if len(result) != 10 {
		t.Errorf("AlphaUpper() length = %v, want 10", len(result))
	}
	for _, c := range result {
		if !strings.ContainsRune(UpperLetters, c) {
			t.Errorf("AlphaUpper() contains non-uppercase char %c", c)
		}
	}
}

func TestAlphaNumeric(t *testing.T) {
	result, err := AlphaNumeric(10)
	if err != nil {
		t.Fatalf("AlphaNumeric() error = %v", err)
	}
	if len(result) != 10 {
		t.Errorf("AlphaNumeric() length = %v, want 10", len(result))
	}
	for _, c := range result {
		if !strings.ContainsRune(Alphanumeric, c) {
			t.Errorf("AlphaNumeric() contains invalid char %c", c)
		}
	}
}

func TestRandomness(t *testing.T) {
	// 生成多个字符串，验证它们不完全相同（概率极低）
	results := make(map[string]bool)
	for i := 0; i < 100; i++ {
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
	for i := 0; i < b.N; i++ {
		_, _ = String(32, Alphanumeric)
	}
}

func BenchmarkNumeric(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Numeric(32)
	}
}

func BenchmarkAlphaNumeric(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = AlphaNumeric(32)
	}
}
