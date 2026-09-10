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
//		defer c.Close()
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
// 自定义存储需要实现 AtomicStore，包括原子 Take、CompareAndDelete、CompareAndSwap；
// 缺少能力时 New 返回 ErrAtomicStoreRequired。Get 应返回验证码答案，Delete 应保持幂等，
// 验证码不存在或已过期时返回 ErrNotFound。刷新过程中被消费、删除或更新时，
// Regenerate 返回 ErrNotFound，不会恢复旧验证码。存储生命周期由调用方管理。
//
// # 滑动验证
//
// 使用 NewSlide(store) 创建滑动验证器，store 必须同时实现 Store 和原子 Taker。
// Generate 返回一次性 token 和目标距离；前端提交 SlideVerifyData，包含距离、
// 毫秒时长和 TrackPoint 轨迹。默认检查 300ms–10s 时长、终点容差、轨迹起终点、
// 时间顺序和速度方差；Duration 必须等于轨迹最后一点的 T。
// 验证成功或失败都会消费 token。默认有效期为 5 分钟。
// 示例页面：在仓库根目录运行 go run ./captcha/example/slider，打开 http://127.0.0.1:8080。
//
// 前端将轨道的完整可滑动距离映射到 SlideChallenge.Distance，适配不同屏幕尺寸。
// WithSlideDistance、WithSlideDuration 和 WithSlideMinSpeedVariance 可调整行为规则。
// 轨迹来自客户端，启发式校验不能证明是真人。发送短信等业务操作应在服务端
// Verify 成功后执行；已消费的 token 不能作为另一个业务接口的通行凭证。
//
// # 注意事项
//
//   - New 返回错误，库代码不会直接退出调用方进程
//   - Verify 在答案正确时会自动删除验证码
//   - Regenerate 只能刷新已存在的验证码，空 ID 返回 ErrInvalidID
//   - 内存适配器适合单进程服务，生产多实例部署建议使用共享缓存后端
//   - base64 生成器返回的 data 可直接作为图片或音频的 base64 内容
package captcha
