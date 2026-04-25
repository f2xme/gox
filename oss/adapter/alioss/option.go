package alioss

import (
	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/f2xme/gox/oss"
)

// Options 定义阿里云 OSS 配置选项
type Options struct {
	// Endpoint OSS 端点地址
	Endpoint string
	// AccessKeyID Access Key ID
	AccessKeyID string
	// AccessKeySecret Access Key Secret
	AccessKeySecret string
	// Bucket 存储桶名称
	Bucket string
	// SecurityToken STS 安全令牌
	SecurityToken string
	// EnableCRC 是否启用 CRC 校验
	EnableCRC bool
	// Timeout 超时时间，单位秒
	Timeout int64
}

// Option 定义配置选项函数
type Option func(*Options)

// WithEndpoint 设置 OSS 端点地址
//
// 示例：
//
//	alioss.New(alioss.WithEndpoint("oss-cn-hangzhou.aliyuncs.com"))
func WithEndpoint(endpoint string) Option {
	return func(o *Options) {
		o.Endpoint = endpoint
	}
}

// WithCredentials 设置访问凭证
//
// 示例：
//
//	alioss.New(alioss.WithCredentials("access-key-id", "access-key-secret"))
func WithCredentials(accessKeyID, accessKeySecret string) Option {
	return func(o *Options) {
		o.AccessKeyID = accessKeyID
		o.AccessKeySecret = accessKeySecret
	}
}

// WithBucket 设置默认存储桶
//
// 示例：
//
//	alioss.New(alioss.WithBucket("my-bucket"))
func WithBucket(bucket string) Option {
	return func(o *Options) {
		o.Bucket = bucket
	}
}

// WithSecurityToken 设置 STS 安全令牌
//
// 示例：
//
//	alioss.New(
//		alioss.WithCredentials(keyID, keySecret),
//		alioss.WithSecurityToken("STS-token"))
func WithSecurityToken(token string) Option {
	return func(o *Options) {
		o.SecurityToken = token
	}
}

// WithEnableCRC 启用 CRC 校验
//
// 示例：
//
//	alioss.New(alioss.WithEnableCRC(true))
func WithEnableCRC(enable bool) Option {
	return func(o *Options) {
		o.EnableCRC = enable
	}
}

// WithTimeout 设置超时时间（秒）
//
// 示例：
//
//	alioss.New(alioss.WithTimeout(60))
func WithTimeout(timeout int64) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// defaultOptions 返回默认配置选项
func defaultOptions() Options {
	return Options{
		EnableCRC: false,
		Timeout:   0,
	}
}

// Validate 验证配置选项的有效性
func (o *Options) Validate() error {
	if o.Endpoint == "" {
		return oss.NewError(oss.ErrCodeInvalidArgument, "endpoint is required")
	}
	if o.AccessKeyID == "" {
		return oss.NewError(oss.ErrCodeInvalidArgument, "access key id is required")
	}
	if o.AccessKeySecret == "" {
		return oss.NewError(oss.ErrCodeInvalidArgument, "access key secret is required")
	}
	if o.Bucket == "" {
		return oss.NewError(oss.ErrCodeInvalidArgument, "bucket is required")
	}
	return nil
}

// buildClientOptions 构建阿里云 SDK 客户端选项
func (o *Options) buildClientOptions() []aliyunoss.ClientOption {
	var opts []aliyunoss.ClientOption

	if o.SecurityToken != "" {
		opts = append(opts, aliyunoss.SecurityToken(o.SecurityToken))
	}

	if o.EnableCRC {
		opts = append(opts, aliyunoss.EnableCRC(true))
	}

	if o.Timeout > 0 {
		opts = append(opts, aliyunoss.Timeout(o.Timeout, o.Timeout))
	}

	return opts
}
