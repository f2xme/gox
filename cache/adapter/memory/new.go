package memory

import (
	"log"

	"github.com/f2xme/gox/cache"
	"github.com/f2xme/gox/config"
)

// New 使用给定选项创建一个新的内存缓存。
// 重要：返回的缓存会启动一个后台清理 goroutine。
// 使用完毕后必须调用 Close() 以防止 goroutine 泄漏。
// 使用 defer 确保清理：
//
//	c, _ := memory.New()
//	defer c.Close()
func New(opts ...Option) (cache.Store, error) {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(&cfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	c := &memCache{
		items:  make(map[string]*item),
		cfg:    &cfg,
		stopCh: make(chan struct{}),
	}

	// 如果设置了 maxSize，初始化淘汰策略
	if cfg.MaxSize > 0 {
		switch cfg.EvictionPolicy {
		case "lfu":
			c.eviction = newLFUPolicy()
		default: // "lru"
			c.eviction = newLRUPolicy()
		}
	}

	// 启动清理 goroutine
	go c.cleanupLoop()

	return c, nil
}

// MustNew 创建一个新的内存缓存，出错时终止程序。
func MustNew(opts ...Option) cache.Store {
	c, err := New(opts...)
	if err != nil {
		log.Fatalf("memory: failed to create cache: %v", err)
	}
	return c
}

// NewWithConfig 使用 config.Config 中的配置创建一个新的内存缓存。
// 配置键：
//   - cache.memory.maxSize (int): 最大条目数（0 = 无限制）
//   - cache.memory.cleanupInterval (duration): 清理间隔（默认：1m）
//   - cache.memory.evictionPolicy (string): "lru" 或 "lfu"（默认："lru"）
func NewWithConfig(cfg config.Config) (cache.Store, error) {
	opts := []Option{}

	if maxSize := cfg.GetInt("cache.memory.maxSize"); maxSize > 0 {
		opts = append(opts, WithMaxSize(maxSize))
	} else if maxSize := cfg.GetInt("cache.mem.maxSize"); maxSize > 0 {
		opts = append(opts, WithMaxSize(maxSize))
	}

	if interval := cfg.GetDuration("cache.memory.cleanupInterval"); interval > 0 {
		opts = append(opts, WithCleanupInterval(interval))
	} else if interval := cfg.GetDuration("cache.mem.cleanupInterval"); interval > 0 {
		opts = append(opts, WithCleanupInterval(interval))
	}

	if policy := cfg.GetString("cache.memory.evictionPolicy"); policy != "" {
		opts = append(opts, WithEvictionPolicy(policy))
	} else if policy := cfg.GetString("cache.mem.evictionPolicy"); policy != "" {
		opts = append(opts, WithEvictionPolicy(policy))
	}

	return New(opts...)
}

// MustNewWithConfig 使用配置创建一个新的内存缓存，出错时终止程序。
func MustNewWithConfig(cfg config.Config) cache.Store {
	c, err := NewWithConfig(cfg)
	if err != nil {
		log.Fatalf("memory: failed to create cache from config: %v", err)
	}
	return c
}
