package volcengine

// Options 定义火山引擎短信配置选项
type Options struct {
	// AccessKeyID 火山引擎访问密钥 ID
	AccessKeyID string
	// AccessKeySecret 火山引擎访问密钥 Secret
	AccessKeySecret string
	// Region 火山引擎地域
	Region string
	// SignName 短信签名名称
	SignName string
}

// Option 定义配置选项函数
type Option func(*Options)

// defaultOptions 返回默认配置选项
func defaultOptions() Options {
	return Options{
		Region: "cn-north-1",
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

// WithRegion 设置地域
//
// 默认值: cn-north-1
//
// 示例：
//
//	New(WithRegion("cn-north-1"))
func WithRegion(region string) Option {
	return func(o *Options) {
		o.Region = region
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
