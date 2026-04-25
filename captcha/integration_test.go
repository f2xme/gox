package captcha_test

import (
	"context"
	"testing"

	"github.com/f2xme/gox/cache/adapter/mem"
	cacheadapter "github.com/f2xme/gox/captcha/adapter/cache"
	"github.com/f2xme/gox/captcha/adapter/memory"
	"github.com/f2xme/gox/captcha/generator/base64"
)

// TestIntegration_MemoryAdapter 测试内存适配器的完整流程
func TestIntegration_MemoryAdapter(t *testing.T) {
	ctx := context.Background()

	// 使用便捷构造函数
	c := memory.NewCaptcha(
		memory.WithCaptchaType(base64.TypeDigit),
		memory.WithLength(6),
	)

	// 生成验证码
	id, data, err := c.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if id == "" || data == "" {
		t.Fatal("Generated captcha has empty id or data")
	}

	// 验证错误答案
	ok, err := c.Verify(ctx, id, "wrong")
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if ok {
		t.Error("Wrong answer should not verify")
	}

	// 重新生成
	newData, err := c.Regenerate(ctx, id)
	if err != nil {
		t.Fatalf("Regenerate() error = %v", err)
	}
	if newData == "" {
		t.Error("Regenerated data is empty")
	}

	// 删除
	if err := c.Delete(ctx, id); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// 验证已删除
	ok, err = c.Verify(ctx, id, "anything")
	if err != nil {
		t.Fatalf("Verify() after delete error = %v", err)
	}
	if ok {
		t.Error("Deleted captcha should not verify")
	}
}

// TestIntegration_CacheAdapter 测试缓存适配器的完整流程
func TestIntegration_CacheAdapter(t *testing.T) {
	ctx := context.Background()

	// 创建 cache 实例
	c, _ := mem.New()

	// 使用便捷构造函数
	captcha := cacheadapter.NewCaptcha(c,
		cacheadapter.WithCaptchaType(base64.TypeString),
		cacheadapter.WithLength(4),
	)

	// 生成验证码
	id, data, err := captcha.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if id == "" || data == "" {
		t.Fatal("Generated captcha has empty id or data")
	}

	// 验证错误答案
	ok, err := captcha.Verify(ctx, id, "WRONG")
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if ok {
		t.Error("Wrong answer should not verify")
	}

	// 重新生成
	newData, err := captcha.Regenerate(ctx, id)
	if err != nil {
		t.Fatalf("Regenerate() error = %v", err)
	}
	if newData == "" {
		t.Error("Regenerated data is empty")
	}

	// 删除
	if err := captcha.Delete(ctx, id); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// 验证已删除
	ok, err = captcha.Verify(ctx, id, "anything")
	if err != nil {
		t.Fatalf("Verify() after delete error = %v", err)
	}
	if ok {
		t.Error("Deleted captcha should not verify")
	}
}

// TestIntegration_MultipleTypes 测试不同类型的验证码
func TestIntegration_MultipleTypes(t *testing.T) {
	ctx := context.Background()

	types := []struct {
		name string
		typ  base64.CaptchaType
	}{
		{"digit", base64.TypeDigit},
		{"string", base64.TypeString},
		{"math", base64.TypeMath},
		{"audio", base64.TypeAudio},
	}

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			c := memory.NewCaptcha(memory.WithCaptchaType(tt.typ))

			id, data, err := c.Generate(ctx)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if id == "" || data == "" {
				t.Error("Generated captcha has empty id or data")
			}

			// 清理
			_ = c.Delete(ctx, id)
		})
	}
}
