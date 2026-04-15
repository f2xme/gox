package oss

import "time"

const (
	MethodGet    = "GET"    // HTTP GET 方法
	MethodPut    = "PUT"    // HTTP PUT 方法
	MethodDelete = "DELETE" // HTTP DELETE 方法
)

// PutOption 定义上传选项函数
type PutOption func(*PutOptions)

// PutOptions 定义上传配置选项
type PutOptions struct {
	ContentType string            // 内容类型
	Metadata    map[string]string // 自定义元数据
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
	Start int64 // 起始字节位置（-1 表示不使用范围下载）
	End   int64 // 结束字节位置
}

// WithRange 设置范围下载
//
// 示例：
//
//	// 下载前 1KB
//	storage.Get(ctx, key, oss.WithRange(0, 1023))
func WithRange(start, end int64) GetOption {
	return func(o *GetOptions) {
		o.Start = start
		o.End = end
	}
}

// ListOption 定义列表选项函数
type ListOption func(*ListOptions)

// ListOptions 定义列表配置选项
type ListOptions struct {
	Prefix    string // 对象键前缀
	Delimiter string // 分隔符
	MaxKeys   int    // 最大返回数量
	Marker    string // 分页标记
}

// DefaultListOptions 返回默认列表选项
func DefaultListOptions() *ListOptions {
	return &ListOptions{MaxKeys: 1000}
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
		o.MaxKeys = maxKeys
	}
}

// WithMarker 设置分页标记
//
// 示例：
//
//	storage.List(ctx, oss.WithMarker("last-key"))
func WithMarker(marker string) ListOption {
	return func(o *ListOptions) {
		o.Marker = marker
	}
}

// PresignedOption 定义预签名选项函数
type PresignedOption func(*PresignedOptions)

// PresignedOptions 定义预签名配置选项
type PresignedOptions struct {
	Method  string        // HTTP 方法
	Expires time.Duration // 过期时间
}

// DefaultPresignedOptions 返回默认预签名选项
func DefaultPresignedOptions() *PresignedOptions {
	return &PresignedOptions{
		Method:  MethodGet,
		Expires: 15 * time.Minute,
	}
}

// WithMethod 设置 HTTP 方法
//
// 示例：
//
//	storage.PresignedURL(ctx, key, oss.WithMethod(oss.MethodPut))
func WithMethod(method string) PresignedOption {
	return func(o *PresignedOptions) {
		o.Method = method
	}
}

// WithExpires 设置过期时间
//
// 示例：
//
//	storage.PresignedURL(ctx, key, oss.WithExpires(30*time.Minute))
func WithExpires(expires time.Duration) PresignedOption {
	return func(o *PresignedOptions) {
		o.Expires = expires
	}
}
