package main

import (
	"fmt"
	"time"

	"github.com/f2xme/gox/captcha"
)

func main() {
	fmt.Println("=== captcha 包使用示例 ===")

	// 示例 1: 数字验证码（默认）
	fmt.Println("\n示例 1: 数字验证码")
	store := captcha.NewMemoryStore(100, 5*time.Minute)
	digitCaptcha := captcha.New(store,
		captcha.WithType(captcha.TypeDigit),
		captcha.WithLength(6),
	)

	id1, b64s1, err := digitCaptcha.Generate()
	if err != nil {
		fmt.Printf("生成失败: %v\n", err)
	} else {
		fmt.Printf("验证码 ID: %s\n", id1)
		fmt.Printf("验证码图片（base64，前 50 字符）: %s...\n", b64s1[:50])
		fmt.Printf("提示：在浏览器中可以使用 data:image/png;base64,%s 查看图片\n", b64s1[:20])

		// 模拟用户输入验证码
		fmt.Println("\n验证测试:")
		fmt.Printf("验证错误答案 '000000': %v\n", digitCaptcha.Verify(id1, "000000"))
		// 注意：实际使用中，正确答案需要用户查看图片后输入
		// 这里无法验证成功，因为我们不知道生成的验证码内容
	}

	// 示例 2: 字母数字混合验证码
	fmt.Println("\n示例 2: 字母数字混合验证码")
	stringCaptcha := captcha.New(store,
		captcha.WithType(captcha.TypeString),
		captcha.WithLength(5),
		captcha.WithWidth(300),
		captcha.WithHeight(100),
	)

	id2, b64s2, err := stringCaptcha.Generate()
	if err != nil {
		fmt.Printf("生成失败: %v\n", err)
	} else {
		fmt.Printf("验证码 ID: %s\n", id2)
		fmt.Printf("验证码图片（base64，前 50 字符）: %s...\n", b64s2[:50])
	}

	// 示例 3: 算术表达式验证码
	fmt.Println("\n示例 3: 算术表达式验证码")
	mathCaptcha := captcha.New(store,
		captcha.WithType(captcha.TypeMath),
		captcha.WithLength(4),
	)

	id3, b64s3, err := mathCaptcha.Generate()
	if err != nil {
		fmt.Printf("生成失败: %v\n", err)
	} else {
		fmt.Printf("验证码 ID: %s\n", id3)
		fmt.Printf("验证码图片（base64，前 50 字符）: %s...\n", b64s3[:50])
		fmt.Println("提示：算术验证码显示类似 '1+2=?' 的表达式，用户需要输入计算结果")
	}

	// 示例 4: 音频验证码
	fmt.Println("\n示例 4: 音频验证码")
	audioCaptcha := captcha.New(store,
		captcha.WithType(captcha.TypeAudio),
		captcha.WithLength(6),
		captcha.WithLanguage("en"),
	)

	id4, b64s4, err := audioCaptcha.Generate()
	if err != nil {
		fmt.Printf("生成失败: %v\n", err)
	} else {
		fmt.Printf("验证码 ID: %s\n", id4)
		fmt.Printf("音频数据（base64，前 50 字符）: %s...\n", b64s4[:50])
		fmt.Println("提示：在浏览器中可以使用 data:audio/wav;base64,... 播放音频")
	}

	// 示例 5: 自定义配置
	fmt.Println("\n示例 5: 自定义配置（高噪点）")
	customCaptcha := captcha.New(store,
		captcha.WithType(captcha.TypeDigit),
		captcha.WithLength(4),
		captcha.WithWidth(200),
		captcha.WithHeight(60),
		captcha.WithNoiseCount(10), // 增加噪点数量
	)

	id5, b64s5, err := customCaptcha.Generate()
	if err != nil {
		fmt.Printf("生成失败: %v\n", err)
	} else {
		fmt.Printf("验证码 ID: %s\n", id5)
		fmt.Printf("验证码图片（base64，前 50 字符）: %s...\n", b64s5[:50])
	}

	// 示例 6: 验证流程演示
	fmt.Println("\n示例 6: 验证流程演示")
	demoCaptcha := captcha.New(store,
		captcha.WithType(captcha.TypeDigit),
		captcha.WithLength(4),
	)

	demoID, _, _ := demoCaptcha.Generate()
	fmt.Printf("生成验证码 ID: %s\n", demoID)

	// 模拟验证失败
	fmt.Printf("验证错误答案: %v\n", demoCaptcha.Verify(demoID, "0000"))

	// 注意：验证失败后，验证码仍然有效（可以重试）
	// 但验证成功后，验证码会被自动删除，防止重复使用

	// 示例 7: 内存存储配置
	fmt.Println("\n示例 7: 内存存储配置")
	fmt.Println("创建内存存储：保留 1000 个验证码，5 分钟过期")
	largeStore := captcha.NewMemoryStore(1000, 5*time.Minute)
	fmt.Printf("存储实例创建成功: %T\n", largeStore)

	// 示例 8: 批量生成验证码
	fmt.Println("\n示例 8: 批量生成验证码")
	batchCaptcha := captcha.New(store,
		captcha.WithType(captcha.TypeDigit),
		captcha.WithLength(4),
	)

	fmt.Println("批量生成 5 个验证码:")
	for i := 1; i <= 5; i++ {
		id, _, err := batchCaptcha.Generate()
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
}
