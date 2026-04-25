package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/f2xme/gox/captcha"
	"github.com/f2xme/gox/captcha/adapter/memory"
	"github.com/f2xme/gox/captcha/generator/base64"
)

func main() {
	fmt.Println("=== captcha 包使用示例 ===")
	ctx := context.Background()
	mustCaptcha := func(c captcha.Service, err error) captcha.Service {
		if err != nil {
			log.Fatalf("创建验证码实例失败: %v", err)
		}
		return c
	}
	printfln := func(format string, args ...any) {
		fmt.Printf(format, args...)
		fmt.Println()
	}

	// 示例 1: 使用便捷构造函数创建数字验证码
	fmt.Println()
	fmt.Println("示例 1: 数字验证码（默认）")
	digitCaptcha := mustCaptcha(memory.NewCaptcha(
		memory.WithCaptchaTTL(5*time.Minute),
		memory.WithGenerator(
			base64.WithType(base64.TypeDigit),
			base64.WithLength(6),
		),
	))

	challenge1, err := digitCaptcha.Generate(ctx)
	b64s1 := challenge1.Data
	if err != nil {
		printfln("生成失败: %v", err)
	} else {
		printfln("验证码 ID: %s", challenge1.ID)
		if len(b64s1) > 50 {
			printfln("验证码图片（base64，前 50 字符）: %s...", b64s1[:50])
		} else {
			printfln("验证码图片（base64）: %s", b64s1)
		}
		if len(b64s1) > 20 {
			printfln("提示：在浏览器中可以使用 data:image/png;base64,%s 查看图片", b64s1[:20])
		}

		// 模拟用户输入验证码
		fmt.Println()
		fmt.Println("验证测试:")
		ok, _ := digitCaptcha.Verify(ctx, challenge1.ID, "000000")
		printfln("验证错误答案 '000000': %v", ok)
		// 注意：实际使用中，正确答案需要用户查看图片后输入
		// 这里无法验证成功，因为我们不知道生成的验证码内容
	}

	// 示例 2: 字母数字混合验证码
	fmt.Println()
	fmt.Println("示例 2: 字母数字混合验证码")
	stringCaptcha := mustCaptcha(memory.NewCaptcha(
		memory.WithGenerator(
			base64.WithType(base64.TypeString),
			base64.WithLength(5),
			base64.WithSize(300, 100),
		),
	))

	challenge2, err := stringCaptcha.Generate(ctx)
	b64s2 := challenge2.Data
	if err != nil {
		printfln("生成失败: %v", err)
	} else {
		printfln("验证码 ID: %s", challenge2.ID)
		printfln("验证码图片（base64，前 50 字符）: %s...", b64s2[:50])
	}

	// 示例 3: 算术表达式验证码
	fmt.Println()
	fmt.Println("示例 3: 算术表达式验证码")
	mathCaptcha := mustCaptcha(memory.NewCaptcha(
		memory.WithGenerator(
			base64.WithType(base64.TypeMath),
			base64.WithLength(4),
		),
	))

	challenge3, err := mathCaptcha.Generate(ctx)
	b64s3 := challenge3.Data
	if err != nil {
		printfln("生成失败: %v", err)
	} else {
		printfln("验证码 ID: %s", challenge3.ID)
		printfln("验证码图片（base64，前 50 字符）: %s...", b64s3[:50])
		fmt.Println("提示：算术验证码显示类似 '1+2=?' 的表达式，用户需要输入计算结果")
	}

	// 示例 4: 音频验证码
	fmt.Println()
	fmt.Println("示例 4: 音频验证码")
	audioCaptcha := mustCaptcha(memory.NewCaptcha(
		memory.WithGenerator(
			base64.WithType(base64.TypeAudio),
			base64.WithLength(6),
			base64.WithLanguage("en"),
		),
	))

	challenge4, err := audioCaptcha.Generate(ctx)
	b64s4 := challenge4.Data
	if err != nil {
		printfln("生成失败: %v", err)
	} else {
		printfln("验证码 ID: %s", challenge4.ID)
		printfln("音频数据（base64，前 50 字符）: %s...", b64s4[:50])
		fmt.Println("提示：在浏览器中可以使用 data:audio/wav;base64,... 播放音频")
	}

	// 示例 5: 自定义配置（高噪点）
	fmt.Println()
	fmt.Println("示例 5: 自定义配置（高噪点）")
	customCaptcha := mustCaptcha(memory.NewCaptcha(
		memory.WithGenerator(
			base64.WithType(base64.TypeDigit),
			base64.WithLength(4),
			base64.WithSize(200, 60),
			base64.WithNoiseCount(10), // 增加噪点数量
		),
	))

	challenge5, err := customCaptcha.Generate(ctx)
	b64s5 := challenge5.Data
	if err != nil {
		printfln("生成失败: %v", err)
	} else {
		printfln("验证码 ID: %s", challenge5.ID)
		printfln("验证码图片（base64，前 50 字符）: %s...", b64s5[:50])
	}

	// 示例 6: 验证流程演示
	fmt.Println()
	fmt.Println("示例 6: 验证流程演示")
	demoCaptcha := mustCaptcha(memory.NewCaptcha(
		memory.WithGenerator(
			base64.WithType(base64.TypeDigit),
			base64.WithLength(4),
		),
	))

	demoChallenge, _ := demoCaptcha.Generate(ctx)
	printfln("生成验证码 ID: %s", demoChallenge.ID)

	// 模拟验证失败
	ok, _ := demoCaptcha.Verify(ctx, demoChallenge.ID, "0000")
	printfln("验证错误答案: %v", ok)

	// 注意：验证失败后，验证码仍然有效（可以重试）
	// 但验证成功后，验证码会被自动删除，防止重复使用

	// 示例 7: 内存存储配置
	fmt.Println()
	fmt.Println("示例 7: 内存存储配置")
	fmt.Println("创建内存存储：保留 1000 个验证码，5 分钟过期")
	// 使用 New() 创建自定义存储，然后传给 captcha.New()
	store := memory.New(
		memory.WithMaxSize(1000),
		memory.WithTTL(5*time.Minute),
	)
	gen, err := base64.New(
		base64.WithType(base64.TypeDigit),
		base64.WithLength(4),
	)
	if err != nil {
		log.Fatalf("创建生成器失败: %v", err)
	}
	largeStoreCaptcha := mustCaptcha(captcha.New(store, captcha.WithGenerator(gen)))
	printfln("验证码实例创建成功: %T", largeStoreCaptcha)

	// 示例 8: 批量生成验证码
	fmt.Println()
	fmt.Println("示例 8: 批量生成验证码")
	batchCaptcha := mustCaptcha(memory.NewCaptcha(
		memory.WithGenerator(
			base64.WithType(base64.TypeDigit),
			base64.WithLength(4),
		),
	))

	fmt.Println("批量生成 5 个验证码:")
	for i := 1; i <= 5; i++ {
		challenge, err := batchCaptcha.Generate(ctx)
		if err != nil {
			printfln("  %d. 生成失败: %v", i, err)
		} else {
			printfln("  %d. ID: %s", i, challenge.ID)
		}
	}

	fmt.Println()
	fmt.Println("=== 示例结束 ===")
	fmt.Println()
	fmt.Println("提示：")
	fmt.Println("- 验证码图片为 base64 编码的 PNG 格式")
	fmt.Println("- 音频验证码为 base64 编码的 WAV 格式")
	fmt.Println("- 验证成功后验证码会自动删除，防止重复使用")
	fmt.Println("- 验证码过期后无法通过验证")
	fmt.Println("- 在 Web 应用中，可以将 base64 数据直接嵌入 HTML")
	fmt.Println("- 所有操作都支持 context.Context 进行超时控制")
}
