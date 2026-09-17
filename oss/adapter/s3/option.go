package s3

import "net/http"

// Options 定义 S3 配置；默认 Region 为 us-east-1，使用虚拟主机寻址。
type Options struct {
	// Endpoint 自定义 HTTP(S) 端点；空值使用 AWS S3 端点。
	Endpoint string
	// Region 签名地域；Cloudflare R2 使用 auto。
	Region string
	// AccessKeyID 访问密钥 ID。
	AccessKeyID string
	// AccessKeySecret 访问密钥。
	AccessKeySecret string
	// SecurityToken 临时凭证的会话令牌。
	SecurityToken string
	// Bucket 默认对象操作使用的存储桶。
	Bucket string
	// UsePathStyle 是否使用 endpoint/bucket/key 形式寻址。
	UsePathStyle bool
	// HTTPClient 自定义 HTTP 客户端；nil 使用 SDK 默认客户端。
	HTTPClient *http.Client
}

// Option 定义 S3 配置选项函数。
type Option func(*Options)

// WithEndpoint 设置自定义 HTTP(S) 端点。
func WithEndpoint(endpoint string) Option {
	return func(o *Options) { o.Endpoint = endpoint }
}

// WithRegion 设置签名地域；Cloudflare R2 使用 auto。
func WithRegion(region string) Option {
	return func(o *Options) { o.Region = region }
}

// WithCredentials 设置访问凭证。
func WithCredentials(accessKeyID, accessKeySecret string) Option {
	return func(o *Options) { o.AccessKeyID, o.AccessKeySecret = accessKeyID, accessKeySecret }
}

// WithSecurityToken 设置临时凭证的会话令牌。
func WithSecurityToken(token string) Option {
	return func(o *Options) { o.SecurityToken = token }
}

// WithBucket 设置默认存储桶。
func WithBucket(bucket string) Option {
	return func(o *Options) { o.Bucket = bucket }
}

// WithUsePathStyle 设置路径寻址；MinIO 等服务通常需要开启。
func WithUsePathStyle(enabled bool) Option {
	return func(o *Options) { o.UsePathStyle = enabled }
}

// WithHTTPClient 设置 HTTP 客户端，可用于配置超时或代理；调用后不应再修改客户端。
func WithHTTPClient(client *http.Client) Option {
	return func(o *Options) { o.HTTPClient = client }
}
