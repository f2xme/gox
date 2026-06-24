package memory

import (
	"errors"
	"strings"
)

// Option 定义内存 OSS 适配器选项函数。
type Option func(*Options)

// Options 定义内存 OSS 适配器配置。
type Options struct {
	// BucketName 默认对象操作使用的存储桶名称。
	BucketName string
	// SignURLBase 生成测试预签名 URL 时使用的基础地址。
	SignURLBase string
}

func defaultOptions() Options {
	return Options{
		BucketName:  "memory",
		SignURLBase: "memory://oss",
	}
}

func (o Options) validate() error {
	if strings.TrimSpace(o.BucketName) == "" {
		return errors.New("memory: bucket name is required")
	}
	if strings.TrimSpace(o.SignURLBase) == "" {
		return errors.New("memory: sign url base is required")
	}
	return nil
}

// WithBucketName 设置默认对象操作使用的存储桶名称。
//
// 示例：
//
//	storage, _ := memory.New(memory.WithBucketName("test-bucket"))
func WithBucketName(bucket string) Option {
	return func(o *Options) {
		o.BucketName = bucket
	}
}

// WithSignURLBase 设置测试预签名 URL 的基础地址。
//
// 示例：
//
//	storage, _ := memory.New(memory.WithSignURLBase("https://example.test/oss"))
func WithSignURLBase(base string) Option {
	return func(o *Options) {
		o.SignURLBase = strings.TrimRight(base, "/")
	}
}
