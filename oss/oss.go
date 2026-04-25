package oss

import (
	"context"
	"io"
	"time"
)

// Storage 对象存储统一接口
type Storage interface {
	// Put 上传对象
	Put(ctx context.Context, key string, reader io.Reader, opts ...PutOption) error
	// Get 下载对象
	Get(ctx context.Context, key string, opts ...GetOption) (io.ReadCloser, error)
	// Delete 删除对象
	Delete(ctx context.Context, key string) error
	// Stat 获取对象元信息
	Stat(ctx context.Context, key string) (*ObjectInfo, error)
	// Exists 检查对象是否存在
	Exists(ctx context.Context, key string) (bool, error)
	// List 列出对象
	List(ctx context.Context, opts ...ListOption) (*ListResult, error)
	// SignURL 生成预签名 URL
	SignURL(ctx context.Context, key string, opts ...SignOption) (string, error)
}

// BucketStorage 定义存储桶管理能力
type BucketStorage interface {
	// CreateBucket 创建存储桶
	CreateBucket(ctx context.Context, bucket string, opts ...BucketOption) error
	// DeleteBucket 删除存储桶
	DeleteBucket(ctx context.Context, bucket string) error
	// ListBuckets 列出存储桶
	ListBuckets(ctx context.Context) ([]*Bucket, error)
}

// Object 定义对象列表中的基础信息
type Object struct {
	// Key 对象键
	Key string
	// Size 对象大小，单位字节
	Size int64
	// LastModified 最后修改时间
	LastModified time.Time
	// ETag 对象实体标签
	ETag string
	// ContentType 内容类型
	ContentType string
}

// ObjectInfo 定义对象元信息
type ObjectInfo struct {
	// Key 对象键
	Key string
	// Size 对象大小，单位字节
	Size int64
	// LastModified 最后修改时间
	LastModified time.Time
	// ETag 对象实体标签
	ETag string
	// ContentType 内容类型
	ContentType string
	// Metadata 用户自定义元数据
	Metadata map[string]string
}

// ListResult 定义对象列表结果
type ListResult struct {
	// Objects 对象列表
	Objects []*Object
	// Prefixes 按分隔符折叠出的公共前缀
	Prefixes []string
	// NextToken 下一页令牌，空值表示没有下一页
	NextToken string
	// Truncated 表示结果是否被截断
	Truncated bool
}

// Bucket 定义存储桶信息
type Bucket struct {
	// Name 存储桶名称
	Name string
	// CreationDate 创建时间
	CreationDate time.Time
	// Region 所属地域
	Region string
}
