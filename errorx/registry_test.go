package errorx

import (
	"testing"
)

func TestRegister(t *testing.T) {
	// Clear registry for testing
	registry = make(map[string]map[string]string)

	Register("TEST001", "en", "Test error in English")
	Register("TEST001", "zh", "测试错误")

	msg, ok := getMessage("TEST001", "en")
	if !ok {
		t.Error("expected to find registered message")
	}
	if msg != "Test error in English" {
		t.Errorf("expected 'Test error in English', got %q", msg)
	}

	msg, ok = getMessage("TEST001", "zh")
	if !ok {
		t.Error("expected to find registered message")
	}
	if msg != "测试错误" {
		t.Errorf("expected '测试错误', got %q", msg)
	}
}

func TestSetDefaultLang(t *testing.T) {
	SetDefaultLang("zh")
	if defaultLang != "zh" {
		t.Errorf("expected 'zh', got %q", defaultLang)
	}

	// Reset to default
	SetDefaultLang("en")
}

func TestNewCodeWithLang(t *testing.T) {
	// Clear and setup registry
	registry = make(map[string]map[string]string)
	Register("VAL001", "en", "Validation failed")
	Register("VAL001", "zh", "验证失败")

	err := NewCodeWithLang("VAL001", "en")
	if err.Code != "VAL001" {
		t.Errorf("expected 'VAL001', got %q", err.Code)
	}
	if err.Message != "Validation failed" {
		t.Errorf("expected 'Validation failed', got %q", err.Message)
	}

	err = NewCodeWithLang("VAL001", "zh")
	if err.Message != "验证失败" {
		t.Errorf("expected '验证失败', got %q", err.Message)
	}
}

func TestNewCodeWithLangFallback(t *testing.T) {
	// Clear and setup registry
	registry = make(map[string]map[string]string)
	Register("TEST002", "en", "Test error")

	// Request non-existent language, should fallback to en
	err := NewCodeWithLang("TEST002", "fr")
	if err.Message != "Test error" {
		t.Errorf("expected 'Test error', got %q", err.Message)
	}
}

func TestNewCodeWithLangUnregistered(t *testing.T) {
	// Clear registry
	registry = make(map[string]map[string]string)

	// Request unregistered code
	err := NewCodeWithLang("UNKNOWN", "en")
	if err.Code != "UNKNOWN" {
		t.Errorf("expected 'UNKNOWN', got %q", err.Code)
	}
	if err.Message != "UNKNOWN" {
		t.Errorf("expected 'UNKNOWN', got %q", err.Message)
	}
}

func TestRegisterConcurrency(t *testing.T) {
	// Clear registry
	registry = make(map[string]map[string]string)

	// Test concurrent registration
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			Register("CONCURRENT", "en", "Concurrent test")
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	msg, ok := getMessage("CONCURRENT", "en")
	if !ok {
		t.Error("expected to find registered message")
	}
	if msg != "Concurrent test" {
		t.Errorf("expected 'Concurrent test', got %q", msg)
	}
}
