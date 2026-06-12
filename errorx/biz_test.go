package errorx

import (
	"errors"
	"fmt"
	"testing"
)

func resetRegistryForTest() {
	registryMu.Lock()
	defer registryMu.Unlock()

	registry = make(map[string]map[string]string)
	defaultLang = defaultEnglish
}

func TestNewBiz(t *testing.T) {
	resetRegistryForTest()
	Register("USER_NOT_FOUND", "zh", "用户不存在")
	SetDefaultLang("zh")

	err := NewBiz("USER_NOT_FOUND")
	if err.Code != "USER_NOT_FOUND" {
		t.Errorf("expected 'USER_NOT_FOUND', got %q", err.Code)
	}
	if err.Message != "用户不存在" {
		t.Errorf("expected '用户不存在', got %q", err.Message)
	}
	if err.Error() != "用户不存在" {
		t.Errorf("expected '用户不存在', got %q", err.Error())
	}
}

func TestNewBizUnregistered(t *testing.T) {
	resetRegistryForTest()

	err := NewBiz("UNKNOWN")
	if err.Code != "UNKNOWN" {
		t.Errorf("expected 'UNKNOWN', got %q", err.Code)
	}
	if err.Message != "UNKNOWN" {
		t.Errorf("expected 'UNKNOWN', got %q", err.Message)
	}
}

func TestNewBizWithMessage(t *testing.T) {
	err := NewBizWithMessage("USER_EXISTS", "手机号已注册")
	if err.Code != "USER_EXISTS" {
		t.Errorf("expected 'USER_EXISTS', got %q", err.Code)
	}
	if err.Message != "手机号已注册" {
		t.Errorf("expected '手机号已注册', got %q", err.Message)
	}
}

func TestNewBizMessage(t *testing.T) {
	err := NewBizMessage("手机号已注册")
	if err.Code != "" {
		t.Errorf("expected empty code, got %q", err.Code)
	}
	if err.Message != "手机号已注册" {
		t.Errorf("expected '手机号已注册', got %q", err.Message)
	}
	if err.Error() != "手机号已注册" {
		t.Errorf("expected '手机号已注册', got %q", err.Error())
	}
	if !IsBiz(err) {
		t.Error("expected message-only error to be business error")
	}
	if code := GetBizCode(err); code != "" {
		t.Errorf("expected empty code, got %q", code)
	}
}

func TestNewBizLang(t *testing.T) {
	resetRegistryForTest()
	Register("VERIFY_CODE_INVALID", "en", "Invalid verification code")
	Register("VERIFY_CODE_INVALID", "zh", "验证码错误")

	tests := []struct {
		name string
		lang string
		want string
	}{
		{
			name: "english",
			lang: "en",
			want: "Invalid verification code",
		},
		{
			name: "chinese",
			lang: "zh",
			want: "验证码错误",
		},
		{
			name: "fallback to english",
			lang: "fr",
			want: "Invalid verification code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewBizLang("VERIFY_CODE_INVALID", tt.lang)
			if err.Message != tt.want {
				t.Errorf("expected %q, got %q", tt.want, err.Message)
			}
		})
	}
}

func TestBizErrorLocalize(t *testing.T) {
	resetRegistryForTest()
	Register("USER_NOT_FOUND", "en", "User not found")
	Register("USER_NOT_FOUND", "zh", "用户不存在")

	err := NewBizWithMessage("USER_NOT_FOUND", "默认消息")
	if got := err.Localize("zh"); got != "用户不存在" {
		t.Errorf("expected '用户不存在', got %q", got)
	}
	if got := err.Localize("en"); got != "User not found" {
		t.Errorf("expected 'User not found', got %q", got)
	}
	if got := err.Localize("ja"); got != "User not found" {
		t.Errorf("expected fallback 'User not found', got %q", got)
	}

	unregistered := NewBizWithMessage("UNKNOWN", "fallback")
	if got := unregistered.Localize("zh"); got != "fallback" {
		t.Errorf("expected 'fallback', got %q", got)
	}

	messageOnly := NewBizMessage("仅提示消息")
	if got := messageOnly.Localize("zh"); got != "仅提示消息" {
		t.Errorf("expected '仅提示消息', got %q", got)
	}
}

func TestIsBiz(t *testing.T) {
	bizErr := NewBizWithMessage("USER_NOT_FOUND", "用户不存在")

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "biz error",
			err:  bizErr,
			want: true,
		},
		{
			name: "wrapped biz error",
			err:  fmt.Errorf("handle request: %w", bizErr),
			want: true,
		},
		{
			name: "system errorx error",
			err:  New("system failed"),
			want: false,
		},
		{
			name: "standard error",
			err:  errors.New("standard error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBiz(tt.err)
			if got != tt.want {
				t.Errorf("IsBiz() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBizHelpers(t *testing.T) {
	resetRegistryForTest()
	Register("USER_NOT_FOUND", "en", "User not found")
	Register("USER_NOT_FOUND", "zh", "用户不存在")

	err := NewBizWithMessage("USER_NOT_FOUND", "默认消息")
	wrapped := fmt.Errorf("handle request: %w", err)

	if code := GetBizCode(wrapped); code != "USER_NOT_FOUND" {
		t.Errorf("expected 'USER_NOT_FOUND', got %q", code)
	}
	if got := LocalizeBiz(wrapped, "zh"); got != "用户不存在" {
		t.Errorf("expected '用户不存在', got %q", got)
	}

	stdErr := errors.New("standard error")
	if code := GetBizCode(stdErr); code != "" {
		t.Errorf("expected empty code, got %q", code)
	}
	if got := LocalizeBiz(stdErr, "zh"); got != "standard error" {
		t.Errorf("expected 'standard error', got %q", got)
	}
}

func TestBizErrorIsDistinctFromError(t *testing.T) {
	if Is(NewBizWithMessage("USER_NOT_FOUND", "用户不存在")) {
		t.Error("expected business error to be distinct from errorx.Error")
	}
}
