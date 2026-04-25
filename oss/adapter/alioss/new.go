package alioss

import (
	"log"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"

	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/oss"
)

// New 创建一个新的阿里云 OSS 存储实例
//
// 示例：
//
//	storage, err := alioss.New(
//		alioss.WithEndpoint("oss-cn-hangzhou.aliyuncs.com"),
//		alioss.WithCredentials("access-key-id", "access-key-secret"),
//		alioss.WithBucket("my-bucket"),
//	)
//
// 返回值：
//   - *Storage: 存储实例
//   - error: 创建失败时返回错误
func New(opts ...Option) (*Storage, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	return NewWithOptions(&o)
}

// MustNew 创建一个新的阿里云 OSS 存储实例，失败时终止程序
func MustNew(opts ...Option) *Storage {
	storage, err := New(opts...)
	if err != nil {
		log.Fatalf("alioss: create storage instance failed: %v", err)
	}
	return storage
}

// NewWithOptions 使用 Options 创建一个新的阿里云 OSS 存储实例
func NewWithOptions(opts *Options) (*Storage, error) {
	if opts == nil {
		return nil, oss.NewError(oss.ErrCodeInvalidArgument, "options is required")
	}

	o := *opts
	if err := o.Validate(); err != nil {
		return nil, err
	}

	client, err := aliyunoss.New(o.Endpoint, o.AccessKeyID, o.AccessKeySecret, o.buildClientOptions()...)
	if err != nil {
		return nil, oss.WrapError(oss.ErrCodeInternal, "failed to create client", err)
	}

	bucketHandle, err := client.Bucket(o.Bucket)
	if err != nil {
		return nil, oss.WrapError(oss.ErrCodeInternal, "failed to get bucket", err)
	}

	return &Storage{
		client:       client,
		bucket:       o.Bucket,
		bucketHandle: bucketHandle,
	}, nil
}

// MustNewWithOptions 使用 Options 创建一个新的阿里云 OSS 存储实例，失败时终止程序
func MustNewWithOptions(opts *Options) *Storage {
	storage, err := NewWithOptions(opts)
	if err != nil {
		log.Fatalf("alioss: use options to create storage instance failed: %v", err)
	}
	return storage
}

// NewWithConfig 使用 config.Config 创建一个新的阿里云 OSS 存储实例
//
// 可选 prefix 参数用于自定义配置键前缀，默认值为 "oss.alioss"。
//
// 配置键：
//   - {prefix}.endpoint: OSS 端点地址（必需）
//   - {prefix}.accessKeyID: Access Key ID（必需）
//   - {prefix}.accessKeySecret: Access Key Secret（必需）
//   - {prefix}.bucket: 存储桶名称（必需）
//   - {prefix}.securityToken: STS 安全令牌（可选）
//   - {prefix}.enableCRC: 是否启用 CRC 校验（可选）
//   - {prefix}.timeout: 超时时间，单位秒（可选）
func NewWithConfig(cfg config.Config, prefix ...string) (*Storage, error) {
	if cfg == nil {
		return nil, oss.NewError(oss.ErrCodeInvalidArgument, "config is required")
	}

	p := "oss.alioss"
	if len(prefix) > 0 && prefix[0] != "" {
		p = prefix[0]
	}

	opts := []Option{
		WithEndpoint(cfg.GetString(p + ".endpoint")),
		WithCredentials(
			cfg.GetString(p+".accessKeyID"),
			cfg.GetString(p+".accessKeySecret"),
		),
		WithBucket(cfg.GetString(p + ".bucket")),
	}

	if token := cfg.GetString(p + ".securityToken"); token != "" {
		opts = append(opts, WithSecurityToken(token))
	}
	if cfg.GetBool(p + ".enableCRC") {
		opts = append(opts, WithEnableCRC(true))
	}
	if timeout := cfg.GetInt64(p + ".timeout"); timeout > 0 {
		opts = append(opts, WithTimeout(timeout))
	}

	return New(opts...)
}

// MustNewWithConfig 使用 config.Config 创建一个新的阿里云 OSS 存储实例，失败时终止程序
func MustNewWithConfig(cfg config.Config, prefix ...string) *Storage {
	storage, err := NewWithConfig(cfg, prefix...)
	if err != nil {
		log.Fatalf("alioss: use config to create storage instance failed: %v", err)
	}
	return storage
}
