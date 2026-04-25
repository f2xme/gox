package aliyun

// Options 定义阿里云短信配置选项
type Options struct {
	// AccessKeyID 阿里云访问密钥 ID
	AccessKeyID string
	// AccessKeySecret 阿里云访问密钥 Secret
	AccessKeySecret string
	// Endpoint 阿里云短信服务端点
	Endpoint string
	// SignName 短信签名名称
	SignName string
}

// Option 定义配置选项函数
type Option func(*Options)

// defaultOptions 返回默认配置选项
func defaultOptions() Options {
	return Options{
		Endpoint: "dysmsapi.aliyuncs.com",
	}
}

// WithAccessKeyID 设置访问密钥 ID
//
// 示例：
//
//	New(WithAccessKeyID("your-key-id"))
func WithAccessKeyID(id string) Option {
	return func(o *Options) {
		o.AccessKeyID = id
	}
}

// WithAccessKeySecret 设置访问密钥 Secret
//
// 示例：
//
//	New(WithAccessKeySecret("your-key-secret"))
func WithAccessKeySecret(secret string) Option {
	return func(o *Options) {
		o.AccessKeySecret = secret
	}
}

// WithEndpoint 设置短信服务端点
//
// 默认值: dysmsapi.aliyuncs.com
//
// 示例：
//
//	New(WithEndpoint("dysmsapi.aliyuncs.com"))
func WithEndpoint(endpoint string) Option {
	return func(o *Options) {
		o.Endpoint = endpoint
	}
}

// WithSignName 设置短信签名名称
//
// 示例：
//
//	New(WithSignName("your-sign-name"))
func WithSignName(signName string) Option {
	return func(o *Options) {
		o.SignName = signName
	}
}
