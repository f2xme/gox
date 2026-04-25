package aliyun

import (
	"fmt"
	"log"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	goxconfig "github.com/f2xme/gox/config"
	"github.com/f2xme/gox/sms"
)

type aliyunSMS struct {
	options Options
	client  *dysmsapi.Client
}

var _ sms.SMS = (*aliyunSMS)(nil)

// validateOptions validates the options and sets defaults
func validateOptions(o *Options) error {
	if o.Endpoint == "" {
		o.Endpoint = "dysmsapi.aliyuncs.com"
	}

	if o.AccessKeyID == "" {
		return fmt.Errorf("aliyun sms: access key id is required")
	}
	if o.AccessKeySecret == "" {
		return fmt.Errorf("aliyun sms: access key secret is required")
	}
	if o.SignName == "" {
		return fmt.Errorf("aliyun sms: sign name is required")
	}

	return nil
}

// createClient creates a new Aliyun SMS client from validated options
func createClient(o *Options) (*dysmsapi.Client, error) {
	aliConfig := &openapi.Config{
		AccessKeyId:     tea.String(o.AccessKeyID),
		AccessKeySecret: tea.String(o.AccessKeySecret),
		Endpoint:        tea.String(o.Endpoint),
	}

	client, err := dysmsapi.NewClient(aliConfig)
	if err != nil {
		return nil, fmt.Errorf("aliyun sms: create client: %w", err)
	}

	return client, nil
}

// New 创建由阿里云支持的 sms.SMS
//
// 使用选项模式配置客户端：
//
//	client, err := aliyun.New(
//		aliyun.WithAccessKeyID("your-key-id"),
//		aliyun.WithAccessKeySecret("your-key-secret"),
//		aliyun.WithSignName("your-sign-name"),
//	)
func New(opts ...Option) (sms.SMS, error) {
	o := Options{}
	for _, opt := range opts {
		opt(&o)
	}

	if err := validateOptions(&o); err != nil {
		return nil, err
	}

	client, err := createClient(&o)
	if err != nil {
		return nil, err
	}

	return &aliyunSMS{
		options: o,
		client:  client,
	}, nil
}

// Send 发送短信消息
func (s *aliyunSMS) Send(phone, templateCode, templateParam string) error {
	request := &dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(phone),
		SignName:      tea.String(s.options.SignName),
		TemplateCode:  tea.String(templateCode),
		TemplateParam: tea.String(templateParam),
	}

	resp, err := s.client.SendSms(request)
	if err != nil {
		return fmt.Errorf("send sms: %w", err)
	}

	if resp.Body == nil || resp.Body.Code == nil || *resp.Body.Code != "OK" {
		msg := "unknown error"
		if resp.Body != nil && resp.Body.Message != nil {
			msg = *resp.Body.Message
		}
		return fmt.Errorf("send sms failed: %s", msg)
	}

	return nil
}

// MustNew 创建由阿里云支持的 sms.SMS，如果失败则 log.Fatal
func MustNew(opts ...Option) sms.SMS {
	client, err := New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// NewWithConfig creates a sms.SMS backed by Aliyun with configuration from config.Config.
// The optional prefix parameter allows customizing the configuration key prefix (default: "sms").
// Configuration keys:
//   - {prefix}.aliyun.accessKeyID (string): Aliyun access key ID (required)
//   - {prefix}.aliyun.accessKeySecret (string): Aliyun access key secret (required)
//   - {prefix}.aliyun.endpoint (string): Aliyun SMS endpoint (optional, default: dysmsapi.aliyuncs.com)
//   - {prefix}.aliyun.signName (string): SMS signature name (required)
func NewWithConfig(cfg goxconfig.Config, prefix ...string) (sms.SMS, error) {
	p := "sms"
	if len(prefix) > 0 && prefix[0] != "" {
		p = prefix[0]
	}
	return New(
		WithAccessKeyID(cfg.GetString(p+".aliyun.accessKeyID")),
		WithAccessKeySecret(cfg.GetString(p+".aliyun.accessKeySecret")),
		WithEndpoint(cfg.GetString(p+".aliyun.endpoint")),
		WithSignName(cfg.GetString(p+".aliyun.signName")),
	)
}

// MustNewWithConfig creates a sms.SMS backed by Aliyun with configuration from config.Config.
// Calls log.Fatal if creation fails.
// The optional prefix parameter allows customizing the configuration key prefix (default: "sms").
func MustNewWithConfig(cfg goxconfig.Config, prefix ...string) sms.SMS {
	client, err := NewWithConfig(cfg, prefix...)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// NewWithOptions creates a sms.SMS backed by Aliyun using an Options struct.
func NewWithOptions(opts *Options) (sms.SMS, error) {
	if opts == nil {
		return nil, fmt.Errorf("aliyun sms: options cannot be nil")
	}

	o := *opts
	if err := validateOptions(&o); err != nil {
		return nil, err
	}

	client, err := createClient(&o)
	if err != nil {
		return nil, err
	}

	return &aliyunSMS{
		options: o,
		client:  client,
	}, nil
}

// MustNewWithOptions creates a sms.SMS backed by Aliyun using an Options struct.
// Calls log.Fatal if creation fails.
func MustNewWithOptions(opts *Options) sms.SMS {
	client, err := NewWithOptions(opts)
	if err != nil {
		log.Fatal(err)
	}
	return client
}
