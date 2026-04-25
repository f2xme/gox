package captcha_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/f2xme/gox/captcha"
	"github.com/f2xme/gox/captcha/adapter/memory"
	"github.com/f2xme/gox/captcha/generator/base64"
)

func mustCaptcha(c captcha.Service, err error) captcha.Service {
	if err != nil {
		log.Fatalf("创建验证码实例失败: %v", err)
	}
	return c
}

func mustNoError(err error) {
	if err != nil {
		log.Fatalf("示例执行失败: %v", err)
	}
}

func printfln(format string, args ...any) {
	fmt.Printf(format, args...)
	fmt.Println()
}

// Example 演示基本的验证码生成和验证
func Example() {
	ctx := context.Background()

	// 使用便捷构造函数创建验证码实例
	c := mustCaptcha(memory.NewCaptcha())

	// 生成验证码
	challenge, err := c.Generate(ctx)
	mustNoError(err)

	printfln("验证码 ID: %s", challenge.ID)
	printfln("验证码数据长度: %d", len(challenge.Data))

	// 验证码验证（实际应用中答案由用户输入）
	// 这里我们无法预知答案，所以只演示 API 调用
	_, _ = c.Verify(ctx, challenge.ID, "1234")
}

// ExampleNewCaptcha 演示如何使用便捷构造函数创建验证码实例
func ExampleNewCaptcha() {
	// 使用默认配置
	c := mustCaptcha(memory.NewCaptcha())
	printfln("验证码实例类型: %T", c)
}

// ExampleNewCaptcha_withOptions 演示如何使用选项创建验证码实例
func ExampleNewCaptcha_withOptions() {
	c := mustCaptcha(memory.NewCaptcha(
		memory.WithCaptchaType(base64.TypeDigit),
		memory.WithLength(6),
		memory.WithSize(300, 100),
		memory.WithCaptchaTTL(10*time.Minute),
	))
	printfln("验证码实例类型: %T", c)
}

// Example_memoryAdapter 演示如何使用内存适配器
func Example_memoryAdapter() {
	// 创建内存适配器，保留 1000 个验证码
	store := memory.New(memory.WithMaxSize(1000))
	printfln("存储类型: %T", store)
}

// ExampleWithCaptchaType 演示如何设置验证码类型
func ExampleWithCaptchaType() {
	ctx := context.Background()

	// 数字验证码
	digitCaptcha := mustCaptcha(memory.NewCaptcha(memory.WithCaptchaType(base64.TypeDigit)))
	challenge1, _ := digitCaptcha.Generate(ctx)
	printfln("数字验证码 ID: %s", challenge1.ID)

	// 字母验证码
	stringCaptcha := mustCaptcha(memory.NewCaptcha(memory.WithCaptchaType(base64.TypeString)))
	challenge2, _ := stringCaptcha.Generate(ctx)
	printfln("字母验证码 ID: %s", challenge2.ID)

	// 算术验证码
	mathCaptcha := mustCaptcha(memory.NewCaptcha(memory.WithCaptchaType(base64.TypeMath)))
	challenge3, _ := mathCaptcha.Generate(ctx)
	printfln("算术验证码 ID: %s", challenge3.ID)

	// 音频验证码
	audioCaptcha := mustCaptcha(memory.NewCaptcha(memory.WithCaptchaType(base64.TypeAudio)))
	challenge4, _ := audioCaptcha.Generate(ctx)
	printfln("音频验证码 ID: %s", challenge4.ID)
}

// ExampleWithLength 演示如何设置验证码长度
func ExampleWithLength() {
	// 4 位验证码
	c4 := mustCaptcha(memory.NewCaptcha(memory.WithLength(4)))
	printfln("4 位验证码: %T", c4)

	// 6 位验证码
	c6 := mustCaptcha(memory.NewCaptcha(memory.WithLength(6)))
	printfln("6 位验证码: %T", c6)
}

// ExampleWithSize 演示如何设置验证码尺寸
func ExampleWithSize() {
	c := mustCaptcha(memory.NewCaptcha(memory.WithSize(300, 100)))
	printfln("验证码实例: %T", c)
}

// ExampleWithNoiseCount 演示如何设置噪点数量
func ExampleWithNoiseCount() {
	c := mustCaptcha(memory.NewCaptcha(memory.WithNoiseCount(5)))
	printfln("验证码实例: %T", c)
}

// Example_generate 演示如何生成验证码
func Example_generate() {
	ctx := context.Background()
	c := mustCaptcha(memory.NewCaptcha())

	challenge, err := c.Generate(ctx)
	mustNoError(err)

	printfln("ID 长度: %d", len(challenge.ID))
	printfln("数据长度 > 0: %v", len(challenge.Data) > 0)
}

// Example_verify 演示如何验证验证码
func Example_verify() {
	ctx := context.Background()
	c := mustCaptcha(memory.NewCaptcha())

	// 生成验证码
	challenge, err := c.Generate(ctx)
	mustNoError(err)

	// 验证错误答案
	result, _ := c.Verify(ctx, challenge.ID, "wrong")
	printfln("错误答案验证结果: %v", result)
}

// Example_delete 演示如何删除验证码
func Example_delete() {
	ctx := context.Background()
	c := mustCaptcha(memory.NewCaptcha())

	// 生成验证码
	challenge, err := c.Generate(ctx)
	mustNoError(err)

	// 删除验证码
	err = c.Delete(ctx, challenge.ID)
	mustNoError(err)

	fmt.Println("验证码已删除")
}

// Example_regenerate 演示如何重新生成验证码
func Example_regenerate() {
	ctx := context.Background()
	c := mustCaptcha(memory.NewCaptcha())

	// 生成验证码
	challenge, err := c.Generate(ctx)
	mustNoError(err)

	printfln("原始数据长度: %d", len(challenge.Data))

	// 重新生成
	regenerated, err := c.Regenerate(ctx, challenge.ID)
	mustNoError(err)

	printfln("新数据长度: %d", len(regenerated.Data))
}
