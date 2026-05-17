package volcengine

import (
	"context"
	"errors"
	"fmt"

	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/sms"
)

type volcengineSMS struct {
	options Options
}

// ErrNotImplemented indicates that this adapter does not yet connect to
// Volcengine SMS.
var ErrNotImplemented = errors.New("volcengine sms provider is not implemented")

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

// New validates options and returns ErrNotImplemented.
//
// 当前版本不会创建真实的火山引擎短信客户端。保留此构造函数是为了避免
// 静默返回一个无法发送短信的占位客户端。
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

	return nil, ErrNotImplemented
}

// Send returns ErrNotImplemented.
func (s *volcengineSMS) Send(ctx context.Context, message sms.Message) error {
	return ErrNotImplemented
}

// MustNew calls New and panics on any error, including ErrNotImplemented.
func MustNew(opts ...Option) sms.SMS {
	client, err := New(opts...)
	if err != nil {
		panic(fmt.Errorf("volcengine sms: create client failed: %w", err))
	}
	return client
}

// NewWithConfig reads options from config.Config and returns ErrNotImplemented
// after validation succeeds.
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

// MustNewWithConfig calls NewWithConfig and panics on any error, including
// ErrNotImplemented.
// 可选 prefix 参数用于自定义配置键前缀，默认值为 "sms"。
func MustNewWithConfig(cfg config.Config, prefix ...string) sms.SMS {
	client, err := NewWithConfig(cfg, prefix...)
	if err != nil {
		panic(fmt.Errorf("volcengine sms: create client from config failed: %w", err))
	}
	return client
}

// NewWithOptions validates Options and returns ErrNotImplemented.
func NewWithOptions(opts *Options) (sms.SMS, error) {
	if opts == nil {
		return nil, fmt.Errorf("volcengine sms: options cannot be nil")
	}

	o := *opts
	if err := validateOptions(&o); err != nil {
		return nil, err
	}

	return nil, ErrNotImplemented
}

// MustNewWithOptions calls NewWithOptions and panics on any error, including
// ErrNotImplemented.
func MustNewWithOptions(opts *Options) sms.SMS {
	client, err := NewWithOptions(opts)
	if err != nil {
		panic(fmt.Errorf("volcengine sms: create client from options failed: %w", err))
	}
	return client
}
