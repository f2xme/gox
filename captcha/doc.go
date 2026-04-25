// Package captcha 提供验证码生成、存储、验证和消费能力。
//
// captcha 包只定义验证码服务的核心流程和存储接口。图片、音频等验证码内容
// 由 generator 子包生成，内存、缓存等存储后端由 adapter 子包提供。
//
// # 功能特性
//
//   - 支持 context.Context，便于接入 Web 请求超时和取消
//   - 支持自定义验证码生成器和存储后端
//   - 默认验证成功后自动删除验证码，防止重复使用
//   - 支持重新生成验证码内容并保持原 ID 不变
//   - 支持可配置过期时间和随机 ID 长度
//
// # 快速开始
//
// 使用内存适配器创建验证码服务：
//
//	package main
//
//	import (
//		"context"
//		"fmt"
//		"log"
//
//		"github.com/f2xme/gox/captcha/adapter/memory"
//		"github.com/f2xme/gox/captcha/generator/base64"
//	)
//
//	func main() {
//		ctx := context.Background()
//
//		c, err := memory.NewCaptcha(
//			memory.WithCaptchaType(base64.TypeDigit),
//			memory.WithLength(6),
//		)
//		if err != nil {
//			log.Fatalf("创建验证码服务失败: %v", err)
//		}
//
//		challenge, err := c.Generate(ctx)
//		if err != nil {
//			log.Fatalf("生成验证码失败: %v", err)
//		}
//
//		fmt.Println("验证码 ID:", challenge.ID)
//		fmt.Println("验证码数据:", challenge.Data)
//
//		ok, err := c.Verify(ctx, challenge.ID, "用户输入")
//		if err != nil {
//			log.Fatalf("验证验证码失败: %v", err)
//		}
//		fmt.Println("验证结果:", ok)
//	}
//
// # 自定义存储
//
// 可以直接组合 Store 和 Generator 创建验证码服务：
//
//	store := memory.New(memory.WithMaxSize(1000))
//	gen, err := base64.New(base64.WithLength(4))
//	if err != nil {
//		return err
//	}
//
//	c, err := captcha.New(store, captcha.WithGenerator(gen))
//	if err != nil {
//		return err
//	}
//
// 自定义存储需要实现 Store 接口。Get 应返回验证码答案，Delete 应保持幂等，
// 验证码不存在或已过期时返回 ErrNotFound。需要原子消费能力时可以额外实现 Taker。
//
// # 注意事项
//
//   - New 返回错误，库代码不会直接退出调用方进程
//   - Verify 在答案正确时会自动删除验证码
//   - Regenerate 只能刷新已存在的验证码，空 ID 返回 ErrInvalidID
//   - 内存适配器适合单进程服务，生产多实例部署建议使用共享缓存后端
//   - base64 生成器返回的 data 可直接作为图片或音频的 base64 内容
package captcha
