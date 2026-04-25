package tencent

// Options 定义腾讯云短信配置选项
type Options struct {
	// SecretID 腾讯云密钥 ID
	SecretID string
	// SecretKey 腾讯云密钥 Key
	SecretKey string
	// Region 腾讯云地域
	Region string
	// AppID 短信应用 ID
	AppID string
	// SignName 短信签名名称
	SignName string
}

// Option 定义配置选项函数
type Option func(*Options)

// WithSecretID 设置密钥 ID
func WithSecretID(id string) Option {
	return func(o *Options) {
		o.SecretID = id
	}
}

// WithSecretKey 设置密钥 Key
func WithSecretKey(key string) Option {
	return func(o *Options) {
		o.SecretKey = key
	}
}

// WithRegion 设置地域
//
// 默认值: ap-guangzhou
func WithRegion(region string) Option {
	return func(o *Options) {
		o.Region = region
	}
}

// WithAppID 设置短信应用 ID
func WithAppID(appID string) Option {
	return func(o *Options) {
		o.AppID = appID
	}
}

// WithSignName 设置短信签名名称
func WithSignName(signName string) Option {
	return func(o *Options) {
		o.SignName = signName
	}
}
