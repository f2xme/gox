package redis

import (
	"errors"

	"github.com/redis/go-redis/v9"
)

// Options 定义 Redis 会话存储配置。
type Options struct {
	// Addr 是 Redis 服务器地址。
	// 默认为 "localhost:6379"。
	Addr string
	// Password 是 Redis 认证密码。
	Password string
	// DB 是 Redis 数据库编号。
	DB int
	// KeyPrefix 是 Redis key 前缀。
	// 默认 "session:"。
	KeyPrefix string
	// Client 是自定义 Redis 客户端。
	// 提供后 Addr、Password 和 DB 将被忽略。
	Client redis.UniversalClient
}

// Option 定义 Redis 会话存储配置函数。
type Option func(*Options)

func defaultOptions() Options {
	return Options{
		Addr:      "localhost:6379",
		KeyPrefix: "session:",
	}
}

func (o Options) validate() error {
	if o.Client != nil {
		return nil
	}
	if o.Addr == "" {
		return errors.New("redis: addr cannot be empty")
	}
	if o.DB < 0 || o.DB > 15 {
		return errors.New("redis: db must be between 0 and 15")
	}
	return nil
}

// WithAddr 设置 Redis 服务器地址。
func WithAddr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

// WithPassword 设置 Redis 认证密码。
func WithPassword(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}

// WithDB 设置 Redis 数据库编号。
func WithDB(db int) Option {
	return func(o *Options) {
		o.DB = db
	}
}

// WithKeyPrefix 设置 Redis key 前缀。
func WithKeyPrefix(prefix string) Option {
	return func(o *Options) {
		o.KeyPrefix = prefix
	}
}

// WithClient 设置自定义 Redis 客户端。
func WithClient(client redis.UniversalClient) Option {
	return func(o *Options) {
		o.Client = client
	}
}
