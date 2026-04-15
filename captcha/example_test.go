package captcha_test

import (
	"fmt"
	"time"

	"github.com/f2xme/gox/captcha"
)

// Example 演示基本的验证码生成和验证
func Example() {
	// 创建内存存储
	store := captcha.NewMemoryStore(1000, 5*time.Minute)

	// 创建验证码实例
	c := captcha.New(store)

	// 生成验证码
	id, b64s, err := c.Generate()
	if err != nil {
		panic(err)
	}

	fmt.Printf("验证码 ID: %s\n", id)
	fmt.Printf("验证码数据长度: %d\n", len(b64s))

	// 验证码验证（实际应用中答案由用户输入）
	// 这里我们无法预知答案，所以只演示 API 调用
	_ = c.Verify(id, "1234")
}

// ExampleNew 演示如何创建验证码实例
func ExampleNew() {
	store := captcha.NewMemoryStore(1000, 5*time.Minute)
	c := captcha.New(store)
	fmt.Printf("验证码实例类型: %T\n", c)
	// Output: 验证码实例类型: *captcha.captchaImpl
}

// ExampleNew_withOptions 演示如何使用选项创建验证码实例
func ExampleNew_withOptions() {
	store := captcha.NewMemoryStore(1000, 5*time.Minute)
	c := captcha.New(store,
		captcha.WithType(captcha.TypeDigit),
		captcha.WithLength(6),
		captcha.WithWidth(300),
		captcha.WithHeight(100),
	)
	fmt.Printf("验证码实例类型: %T\n", c)
	// Output: 验证码实例类型: *captcha.captchaImpl
}

// ExampleNewMemoryStore 演示如何创建内存存储
func ExampleNewMemoryStore() {
	// 保留 1000 个验证码，5 分钟后过期
	store := captcha.NewMemoryStore(1000, 5*time.Minute)
	fmt.Printf("存储类型: %T\n", store)
	// Output: 存储类型: *base64Captcha.memoryStore
}

// ExampleWithType 演示如何设置验证码类型
func ExampleWithType() {
	store := captcha.NewMemoryStore(1000, 5*time.Minute)

	// 数字验证码
	digitCaptcha := captcha.New(store, captcha.WithType(captcha.TypeDigit))
	fmt.Printf("数字验证码: %T\n", digitCaptcha)

	// 字母验证码
	stringCaptcha := captcha.New(store, captcha.WithType(captcha.TypeString))
	fmt.Printf("字母验证码: %T\n", stringCaptcha)

	// 算术验证码
	mathCaptcha := captcha.New(store, captcha.WithType(captcha.TypeMath))
	fmt.Printf("算术验证码: %T\n", mathCaptcha)

	// 音频验证码
	audioCaptcha := captcha.New(store, captcha.WithType(captcha.TypeAudio))
	fmt.Printf("音频验证码: %T\n", audioCaptcha)

	// Output:
	// 数字验证码: *captcha.captchaImpl
	// 字母验证码: *captcha.captchaImpl
	// 算术验证码: *captcha.captchaImpl
	// 音频验证码: *captcha.captchaImpl
}

// ExampleWithLength 演示如何设置验证码长度
func ExampleWithLength() {
	store := captcha.NewMemoryStore(1000, 5*time.Minute)

	// 4 位验证码（默认）
	c4 := captcha.New(store, captcha.WithLength(4))
	fmt.Printf("4 位验证码: %T\n", c4)

	// 6 位验证码
	c6 := captcha.New(store, captcha.WithLength(6))
	fmt.Printf("6 位验证码: %T\n", c6)

	// Output:
	// 4 位验证码: *captcha.captchaImpl
	// 6 位验证码: *captcha.captchaImpl
}

// ExampleWithWidth 演示如何设置验证码宽度
func ExampleWithWidth() {
	store := captcha.NewMemoryStore(1000, 5*time.Minute)
	c := captcha.New(store, captcha.WithWidth(300))
	fmt.Printf("验证码实例: %T\n", c)
	// Output: 验证码实例: *captcha.captchaImpl
}

// ExampleWithHeight 演示如何设置验证码高度
func ExampleWithHeight() {
	store := captcha.NewMemoryStore(1000, 5*time.Minute)
	c := captcha.New(store, captcha.WithHeight(100))
	fmt.Printf("验证码实例: %T\n", c)
	// Output: 验证码实例: *captcha.captchaImpl
}

// ExampleWithNoiseCount 演示如何设置噪点数量
func ExampleWithNoiseCount() {
	store := captcha.NewMemoryStore(1000, 5*time.Minute)
	c := captcha.New(store, captcha.WithNoiseCount(5))
	fmt.Printf("验证码实例: %T\n", c)
	// Output: 验证码实例: *captcha.captchaImpl
}

// ExampleCaptcha_Generate 演示如何生成验证码
func ExampleCaptcha_Generate() {
	store := captcha.NewMemoryStore(1000, 5*time.Minute)
	c := captcha.New(store)

	id, b64s, err := c.Generate()
	if err != nil {
		panic(err)
	}

	fmt.Printf("ID 长度: %d\n", len(id))
	fmt.Printf("数据长度 > 0: %v\n", len(b64s) > 0)
	// Output:
	// ID 长度: 20
	// 数据长度 > 0: true
}

// ExampleCaptcha_Verify 演示如何验证验证码
func ExampleCaptcha_Verify() {
	store := captcha.NewMemoryStore(1000, 5*time.Minute)
	c := captcha.New(store)

	// 生成验证码
	id, _, err := c.Generate()
	if err != nil {
		panic(err)
	}

	// 验证错误答案
	result := c.Verify(id, "wrong")
	fmt.Printf("错误答案验证结果: %v\n", result)

	// Output:
	// 错误答案验证结果: false
}
