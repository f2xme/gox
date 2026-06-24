package memory

import (
	"log"
	"time"

	"github.com/f2xme/gox/oss"
)

// New 使用给定选项创建新的内存 OSS 存储。
func New(opts ...Option) (*Storage, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	if err := options.validate(); err != nil {
		return nil, oss.WrapError(oss.ErrCodeInvalidArgument, "invalid options", err)
	}

	now := time.Now()
	return &Storage{
		objects: make(map[string]*object),
		buckets: map[string]*oss.Bucket{
			options.BucketName: {
				Name:         options.BucketName,
				CreationDate: now,
			},
		},
		options: options,
	}, nil
}

// MustNew 创建新的内存 OSS 存储，失败时使用 log.Fatalf 退出。
func MustNew(opts ...Option) *Storage {
	storage, err := New(opts...)
	if err != nil {
		log.Fatalf("memory: create storage failed: %v", err)
	}
	return storage
}
