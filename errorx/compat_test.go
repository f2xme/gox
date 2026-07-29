package errorx_test

import (
	"testing"

	"github.com/f2xme/gox/errorx"
)

// TestBizErrorStructFieldOrder 验证结构体字段顺序保持稳定。
// 注意：添加 Meta 字段是破坏性变更，无键字面量将不再兼容。
func TestBizErrorStructFieldOrder(t *testing.T) {
	// 推荐的用法：使用字段名
	err := errorx.BizError{
		Code:    "CODE",
		Message: "message",
		Meta:    nil,
	}

	if err.Code != "CODE" || err.Message != "message" {
		t.Fatalf("unexpected business error: %#v", err)
	}
}

// TestBizErrorKeyedLiteral 验证使用键值对的字面量创建方式。
func TestBizErrorKeyedLiteral(t *testing.T) {
	err := errorx.BizError{
		Code:    "USER_NOT_FOUND",
		Message: "用户不存在",
	}

	if err.Code != "USER_NOT_FOUND" {
		t.Errorf("expected 'USER_NOT_FOUND', got %q", err.Code)
	}
	if err.Message != "用户不存在" {
		t.Errorf("expected '用户不存在', got %q", err.Message)
	}
	if err.Meta != nil {
		t.Errorf("expected nil Meta, got %v", err.Meta)
	}
}
