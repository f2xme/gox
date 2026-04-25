package main

import (
	"context"
	"fmt"
	"time"

	"github.com/f2xme/gox/captcha"
	"github.com/f2xme/gox/captcha/adapter/memory"
	"github.com/f2xme/gox/captcha/generator/base64"
)

func main() {
	fmt.Println("=== captcha 包使用示例 ===")
	ctx := context.Background()

	// 示例 1: 使用便捷构造函数创建数字验证码
	fmt.Println("\n示例 1: 数字验证码（默认）")
	digitCaptcha := memory.NewCaptcha(
		memory.WithCaptchaTTL(5*time.Minute),
		memory.WithGenerator(
			base64.WithType(base64.TypeDigit),
			base64.WithLength(6),
		),
	)

	id1, b64s1, err := digitCaptcha.Generate(ctx)
	if err != nil {
		fmt.Printf("生成失败: %v\n", err)
	} else {
		fmt.Printf("验证码 ID: %s\n", id1)
		if len(b64s1) > 50 {
			fmt.Printf("验证码图片（base64，前 50 字符）: %s...\n", b64s1[:50])
		} else {
			fmt.Printf("验证码图片（base64）: %s\n", b64s1)
		}
		if len(b64s1) > 20 {
			fmt.Printf("提示：在浏览器中可以使用 data:image/png;base64,%s 查看图片\n", b64s1[:20])
		}

		// 模拟用户输入验证码
		fmt.Println("\n验证测试:")
		ok, _ := digitCaptcha.Verify(ctx, id1, "000000")
		fmt.Printf("验证错误答案 '000000': %v\n", ok)
		// 注意：实际使用中，正确答案需要用户查看图片后输入
		// 这里无法验证成功，因为我们不知道生成的验证码内容
	}

	// 示例 2: 字母数字混合验证码
	fmt.Println("\n示例 2: 字母数字混合验证码")
	stringCaptcha := memory.NewCaptcha(
		memory.WithGenerator(
			base64.WithType(base64.TypeString),
			base64.WithLength(5),
			base64.WithSize(300, 100),
		),
	)

	id2, b64s2, err := stringCaptcha.Generate(ctx)
	if err != nil {
		fmt.Printf("生成失败: %v\n", err)
	} else {
		fmt.Printf("验证码 ID: %s\n", id2)
		fmt.Printf("验证码图片（base64，前 50 字符）: %s...\n", b64s2[:50])
	}

	// 示例 3: 算术表达式验证码
	fmt.Println("\n示例 3: 算术表达式验证码")
	mathCaptcha := memory.NewCaptcha(
		memory.WithGenerator(
			base64.WithType(base64.TypeMath),
			base64.WithLength(4),
		),
	)

	id3, b64s3, err := mathCaptcha.Generate(ctx)
	if err != nil {
		fmt.Printf("生成失败: %v\n", err)
	} else {
		fmt.Printf("验证码 ID: %s\n", id3)
		fmt.Printf("验证码图片（base64，前 50 字符）: %s...\n", b64s3[:50])
		fmt.Println("提示：算术验证码显示类似 '1+2=?' 的表达式，用户需要输入计算结果")
	}

	// 示例 4: 音频验证码
	fmt.Println("\n示例 4: 音频验证码")
	audioCaptcha := memory.NewCaptcha(
		memory.WithGenerator(
			base64.WithType(base64.TypeAudio),
			base64.WithLength(6),
			base64.WithLanguage("en"),
		),
	)

	id4, b64s4, err := audioCaptcha.Generate(ctx)
	if err != nil {
		fmt.Printf("生成失败: %v\n", err)
	} else {
		fmt.Printf("验证码 ID: %s\n", id4)
		fmt.Printf("音频数据（base64，前 50 字符）: %s...\n", b64s4[:50])
		fmt.Println("提示：在浏览器中可以使用 data:audio/wav;base64,... 播放音频")
	}

	// 示例 5: 自定义配置（高噪点）
	fmt.Println("\n示例 5: 自定义配置（高噪点）")
	customCaptcha := memory.NewCaptcha(
		memory.WithGenerator(
			base64.WithType(base64.TypeDigit),
			base64.WithLength(4),
			base64.WithSize(200, 60),
			base64.WithNoiseCount(10), // 增加噪点数量
		),
	)

	id5, b64s5, err := customCaptcha.Generate(ctx)
	if err != nil {
		fmt.Printf("生成失败: %v\n", err)
	} else {
		fmt.Printf("验证码 ID: %s\n", id5)
		fmt.Printf("验证码图片（base64，前 50 字符）: %s...\n", b64s5[:50])
	}

	// 示例 6: 验证流程演示
	fmt.Println("\n示例 6: 验证流程演示")
	demoCaptcha := memory.NewCaptcha(
		memory.WithGenerator(
			base64.WithType(base64.TypeDigit),
			base64.WithLength(4),
		),
	)

	demoID, _, _ := demoCaptcha.Generate(ctx)
	fmt.Printf("生成验证码 ID: %s\n", demoID)

	// 模拟验证失败
	ok, _ := demoCaptcha.Verify(ctx, demoID, "0000")
	fmt.Printf("验证错误答案: %v\n", ok)

	// 注意：验证失败后，验证码仍然有效（可以重试）
	// 但验证成功后，验证码会被自动删除，防止重复使用

	// 示例 7: 内存存储配置
	fmt.Println("\n示例 7: 内存存储配置")
	fmt.Println("创建内存存储：保留 1000 个验证码，5 分钟过期")
	// 使用 New() 创建自定义存储，然后传给 captcha.New()
	store := memory.New(
		memory.WithMaxSize(1000),
		memory.WithTTL(5*time.Minute),
	)
	largeStoreCaptcha := captcha.New(store)
	fmt.Printf("验证码实例创建成功: %T\n", largeStoreCaptcha)

	// 示例 8: 批量生成验证码
	fmt.Println("\n示例 8: 批量生成验证码")
	batchCaptcha := memory.NewCaptcha(
		memory.WithGenerator(
			base64.WithType(base64.TypeDigit),
			base64.WithLength(4),
		),
	)

	fmt.Println("批量生成 5 个验证码:")
	for i := 1; i <= 5; i++ {
		id, _, err := batchCaptcha.Generate(ctx)
		if err != nil {
			fmt.Printf("  %d. 生成失败: %v\n", i, err)
		} else {
			fmt.Printf("  %d. ID: %s\n", i, id)
		}
	}

	fmt.Println("\n=== 示例结束 ===")
	fmt.Println("\n提示：")
	fmt.Println("- 验证码图片为 base64 编码的 PNG 格式")
	fmt.Println("- 音频验证码为 base64 编码的 WAV 格式")
	fmt.Println("- 验证成功后验证码会自动删除，防止重复使用")
	fmt.Println("- 验证码过期后无法通过验证")
	fmt.Println("- 在 Web 应用中，可以将 base64 数据直接嵌入 HTML")
	fmt.Println("- 所有操作都支持 context.Context 进行超时控制")
}
