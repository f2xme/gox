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
// Take 必须原子执行，不能用独立的 Get 和 Delete 实现。
type Taker interface {
	// Take 获取并删除验证码答案。
	// 如果验证码不存在或已过期，返回 ErrNotFound。
	Take(ctx context.Context, id string) (string, error)
}

// AtomicStore 提供验证码消费和刷新所需的原子操作。
type AtomicStore interface {
	Store
	Taker

	// CompareAndDelete 仅在未过期的答案等于 expected 时删除，返回是否删除。
	CompareAndDelete(ctx context.Context, id, expected string) (bool, error)

	// CompareAndSwap 仅在未过期的答案等于 expected 时更新，返回是否更新。
	// ttl 为 0 时使用适配器默认过期时间；不存在时不能创建新条目。
	CompareAndSwap(ctx context.Context, id, expected, answer string, ttl time.Duration) (bool, error)
}
