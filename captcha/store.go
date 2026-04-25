package captcha

import (
	"context"
	"time"
)

// Store 定义验证码存储接口。
// 所有适配器必须实现此接口以提供验证码的存储、获取和删除功能。
type Store interface {
	// Set 存储验证码答案。
	// ttl 为 0 表示使用适配器的默认过期时间。
	Set(ctx context.Context, id string, answer string, ttl time.Duration) error

	// Get 获取验证码答案。
	// 如果验证码不存在或已过期，返回 ErrNotFound。
	Get(ctx context.Context, id string) (string, error)

	// Delete 删除验证码。
	// 如果验证码不存在不返回错误（幂等操作）。
	Delete(ctx context.Context, id string) error
}

// Taker 定义获取并删除验证码答案的可选能力。
// 存储适配器实现此接口后，ConsumeAlways 策略会优先使用原子消费。
type Taker interface {
	// Take 获取并删除验证码答案。
	// 如果验证码不存在或已过期，返回 ErrNotFound。
	Take(ctx context.Context, id string) (string, error)
}
