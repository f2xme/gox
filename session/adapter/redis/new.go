package redis

import (
	"fmt"

	"github.com/f2xme/gox/session"
	goredis "github.com/redis/go-redis/v9"
)

// New 创建 Redis 会话存储。
func New(opts ...Option) (session.Store, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	if err := options.validate(); err != nil {
		return nil, err
	}

	var client goredis.UniversalClient
	ownsClient := false
	if options.Client != nil {
		client = options.Client
	} else {
		client = goredis.NewClient(&goredis.Options{
			Addr:     options.Addr,
			Password: options.Password,
			DB:       options.DB,
		})
		ownsClient = true
	}

	return &Store{
		client:     client,
		keyPrefix:  options.KeyPrefix,
		ownsClient: ownsClient,
	}, nil
}

// MustNew 创建 Redis 会话存储，失败时 panic。
func MustNew(opts ...Option) session.Store {
	store, err := New(opts...)
	if err != nil {
		panic(fmt.Errorf("redis: failed to create session store: %w", err))
	}
	return store
}
