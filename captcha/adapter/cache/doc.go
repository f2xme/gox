// Package cache 提供基于 gox/cache 的验证码存储适配器。
//
// cache 适配器适合需要共享验证码状态的场景，可以复用 gox/cache 的内存、
// Redis 等后端，实现多实例服务之间的验证码共享。
//
// # 功能特性
//
//   - 复用 gox/cache：支持多种缓存后端
//   - 支持 key 前缀：便于隔离不同业务的验证码数据
//   - 支持默认 TTL：未传入过期时间时使用适配器默认值
//   - 实现 captcha.Store：可直接传入 captcha.New
//
// # 快速开始
//
// 创建 cache 存储适配器：
//
//	c, err := memory.New()
//	if err != nil {
//		return err
//	}
//
//	store := cacheadapter.New(c,
//		cacheadapter.WithTTL(10*time.Minute),
//		cacheadapter.WithPrefix("mycaptcha:"),
//	)
//
// 创建使用 cache 存储的验证码服务：
//
//	captcha, err := cacheadapter.NewCaptcha(c,
//		cacheadapter.WithCaptchaType(base64.TypeString),
//		cacheadapter.WithLength(4),
//	)
//	if err != nil {
//		return err
//	}
//
// # 注意事项
//
//   - 底层缓存的 Delete 需要保持幂等
//   - 生产环境建议为验证码设置独立 key 前缀
//   - 分布式场景下建议选择支持共享状态的缓存后端
package cache
