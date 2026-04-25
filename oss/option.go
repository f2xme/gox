package oss

import "time"

const (
	// MethodGet 表示 HTTP GET 方法
	MethodGet = "GET"
	// MethodPut 表示 HTTP PUT 方法
	MethodPut = "PUT"
	// MethodDelete 表示 HTTP DELETE 方法
	MethodDelete = "DELETE"
)

// PutOption 定义上传选项函数
type PutOption func(*PutOptions)

// PutOptions 定义上传配置选项
type PutOptions struct {
	// ContentType 内容类型，空值时根据对象键自动推断
	ContentType string
	// Metadata 用户自定义元数据
	Metadata map[string]string
}

func defaultPutOptions() PutOptions {
	return PutOptions{}
}

// WithContentType 设置内容类型
//
// 示例：
//
//	storage.Put(ctx, key, reader, oss.WithContentType("image/jpeg"))
func WithContentType(contentType string) PutOption {
	return func(o *PutOptions) {
		o.ContentType = contentType
	}
}

// WithMetadata 设置自定义元数据
//
// 示例：
//
//	storage.Put(ctx, key, reader, oss.WithMetadata(map[string]string{
//		"author": "alice",
//		"version": "1.0",
//	}))
func WithMetadata(metadata map[string]string) PutOption {
	return func(o *PutOptions) {
		o.Metadata = metadata
	}
}

// GetOption 定义下载选项函数
type GetOption func(*GetOptions)

// GetOptions 定义下载配置选项
type GetOptions struct {
	// RangeStart 起始字节位置，-1 表示不使用范围下载
	RangeStart int64
	// RangeEnd 结束字节位置
	RangeEnd int64
}

func defaultGetOptions() GetOptions {
	return GetOptions{RangeStart: -1}
}

// WithRange 设置范围下载
//
// 示例：
//
//	// 下载前 1KB
//	storage.Get(ctx, key, oss.WithRange(0, 1023))
func WithRange(start, end int64) GetOption {
	return func(o *GetOptions) {
		o.RangeStart = start
		o.RangeEnd = end
	}
}

// ListOption 定义列表选项函数
type ListOption func(*ListOptions)

// ListOptions 定义列表配置选项
type ListOptions struct {
	// Prefix 对象键前缀
	Prefix string
	// Delimiter 分隔符，常用 "/" 模拟目录
	Delimiter string
	// Limit 最大返回数量
	Limit int
	// Token 分页令牌
	Token string
}

func defaultListOptions() ListOptions {
	return ListOptions{Limit: 1000}
}

// WithPrefix 设置对象键前缀
//
// 示例：
//
//	storage.List(ctx, oss.WithPrefix("images/"))
func WithPrefix(prefix string) ListOption {
	return func(o *ListOptions) {
		o.Prefix = prefix
	}
}

// WithDelimiter 设置分隔符
//
// 示例：
//
//	storage.List(ctx, oss.WithDelimiter("/"))
func WithDelimiter(delimiter string) ListOption {
	return func(o *ListOptions) {
		o.Delimiter = delimiter
	}
}

// WithMaxKeys 设置最大返回数量
//
// 示例：
//
//	storage.List(ctx, oss.WithMaxKeys(100))
func WithMaxKeys(maxKeys int) ListOption {
	return func(o *ListOptions) {
		o.Limit = maxKeys
	}
}

// WithLimit 设置最大返回数量
//
// 示例：
//
//	storage.List(ctx, oss.WithLimit(100))
func WithLimit(limit int) ListOption {
	return func(o *ListOptions) {
		o.Limit = limit
	}
}

// WithToken 设置分页令牌
//
// 示例：
//
//	storage.List(ctx, oss.WithToken(result.NextToken))
func WithToken(token string) ListOption {
	return func(o *ListOptions) {
		o.Token = token
	}
}

// WithMarker 设置兼容旧调用的分页标记
//
// Deprecated: 使用 WithToken。
func WithMarker(marker string) ListOption {
	return WithToken(marker)
}

// SignOption 定义预签名选项函数
type SignOption func(*SignOptions)

// SignOptions 定义预签名配置选项
type SignOptions struct {
	// Method HTTP 方法
	Method string
	// Expires 过期时间
	Expires time.Duration
	// ContentType PUT 预签名时使用的内容类型
	ContentType string
}

func defaultSignOptions() SignOptions {
	return SignOptions{
		Method:  MethodGet,
		Expires: 15 * time.Minute,
	}
}

// WithMethod 设置 HTTP 方法
//
// 示例：
//
//	storage.SignURL(ctx, key, oss.WithMethod(oss.MethodPut))
func WithMethod(method string) SignOption {
	return func(o *SignOptions) {
		o.Method = method
	}
}

// WithExpires 设置过期时间
//
// 示例：
//
//	storage.SignURL(ctx, key, oss.WithExpires(30*time.Minute))
func WithExpires(expires time.Duration) SignOption {
	return func(o *SignOptions) {
		o.Expires = expires
	}
}

// WithSignContentType 设置预签名请求的内容类型
//
// 示例：
//
//	storage.SignURL(ctx, key, oss.WithMethod(oss.MethodPut), oss.WithSignContentType("image/png"))
func WithSignContentType(contentType string) SignOption {
	return func(o *SignOptions) {
		o.ContentType = contentType
	}
}

// PresignedOption 定义兼容旧名称的预签名选项函数
//
// Deprecated: 使用 SignOption。
type PresignedOption = SignOption

// PresignedOptions 定义兼容旧名称的预签名配置选项
//
// Deprecated: 使用 SignOptions。
type PresignedOptions = SignOptions

// BucketOption 定义存储桶配置选项
type BucketOption func(*BucketOptions)

// BucketOptions 定义存储桶配置选项
type BucketOptions struct {
	// Region 存储桶地域
	Region string
	// ACL 存储桶访问控制策略
	ACL string
}

func defaultBucketOptions() BucketOptions {
	return BucketOptions{}
}

// WithBucketRegion 设置存储桶地域
//
// 示例：
//
//	storage.CreateBucket(ctx, "my-bucket", oss.WithBucketRegion("oss-cn-hangzhou"))
func WithBucketRegion(region string) BucketOption {
	return func(o *BucketOptions) {
		o.Region = region
	}
}

// WithBucketACL 设置存储桶访问控制策略
//
// 示例：
//
//	storage.CreateBucket(ctx, "my-bucket", oss.WithBucketACL("private"))
func WithBucketACL(acl string) BucketOption {
	return func(o *BucketOptions) {
		o.ACL = acl
	}
}

// ApplyPutOptions 合并上传选项
func ApplyPutOptions(opts ...PutOption) PutOptions {
	options := defaultPutOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// ApplyGetOptions 合并下载选项
func ApplyGetOptions(opts ...GetOption) GetOptions {
	options := defaultGetOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// ApplyListOptions 合并列表选项
func ApplyListOptions(opts ...ListOption) ListOptions {
	options := defaultListOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// ApplySignOptions 合并预签名选项
func ApplySignOptions(opts ...SignOption) SignOptions {
	options := defaultSignOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// ApplyBucketOptions 合并存储桶选项
func ApplyBucketOptions(opts ...BucketOption) BucketOptions {
	options := defaultBucketOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return options
}
