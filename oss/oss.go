package oss

import (
	"context"
	"io"
	"time"
)

// Storage 对象存储统一接口
type Storage interface {
	// 基础对象操作
	Put(ctx context.Context, key string, reader io.Reader, opts ...PutOption) error
	Get(ctx context.Context, key string, opts ...GetOption) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Head(ctx context.Context, key string) (*ObjectInfo, error)
	Exists(ctx context.Context, key string) (bool, error)

	// 对象列表
	List(ctx context.Context, opts ...ListOption) ([]*Object, error)

	// Bucket 操作
	CreateBucket(ctx context.Context, bucket string) error
	DeleteBucket(ctx context.Context, bucket string) error
	ListBuckets(ctx context.Context) ([]*Bucket, error)

	// 预签名 URL
	PresignedURL(ctx context.Context, key string, opts ...PresignedOption) (string, error)
}

// Object 对象信息
type Object struct {
	Key          string
	Size         int64
	LastModified time.Time
	ETag         string
	ContentType  string
}

// ObjectInfo 对象详细信息
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ETag         string
	ContentType  string
	Metadata     map[string]string
}

// Bucket 存储桶信息
type Bucket struct {
	Name         string
	CreationDate time.Time
	Region       string
}
