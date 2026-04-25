package captcha_test

import (
	"context"
	"fmt"
	"time"

	"github.com/f2xme/gox/captcha/adapter/memory"
	"github.com/f2xme/gox/captcha/generator/base64"
)

// Example 演示基本的验证码生成和验证
func Example() {
	ctx := context.Background()

	// 使用便捷构造函数创建验证码实例
	c := memory.NewCaptcha()

	// 生成验证码
	id, b64s, err := c.Generate(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("验证码 ID: %s\n", id)
	fmt.Printf("验证码数据长度: %d\n", len(b64s))

	// 验证码验证（实际应用中答案由用户输入）
	// 这里我们无法预知答案，所以只演示 API 调用
	_, _ = c.Verify(ctx, id, "1234")
}

// ExampleNewCaptcha 演示如何使用便捷构造函数创建验证码实例
func ExampleNewCaptcha() {
	// 使用默认配置
	c := memory.NewCaptcha()
	fmt.Printf("验证码实例类型: %T\n", c)
}

// ExampleNewCaptcha_withOptions 演示如何使用选项创建验证码实例
func ExampleNewCaptcha_withOptions() {
	c := memory.NewCaptcha(
		memory.WithCaptchaType(base64.TypeDigit),
		memory.WithLength(6),
		memory.WithSize(300, 100),
		memory.WithCaptchaTTL(10*time.Minute),
	)
	fmt.Printf("验证码实例类型: %T\n", c)
}

// Example_memoryAdapter 演示如何使用内存适配器
func Example_memoryAdapter() {
	// 创建内存适配器，保留 1000 个验证码
	store := memory.New(memory.WithMaxSize(1000))
	fmt.Printf("存储类型: %T\n", store)
}

// ExampleWithCaptchaType 演示如何设置验证码类型
func ExampleWithCaptchaType() {
	ctx := context.Background()

	// 数字验证码
	digitCaptcha := memory.NewCaptcha(memory.WithCaptchaType(base64.TypeDigit))
	id1, _, _ := digitCaptcha.Generate(ctx)
	fmt.Printf("数字验证码 ID: %s\n", id1)

	// 字母验证码
	stringCaptcha := memory.NewCaptcha(memory.WithCaptchaType(base64.TypeString))
	id2, _, _ := stringCaptcha.Generate(ctx)
	fmt.Printf("字母验证码 ID: %s\n", id2)

	// 算术验证码
	mathCaptcha := memory.NewCaptcha(memory.WithCaptchaType(base64.TypeMath))
	id3, _, _ := mathCaptcha.Generate(ctx)
	fmt.Printf("算术验证码 ID: %s\n", id3)

	// 音频验证码
	audioCaptcha := memory.NewCaptcha(memory.WithCaptchaType(base64.TypeAudio))
	id4, _, _ := audioCaptcha.Generate(ctx)
	fmt.Printf("音频验证码 ID: %s\n", id4)
}

// ExampleWithLength 演示如何设置验证码长度
func ExampleWithLength() {
	// 4 位验证码
	c4 := memory.NewCaptcha(memory.WithLength(4))
	fmt.Printf("4 位验证码: %T\n", c4)

	// 6 位验证码
	c6 := memory.NewCaptcha(memory.WithLength(6))
	fmt.Printf("6 位验证码: %T\n", c6)
}

// ExampleWithSize 演示如何设置验证码尺寸
func ExampleWithSize() {
	c := memory.NewCaptcha(memory.WithSize(300, 100))
	fmt.Printf("验证码实例: %T\n", c)
}

// ExampleWithNoiseCount 演示如何设置噪点数量
func ExampleWithNoiseCount() {
	c := memory.NewCaptcha(memory.WithNoiseCount(5))
	fmt.Printf("验证码实例: %T\n", c)
}

// Example_generate 演示如何生成验证码
func Example_generate() {
	ctx := context.Background()
	c := memory.NewCaptcha()

	id, b64s, err := c.Generate(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("ID 长度: %d\n", len(id))
	fmt.Printf("数据长度 > 0: %v\n", len(b64s) > 0)
}

// Example_verify 演示如何验证验证码
func Example_verify() {
	ctx := context.Background()
	c := memory.NewCaptcha()

	// 生成验证码
	id, _, err := c.Generate(ctx)
	if err != nil {
		panic(err)
	}

	// 验证错误答案
	result, _ := c.Verify(ctx, id, "wrong")
	fmt.Printf("错误答案验证结果: %v\n", result)
}

// Example_delete 演示如何删除验证码
func Example_delete() {
	ctx := context.Background()
	c := memory.NewCaptcha()

	// 生成验证码
	id, _, err := c.Generate(ctx)
	if err != nil {
		panic(err)
	}

	// 删除验证码
	err = c.Delete(ctx, id)
	if err != nil {
		panic(err)
	}

	fmt.Println("验证码已删除")
}

// Example_regenerate 演示如何重新生成验证码
func Example_regenerate() {
	ctx := context.Background()
	c := memory.NewCaptcha()

	// 生成验证码
	id, data1, err := c.Generate(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("原始数据长度: %d\n", len(data1))

	// 重新生成
	data2, err := c.Regenerate(ctx, id)
	if err != nil {
		panic(err)
	}

	fmt.Printf("新数据长度: %d\n", len(data2))
}
