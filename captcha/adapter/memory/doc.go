// Package memory 提供验证码内存存储适配器。
//
// 内存适配器适合单进程服务和测试场景，提供轻量级的验证码答案存储、
// 过期清理和容量限制能力。
//
// # 功能特性
//
//   - 内存存储：无需外部依赖，创建后即可使用
//   - 自动过期：支持默认 TTL 和定期清理
//   - 容量限制：支持按写入顺序淘汰超出限制的验证码
//   - 幂等关闭：Close 可以安全重复调用
//
// # 快速开始
//
// 创建内存存储：
//
//	store := memory.New(
//		memory.WithTTL(10*time.Minute),
//		memory.WithCleanupInterval(2*time.Minute),
//		memory.WithMaxSize(10000),
//	)
//
// 创建使用内存存储的验证码服务：
//
//	c, err := memory.NewCaptcha(
//		memory.WithCaptchaType(base64.TypeDigit),
//		memory.WithLength(6),
//	)
//	if err != nil {
//		return err
//	}
//
// # 注意事项
//
//   - 内存存储不会跨进程共享数据
//   - 服务重启后已生成的验证码会丢失
//   - 多实例生产环境建议使用 cache 适配器接入共享缓存
package memory
