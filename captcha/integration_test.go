package captcha_test

import (
	"context"
	"testing"

	cachememory "github.com/f2xme/gox/cache/adapter/memory"
	cacheadapter "github.com/f2xme/gox/captcha/adapter/cache"
	"github.com/f2xme/gox/captcha/adapter/memory"
	"github.com/f2xme/gox/captcha/generator/base64"
)

// TestIntegration_MemoryAdapter 测试内存适配器的完整流程
func TestIntegration_MemoryAdapter(t *testing.T) {
	ctx := context.Background()

	// 使用便捷构造函数
	c, err := memory.NewCaptcha(
		memory.WithCaptchaType(base64.TypeDigit),
		memory.WithLength(6),
	)
	if err != nil {
		t.Fatalf("NewCaptcha() error = %v", err)
	}

	// 生成验证码
	challenge, err := c.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if challenge.ID == "" || challenge.Data == "" {
		t.Fatal("Generated captcha has empty id or data")
	}

	// 验证错误答案
	ok, err := c.Verify(ctx, challenge.ID, "wrong")
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if ok {
		t.Error("Wrong answer should not verify")
	}

	// 重新生成
	newChallenge, err := c.Regenerate(ctx, challenge.ID)
	if err != nil {
		t.Fatalf("Regenerate() error = %v", err)
	}
	if newChallenge.Data == "" {
		t.Error("Regenerated data is empty")
	}

	// 删除
	if err := c.Delete(ctx, challenge.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// 验证已删除
	ok, err = c.Verify(ctx, challenge.ID, "anything")
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
	c, _ := cachememory.New()

	// 使用便捷构造函数
	captcha, err := cacheadapter.NewCaptcha(c,
		cacheadapter.WithCaptchaType(base64.TypeString),
		cacheadapter.WithLength(4),
	)
	if err != nil {
		t.Fatalf("NewCaptcha() error = %v", err)
	}

	// 生成验证码
	challenge, err := captcha.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if challenge.ID == "" || challenge.Data == "" {
		t.Fatal("Generated captcha has empty id or data")
	}

	// 验证错误答案
	ok, err := captcha.Verify(ctx, challenge.ID, "WRONG")
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if ok {
		t.Error("Wrong answer should not verify")
	}

	// 重新生成
	newChallenge, err := captcha.Regenerate(ctx, challenge.ID)
	if err != nil {
		t.Fatalf("Regenerate() error = %v", err)
	}
	if newChallenge.Data == "" {
		t.Error("Regenerated data is empty")
	}

	// 删除
	if err := captcha.Delete(ctx, challenge.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// 验证已删除
	ok, err = captcha.Verify(ctx, challenge.ID, "anything")
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
			c, err := memory.NewCaptcha(memory.WithCaptchaType(tt.typ))
			if err != nil {
				t.Fatalf("NewCaptcha() error = %v", err)
			}

			challenge, err := c.Generate(ctx)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if challenge.ID == "" || challenge.Data == "" {
				t.Error("Generated captcha has empty id or data")
			}

			// 清理
			_ = c.Delete(ctx, challenge.ID)
		})
	}
}
