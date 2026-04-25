package redis

import (
	"log"

	"github.com/f2xme/gox/cache"
	"github.com/f2xme/gox/config"
	"github.com/redis/go-redis/v9"
)

// New 使用给定选项创建一个新的 Redis 缓存。
// 默认配置：localhost:6379，无密码，DB 0。
//
// 示例：
//
//	c, err := redis.New(
//		redis.WithAddr("localhost:6379"),
//		redis.WithPassword("secret"),
//		redis.WithDB(1),
//	)
func New(opts ...Option) (cache.Store, error) {
	cfg := defaultOptions()

	for _, opt := range opts {
		opt(&cfg)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// 如果提供了自定义客户端则使用它，否则创建一个新的
	var client redis.UniversalClient
	if cfg.Client != nil {
		client = cfg.Client
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		})
	}

	return &redisCache{
		client: client,
	}, nil
}

// MustNew 创建一个新的 Redis 缓存，出错时终止程序。
func MustNew(opts ...Option) cache.Store {
	c, err := New(opts...)
	if err != nil {
		log.Fatalf("redis: failed to create cache: %v", err)
	}
	return c
}

// NewWithConfig 使用 config.Config 中的配置创建一个新的 Redis 缓存。
// prefix 是可选配置前缀，默认使用 "cache"。
// 配置键：
//   - cache.redis.addr (string): Redis 服务器地址（默认："localhost:6379"）
//   - cache.redis.password (string): Redis 认证密码
//   - cache.redis.db (int): Redis 数据库编号（默认：0）
//   - <prefix>.redis.addr (string): 自定义前缀下的 Redis 服务器地址
//   - <prefix>.redis.password (string): 自定义前缀下的 Redis 认证密码
//   - <prefix>.redis.db (int): 自定义前缀下的 Redis 数据库编号
//
// 示例：
//
//	c, err := redis.NewWithConfig(cfg)
//	c, err := redis.NewWithConfig(cfg, "app")
func NewWithConfig(cfg config.Config, prefix ...string) (cache.Store, error) {
	opts := []Option{}
	configPrefix := "cache"
	if len(prefix) > 0 && prefix[0] != "" {
		configPrefix = prefix[0]
	}

	if addr := cfg.GetString(configPrefix + ".redis.addr"); addr != "" {
		opts = append(opts, WithAddr(addr))
	}

	if password := cfg.GetString(configPrefix + ".redis.password"); password != "" {
		opts = append(opts, WithPassword(password))
	}

	if db := cfg.GetInt(configPrefix + ".redis.db"); db > 0 {
		opts = append(opts, WithDB(db))
	}

	return New(opts...)
}

// MustNewWithConfig 使用配置创建一个新的 Redis 缓存，出错时终止程序。
func MustNewWithConfig(cfg config.Config, prefix ...string) cache.Store {
	c, err := NewWithConfig(cfg, prefix...)
	if err != nil {
		log.Fatalf("redis: failed to create cache from config: %v", err)
	}
	return c
}
