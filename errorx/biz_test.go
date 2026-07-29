package errorx

import (
	"errors"
	"fmt"
	"sync"
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

func TestBizErrorMetaAndParameterization(t *testing.T) {
	resetRegistryForTest()
	Register("INSUFFICIENT_STOCK", "zh", "库存不足，可用数量：{available}，需要数量：{required}")
	Register("INSUFFICIENT_STOCK", "en", "Insufficient stock, available: {available}, required: {required}")
	SetDefaultLang("zh")

	err := NewBiz("INSUFFICIENT_STOCK").
		WithMeta("available", 5).
		WithMeta("required", 10)

	if value, ok := err.GetMeta("available"); !ok || value != 5 {
		t.Errorf("expected available metadata 5, got %v, %v", value, ok)
	}
	if _, ok := err.GetMeta("missing"); ok {
		t.Error("expected missing metadata")
	}
	if got := err.Error(); got != "库存不足，可用数量：5，需要数量：10" {
		t.Errorf("unexpected default message: %q", got)
	}
	if got := err.Localize("en"); got != "Insufficient stock, available: 5, required: 10" {
		t.Errorf("unexpected localized message: %q", got)
	}

	wrapped := fmt.Errorf("handle request: %w", err)
	meta := GetBizMeta(wrapped)
	if meta["required"] != 10 {
		t.Errorf("expected required metadata 10, got %v", meta["required"])
	}
}

func TestBizErrorWithMetaDoesNotMutateOriginal(t *testing.T) {
	original := NewBizWithMessage("TEST", "{value}")
	first := original.WithMeta("value", "first")
	second := original.WithMeta("value", "second")

	if meta := GetBizMeta(original); meta != nil {
		t.Errorf("expected original metadata unchanged, got %v", meta)
	}
	if got := first.Error(); got != "first" {
		t.Errorf("expected first copy unchanged, got %q", got)
	}
	if got := second.Error(); got != "second" {
		t.Errorf("expected second copy unchanged, got %q", got)
	}

	meta := GetBizMeta(first)
	meta["value"] = "changed"
	if got := first.Error(); got != "first" {
		t.Errorf("expected metadata helper to return a copy, got %q", got)
	}
}

func TestBizErrorWithMetaConcurrentReuse(t *testing.T) {
	shared := NewBizWithMessage("TEST", "{value}")
	var wg sync.WaitGroup

	for i := range 20 {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			want := fmt.Sprint(index)
			if got := shared.WithMeta("value", index).Error(); got != want {
				t.Errorf("expected %q, got %q", want, got)
			}
		}(i)
	}
	wg.Wait()
}

func TestBizErrorParameterizationKeepsUnknownPlaceholder(t *testing.T) {
	err := NewBizWithMessage("TEST", "{known} {unknown}").
		WithMeta("known", "value")

	if got := err.Error(); got != "value {unknown}" {
		t.Errorf("expected unknown placeholder unchanged, got %q", got)
	}
}

func TestBizErrorFormatMessageFastPath(t *testing.T) {
	// 测试没有占位符的消息不被处理
	err := NewBizWithMessage("TEST", "no placeholder").
		WithMeta("key", "value")

	if got := err.Error(); got != "no placeholder" {
		t.Errorf("expected fast path to skip replacement, got %q", got)
	}

	// 测试没有元数据的消息
	err2 := NewBizWithMessage("TEST", "message with {placeholder}")
	if got := err2.Error(); got != "message with {placeholder}" {
		t.Errorf("expected no replacement without meta, got %q", got)
	}
}

func TestBizErrorWithMetaMap(t *testing.T) {
	err := NewBizWithMessage("TEST", "user {user_id} action {action}")

	// 批量设置元数据
	meta := map[string]any{
		"user_id": 12345,
		"action":  "login",
	}
	err = err.WithMetaMap(meta)

	if got := err.Error(); got != "user 12345 action login" {
		t.Errorf("expected 'user 12345 action login', got %q", got)
	}

	// 验证不修改原 map
	meta["user_id"] = 99999
	if got := err.Error(); got != "user 12345 action login" {
		t.Errorf("expected original error unchanged, got %q", got)
	}

	// 测试空 map
	err2 := NewBizWithMessage("TEST2", "message")
	err3 := err2.WithMetaMap(nil)
	if err3 != err2 {
		t.Error("expected WithMetaMap(nil) to return same instance")
	}

	err4 := err2.WithMetaMap(map[string]any{})
	if err4 != err2 {
		t.Error("expected WithMetaMap(empty) to return same instance")
	}

	// 测试与 WithMeta 链式调用组合
	err5 := NewBizWithMessage("TEST", "{a} {b} {c}").
		WithMetaMap(map[string]any{"a": "A", "b": "B"}).
		WithMeta("c", "C")

	if got := err5.Error(); got != "A B C" {
		t.Errorf("expected 'A B C', got %q", got)
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
	if meta := GetBizMeta(stdErr); meta != nil {
		t.Errorf("expected nil metadata, got %v", meta)
	}
}

func TestBizErrorIsDistinctFromError(t *testing.T) {
	if Is(NewBizWithMessage("USER_NOT_FOUND", "用户不存在")) {
		t.Error("expected business error to be distinct from errorx.Error")
	}
}

func TestBizHelpersTypedNil(t *testing.T) {
	var bizErr *BizError
	var err error = bizErr

	if IsBiz(err) {
		t.Error("expected typed nil not to be a business error")
	}
	if code := GetBizCode(err); code != "" {
		t.Errorf("expected empty code, got %q", code)
	}
	if meta := GetBizMeta(err); meta != nil {
		t.Errorf("expected nil metadata, got %v", meta)
	}
	if message := LocalizeBiz(err, "zh"); message != "" {
		t.Errorf("expected empty message, got %q", message)
	}
}
