package alioss

import (
	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/oss"
)

// Options 定义阿里云 OSS 配置选项
type Options struct {
	Endpoint        string // OSS 端点地址
	AccessKeyID     string // Access Key ID
	AccessKeySecret string // Access Key Secret
	Bucket          string // 存储桶名称
	SecurityToken   string // STS 安全令牌（可选）
	EnableCRC       bool   // 是否启用 CRC 校验
	Timeout         int64  // 超时时间（秒）
}

// Option 定义配置选项函数
type Option func(*Options)

// WithSecurityToken 设置 STS 安全令牌
//
// 示例：
//
//	alioss.New(endpoint, keyID, keySecret, bucket,
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
//	alioss.New(endpoint, keyID, keySecret, bucket,
//		alioss.WithEnableCRC(true))
func WithEnableCRC(enable bool) Option {
	return func(o *Options) {
		o.EnableCRC = enable
	}
}

// WithTimeout 设置超时时间（秒）
//
// 示例：
//
//	alioss.New(endpoint, keyID, keySecret, bucket,
//		alioss.WithTimeout(60))
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
func (o *Options) buildClientOptions() []alioss.ClientOption {
	var opts []alioss.ClientOption

	if o.SecurityToken != "" {
		opts = append(opts, alioss.SecurityToken(o.SecurityToken))
	}

	if o.EnableCRC {
		opts = append(opts, alioss.EnableCRC(true))
	}

	if o.Timeout > 0 {
		opts = append(opts, alioss.Timeout(o.Timeout, o.Timeout))
	}

	return opts
}

// NewWithConfig 从配置创建阿里云 OSS 存储实例
//
// 配置键：
//   - oss.alioss.endpoint (string): OSS 端点地址（必需）
//   - oss.alioss.accessKeyID (string): Access Key ID（必需）
//   - oss.alioss.accessKeySecret (string): Access Key Secret（必需）
//   - oss.alioss.bucket (string): 默认存储桶名称（必需）
//   - oss.alioss.securityToken (string): STS 令牌（可选）
//   - oss.alioss.enableCRC (bool): 启用 CRC 校验（可选）
//   - oss.alioss.timeout (int): 超时时间（秒）（可选）
func NewWithConfig(cfg config.Config) (*Storage, error) {
	opts := []Option{}

	if token := cfg.GetString("oss.alioss.securityToken"); token != "" {
		opts = append(opts, WithSecurityToken(token))
	}

	if cfg.GetBool("oss.alioss.enableCRC") {
		opts = append(opts, WithEnableCRC(true))
	}

	if timeout := cfg.GetInt64("oss.alioss.timeout"); timeout > 0 {
		opts = append(opts, WithTimeout(timeout))
	}

	return New(
		cfg.GetString("oss.alioss.endpoint"),
		cfg.GetString("oss.alioss.accessKeyID"),
		cfg.GetString("oss.alioss.accessKeySecret"),
		cfg.GetString("oss.alioss.bucket"),
		opts...,
	)
}
