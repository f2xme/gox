package redis

import (
	"errors"

	"github.com/redis/go-redis/v9"
)

// Options 存储 Redis 缓存的配置。
type Options struct {
	// Addr 是 Redis 服务器地址。
	// 默认为 "localhost:6379"。
	Addr string
	// Password 是 Redis 认证密码。
	// 默认为空（无认证）。
	Password string
	// DB 是 Redis 数据库编号。
	// 默认为 0。
	DB int
	// Client 是自定义 Redis 客户端。
	// 当提供时，Addr、Password 和 DB 将被忽略。
	Client redis.UniversalClient
}

// defaultOptions 返回默认配置。
func defaultOptions() Options {
	return Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Client:   nil,
	}
}

// Validate 验证配置。
func (o *Options) Validate() error {
	// 如果提供了自定义客户端，跳过验证
	if o.Client != nil {
		return nil
	}

	// 验证 Addr
	if o.Addr == "" {
		return errors.New("redis: addr cannot be empty")
	}

	// 验证 DB（Redis 默认支持 0-15）
	if o.DB < 0 || o.DB > 15 {
		return errors.New("redis: db must be between 0 and 15")
	}

	return nil
}

// Option 配置 Redis 缓存。
type Option func(*Options)

// WithAddr 设置 Redis 服务器地址。
// 默认为 "localhost:6379"。
//
// 示例：
//
//	redis.New(redis.WithAddr("localhost:6379"))
func WithAddr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

// WithPassword 设置 Redis 认证密码。
//
// 示例：
//
//	redis.New(redis.WithPassword("secret"))
func WithPassword(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}

// WithDB 设置 Redis 数据库编号。
// 默认为 0。
//
// 示例：
//
//	redis.New(redis.WithDB(1))
func WithDB(db int) Option {
	return func(o *Options) {
		o.DB = db
	}
}

// WithClient 设置自定义 Redis 客户端。
// 当提供时，其他连接选项（Addr、Password、DB）将被忽略。
//
// 示例：
//
//	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
//	redis.New(redis.WithClient(client))
func WithClient(client redis.UniversalClient) Option {
	return func(o *Options) {
		o.Client = client
	}
}
