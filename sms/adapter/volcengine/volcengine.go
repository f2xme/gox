package volcengine

import (
	"fmt"
	"log"

	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/sms"
)

type volcengineSMS struct {
	options Options
}

var _ sms.SMS = (*volcengineSMS)(nil)

// validateOptions validates the options and sets defaults
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
	o := Options{}
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
func (s *volcengineSMS) Send(phone, templateCode, templateParam string) error {
	// TODO: Implement Volcengine SMS sending logic
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

// NewWithConfig creates a sms.SMS backed by Volcengine with configuration from config.Config.
// The optional prefix parameter allows customizing the configuration key prefix (default: "sms").
// Configuration keys:
//   - {prefix}.volcengine.accessKeyID (string): Volcengine access key ID (required)
//   - {prefix}.volcengine.accessKeySecret (string): Volcengine access key secret (required)
//   - {prefix}.volcengine.region (string): Volcengine region (optional, default: cn-north-1)
//   - {prefix}.volcengine.signName (string): SMS signature name (required)
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

// MustNewWithConfig creates a sms.SMS backed by Volcengine with configuration from config.Config.
// Calls log.Fatal if creation fails.
// The optional prefix parameter allows customizing the configuration key prefix (default: "sms").
func MustNewWithConfig(cfg config.Config, prefix ...string) sms.SMS {
	client, err := NewWithConfig(cfg, prefix...)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// NewWithOptions creates a sms.SMS backed by Volcengine using an Options struct.
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

// MustNewWithOptions creates a sms.SMS backed by Volcengine using an Options struct.
// Calls log.Fatal if creation fails.
func MustNewWithOptions(opts *Options) sms.SMS {
	client, err := NewWithOptions(opts)
	if err != nil {
		log.Fatal(err)
	}
	return client
}
