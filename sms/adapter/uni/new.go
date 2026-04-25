package uni

import (
	"fmt"
	"log"

	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/sms"
)

// New 创建由 UniSMS 支持的 sms.SMS。
//
// 使用选项模式配置客户端：
//
//	client, err := uni.New(
//		uni.WithAccessKeyID("your-key-id"),
//		uni.WithAccessKeySecret("your-key-secret"),
//		uni.WithSignName("your-sign-name"),
//	)
//
// AccessKeySecret 可为空，适用于 UniSMS 简易验签模式。
func New(opts ...Option) (sms.SMS, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	if err := validateOptions(&o); err != nil {
		return nil, err
	}

	return &uniSMS{
		options: o,
		client:  createClient(&o),
	}, nil
}

// MustNew 创建由 UniSMS 支持的 sms.SMS，如果失败则 log.Fatal。
func MustNew(opts ...Option) sms.SMS {
	client, err := New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// NewWithConfig 使用 config.Config 创建由 UniSMS 支持的 sms.SMS。
// 可选 prefix 参数用于自定义配置键前缀，默认值为 "sms"。
// 配置键：
//   - {prefix}.uni.accessKeyID：UniSMS 访问密钥 ID，必填
//   - {prefix}.uni.accessKeySecret：UniSMS 访问密钥 Secret，简易验签模式可空
//   - {prefix}.uni.signName：短信签名名称，必填
func NewWithConfig(cfg config.Config, prefix ...string) (sms.SMS, error) {
	p := "sms"
	if len(prefix) > 0 && prefix[0] != "" {
		p = prefix[0]
	}
	return New(
		WithAccessKeyID(cfg.GetString(p+".uni.accessKeyID")),
		WithAccessKeySecret(cfg.GetString(p+".uni.accessKeySecret")),
		WithSignName(cfg.GetString(p+".uni.signName")),
	)
}

// MustNewWithConfig 使用 config.Config 创建由 UniSMS 支持的 sms.SMS。
// 如果创建失败则调用 log.Fatal。
// 可选 prefix 参数用于自定义配置键前缀，默认值为 "sms"。
func MustNewWithConfig(cfg config.Config, prefix ...string) sms.SMS {
	client, err := NewWithConfig(cfg, prefix...)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// NewWithOptions 使用 Options 创建由 UniSMS 支持的 sms.SMS。
func NewWithOptions(opts *Options) (sms.SMS, error) {
	if opts == nil {
		return nil, fmt.Errorf("uni sms: options cannot be nil")
	}

	o := *opts
	if err := validateOptions(&o); err != nil {
		return nil, err
	}

	return &uniSMS{
		options: o,
		client:  createClient(&o),
	}, nil
}

// MustNewWithOptions 使用 Options 创建由 UniSMS 支持的 sms.SMS。
// 如果创建失败则调用 log.Fatal。
func MustNewWithOptions(opts *Options) sms.SMS {
	client, err := NewWithOptions(opts)
	if err != nil {
		log.Fatal(err)
	}
	return client
}
