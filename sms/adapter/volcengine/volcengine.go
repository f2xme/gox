package volcengine

import (
	"context"
	"fmt"
	"log"

	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/sms"
)

type volcengineSMS struct {
	options Options
}

var _ sms.SMS = (*volcengineSMS)(nil)

// validateOptions 校验配置选项并设置默认值。
func validateOptions(o *Options) error {
	if o.Region == "" {
		o.Region = "cn-north-1"
	}

	if o.AccessKeyID == "" {
		return fmt.Errorf("volcengine sms: access key id is required")
	}
	if o.AccessKeySecret == "" {
		return fmt.Errorf("volcengine sms: access key secret is required")
	}
	if o.SignName == "" {
		return fmt.Errorf("volcengine sms: sign name is required")
	}

	return nil
}

// New 创建由火山引擎支持的 sms.SMS
//
// 使用选项模式配置客户端：
//
//	client, err := volcengine.New(
//		volcengine.WithAccessKeyID("your-key-id"),
//		volcengine.WithAccessKeySecret("your-key-secret"),
//		volcengine.WithSignName("your-sign-name"),
//	)
func New(opts ...Option) (sms.SMS, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	if err := validateOptions(&o); err != nil {
		return nil, err
	}

	return &volcengineSMS{
		options: o,
	}, nil
}

// Send 发送短信消息
func (s *volcengineSMS) Send(ctx context.Context, message sms.Message) error {
	// TODO: 接入火山引擎短信发送逻辑。
	return fmt.Errorf("volcengine sms provider not implemented yet")
}

// MustNew 创建由火山引擎支持的 sms.SMS，如果失败则 log.Fatal
func MustNew(opts ...Option) sms.SMS {
	client, err := New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// NewWithConfig 使用 config.Config 创建由火山引擎支持的 sms.SMS。
// 可选 prefix 参数用于自定义配置键前缀，默认值为 "sms"。
// 配置键：
//   - {prefix}.volcengine.accessKeyID：火山引擎访问密钥 ID，必填
//   - {prefix}.volcengine.accessKeySecret：火山引擎访问密钥 Secret，必填
//   - {prefix}.volcengine.region：火山引擎地域，选填，默认 cn-north-1
//   - {prefix}.volcengine.signName：短信签名名称，必填
func NewWithConfig(cfg config.Config, prefix ...string) (sms.SMS, error) {
	p := "sms"
	if len(prefix) > 0 && prefix[0] != "" {
		p = prefix[0]
	}
	return New(
		WithAccessKeyID(cfg.GetString(p+".volcengine.accessKeyID")),
		WithAccessKeySecret(cfg.GetString(p+".volcengine.accessKeySecret")),
		WithRegion(cfg.GetString(p+".volcengine.region")),
		WithSignName(cfg.GetString(p+".volcengine.signName")),
	)
}

// MustNewWithConfig 使用 config.Config 创建由火山引擎支持的 sms.SMS。
// 如果创建失败则调用 log.Fatal。
// 可选 prefix 参数用于自定义配置键前缀，默认值为 "sms"。
func MustNewWithConfig(cfg config.Config, prefix ...string) sms.SMS {
	client, err := NewWithConfig(cfg, prefix...)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// NewWithOptions 使用 Options 创建由火山引擎支持的 sms.SMS。
func NewWithOptions(opts *Options) (sms.SMS, error) {
	if opts == nil {
		return nil, fmt.Errorf("volcengine sms: options cannot be nil")
	}

	o := *opts
	if err := validateOptions(&o); err != nil {
		return nil, err
	}

	return &volcengineSMS{
		options: o,
	}, nil
}

// MustNewWithOptions 使用 Options 创建由火山引擎支持的 sms.SMS。
// 如果创建失败则调用 log.Fatal。
func MustNewWithOptions(opts *Options) sms.SMS {
	client, err := NewWithOptions(opts)
	if err != nil {
		log.Fatal(err)
	}
	return client
}
