package cache

import (
	"context"
	"time"
)

// Typed 提供类型安全的缓存操作包装器
type Typed[T any] struct {
	cache      Cache
	serializer Serializer
}

// TypedOption 是 Typed 的配置选项函数
type TypedOption func(*typedConfig)

// typedConfig 存储 Typed 的配置
type typedConfig struct {
	serializer Serializer
}

// WithSerializer 设置自定义序列化器
func WithSerializer(s Serializer) TypedOption {
	return func(c *typedConfig) {
		c.serializer = s
	}
}

// NewTyped 创建一个新的类型安全缓存包装器
// 默认使用 JSONSerializer
func NewTyped[T any](cache Cache, opts ...TypedOption) *Typed[T] {
	cfg := &typedConfig{
		serializer: JSONSerializer,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Typed[T]{
		cache:      cache,
		serializer: cfg.serializer,
	}
}

// Get 获取并反序列化值
// 如果键不存在，返回 ErrNotFound
func (t *Typed[T]) Get(ctx context.Context, key string) (T, error) {
	var zero T

	data, err := t.cache.Get(ctx, key)
	if err != nil {
		return zero, err
	}

	var value T
	if err := t.serializer.Unmarshal(data, &value); err != nil {
		return zero, err
	}

	return value, nil
}

// Set 序列化并存储值
func (t *Typed[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	data, err := t.serializer.Marshal(value)
	if err != nil {
		return err
	}

	return t.cache.Set(ctx, key, data, ttl)
}

// Delete 删除键
func (t *Typed[T]) Delete(ctx context.Context, key string) error {
	return t.cache.Delete(ctx, key)
}

// Exists 检查键是否存在
func (t *Typed[T]) Exists(ctx context.Context, key string) (bool, error) {
	return t.cache.Exists(ctx, key)
}

// GetOrSet 实现 cache-aside 模式
// 首先尝试获取值，如果不存在则调用 fn 计算值并存储
// 这可以防止缓存击穿
func (t *Typed[T]) GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() (T, error)) (T, error) {
	var zero T

	// 尝试从缓存获取
	value, err := t.Get(ctx, key)
	if err == nil {
		// 缓存命中
		return value, nil
	}

	// 如果不是 ErrNotFound，返回错误
	if err != ErrNotFound {
		return zero, err
	}

	// 缓存未命中，调用加载函数
	value, err = fn()
	if err != nil {
		return zero, err
	}

	// 存储到缓存（失败不影响返回值，因为数据已成功加载）
	_ = t.Set(ctx, key, value, ttl)

	return value, nil
}
