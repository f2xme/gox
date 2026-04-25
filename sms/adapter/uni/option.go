package uni

import (
	"fmt"
	"strings"
)

// Options 定义 UniSMS 短信配置选项。
type Options struct {
	// AccessKeyID UniSMS 访问密钥 ID。
	AccessKeyID string
	// AccessKeySecret UniSMS 访问密钥 Secret，简易验签模式可为空。
	AccessKeySecret string
	// SignName 短信签名名称。
	SignName string
}

// Option 定义配置选项函数。
type Option func(*Options)

// defaultOptions 返回默认配置选项。
func defaultOptions() Options {
	return Options{}
}

// validateOptions 校验配置选项并设置默认值。
func validateOptions(o *Options) error {
	if strings.TrimSpace(o.AccessKeyID) == "" {
		return fmt.Errorf("uni sms: access key id is required")
	}
	if strings.TrimSpace(o.SignName) == "" {
		return fmt.Errorf("uni sms: sign name is required")
	}

	return nil
}

// WithAccessKeyID 设置访问密钥 ID。
//
// 示例：
//
//	New(WithAccessKeyID("your-key-id"))
func WithAccessKeyID(id string) Option {
	return func(o *Options) {
		o.AccessKeyID = id
	}
}

// WithAccessKeySecret 设置访问密钥 Secret。
//
// 简易验签模式可不设置该选项。
//
// 示例：
//
//	New(WithAccessKeySecret("your-key-secret"))
func WithAccessKeySecret(secret string) Option {
	return func(o *Options) {
		o.AccessKeySecret = secret
	}
}

// WithSignName 设置短信签名名称。
//
// 示例：
//
//	New(WithSignName("your-sign-name"))
func WithSignName(signName string) Option {
	return func(o *Options) {
		o.SignName = signName
	}
}
