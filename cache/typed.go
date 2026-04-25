package cache

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"
)

// Typed 提供类型安全的缓存操作包装器
type Typed[T any] struct {
	cache           Store
	serializer      Serializer
	group           singleflight.Group
	ignoreSetErrors bool
}

// TypedOption 是 Typed 的配置选项函数
type TypedOption func(*typedConfig)

// typedConfig 存储 Typed 的配置
type typedConfig struct {
	serializer      Serializer
	ignoreSetErrors bool
}

// WithSerializer 设置自定义序列化器
func WithSerializer(s Serializer) TypedOption {
	return func(c *typedConfig) {
		c.serializer = s
	}
}

// WithIgnoreSetErrors 设置 GetOrLoad 在加载成功但缓存写入失败时仍返回加载值。
func WithIgnoreSetErrors() TypedOption {
	return func(c *typedConfig) {
		c.ignoreSetErrors = true
	}
}

// NewTyped 创建一个新的类型安全缓存包装器
// 默认使用 JSONSerializer
func NewTyped[T any](cache Store, opts ...TypedOption) *Typed[T] {
	cfg := &typedConfig{
		serializer: JSONSerializer,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Typed[T]{
		cache:           cache,
		serializer:      cfg.serializer,
		ignoreSetErrors: cfg.ignoreSetErrors,
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

// GetOrLoad 实现 cache-aside 模式
// 首先尝试获取值，如果不存在则调用 fn 计算值并存储
// 使用 singleflight 防止缓存击穿
func (t *Typed[T]) GetOrLoad(ctx context.Context, key string, ttl time.Duration, fn func(context.Context) (T, error)) (T, error) {
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

	// 缓存未命中，使用 singleflight 防止并发重复计算
	result, err, _ := t.group.Do(key, func() (any, error) {
		// 再次检查缓存（可能其他 goroutine 已经设置）
		value, err := t.Get(ctx, key)
		if err == nil {
			return value, nil
		}
		if err != ErrNotFound {
			return zero, err
		}

		// 调用加载函数
		value, err = fn(ctx)
		if err != nil {
			return zero, err
		}

		if err := t.Set(ctx, key, value, ttl); err != nil && !t.ignoreSetErrors {
			return zero, err
		}

		return value, nil
	})

	if err != nil {
		return zero, err
	}

	return result.(T), nil
}

// GetMany 批量获取多个键的值
// 如果底层 cache 实现了 BatchStore 接口，则使用批量操作
// 否则降级为循环调用 Get
func (t *Typed[T]) GetMany(ctx context.Context, keys []string) (map[string]T, error) {
	result := make(map[string]T, len(keys))

	// 尝试使用批量接口
	if mc, ok := t.cache.(BatchStore); ok {
		dataMap, err := mc.GetMany(ctx, keys)
		if err != nil {
			return nil, err
		}

		for key, data := range dataMap {
			var value T
			if err := t.serializer.Unmarshal(data, &value); err != nil {
				return nil, fmt.Errorf("unmarshal key %s: %w", key, err)
			}
			result[key] = value
		}
		return result, nil
	}

	// 降级为循环调用
	for _, key := range keys {
		value, err := t.Get(ctx, key)
		if err != nil {
			if err == ErrNotFound {
				continue // 跳过不存在的键
			}
			return nil, fmt.Errorf("get key %s: %w", key, err)
		}
		result[key] = value
	}

	return result, nil
}

// SetMany 批量设置多个键值对
// 如果底层 cache 实现了 BatchStore 接口，则使用批量操作
// 否则降级为循环调用 Set
func (t *Typed[T]) SetMany(ctx context.Context, items map[string]T, ttl time.Duration) error {
	// 序列化所有值
	dataMap := make(map[string][]byte, len(items))
	for key, value := range items {
		data, err := t.serializer.Marshal(value)
		if err != nil {
			return fmt.Errorf("marshal key %s: %w", key, err)
		}
		dataMap[key] = data
	}

	// 尝试使用批量接口
	if mc, ok := t.cache.(BatchStore); ok {
		return mc.SetMany(ctx, dataMap, ttl)
	}

	// 降级为循环调用
	for key, data := range dataMap {
		if err := t.cache.Set(ctx, key, data, ttl); err != nil {
			return fmt.Errorf("set key %s: %w", key, err)
		}
	}

	return nil
}

// DeleteMany 批量删除多个键
// 如果底层 cache 实现了 BatchStore 接口，则使用批量操作
// 否则降级为循环调用 Delete
func (t *Typed[T]) DeleteMany(ctx context.Context, keys []string) error {
	// 尝试使用批量接口
	if mc, ok := t.cache.(BatchStore); ok {
		return mc.DeleteMany(ctx, keys)
	}

	// 降级为循环调用
	for _, key := range keys {
		if err := t.cache.Delete(ctx, key); err != nil {
			return fmt.Errorf("delete key %s: %w", key, err)
		}
	}

	return nil
}
