package aliyun

import (
	"fmt"

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
	o := Options{
		Endpoint: "dysmsapi.aliyuncs.com",
	}
	for _, opt := range opts {
		opt(&o)
	}

	if o.AccessKeyID == "" {
		return nil, fmt.Errorf("aliyun sms: access key id is required")
	}
	if o.AccessKeySecret == "" {
		return nil, fmt.Errorf("aliyun sms: access key secret is required")
	}
	if o.Endpoint == "" {
		return nil, fmt.Errorf("aliyun sms: endpoint is required")
	}
	if o.SignName == "" {
		return nil, fmt.Errorf("aliyun sms: sign name is required")
	}

	aliConfig := &openapi.Config{
		AccessKeyId:     tea.String(o.AccessKeyID),
		AccessKeySecret: tea.String(o.AccessKeySecret),
		Endpoint:        tea.String(o.Endpoint),
	}

	client, err := dysmsapi.NewClient(aliConfig)
	if err != nil {
		return nil, fmt.Errorf("aliyun sms: create client: %w", err)
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

// NewWithConfig creates a sms.SMS backed by Aliyun with configuration from config.Config.
// Configuration keys:
//   - sms.aliyun.accessKeyID (string): Aliyun access key ID (required)
//   - sms.aliyun.accessKeySecret (string): Aliyun access key secret (required)
//   - sms.aliyun.endpoint (string): Aliyun SMS endpoint (optional, default: dysmsapi.aliyuncs.com)
//   - sms.aliyun.signName (string): SMS signature name (required)
func NewWithConfig(cfg goxconfig.Config) (sms.SMS, error) {
	return New(
		WithAccessKeyID(cfg.GetString("sms.aliyun.accessKeyID")),
		WithAccessKeySecret(cfg.GetString("sms.aliyun.accessKeySecret")),
		WithEndpoint(cfg.GetString("sms.aliyun.endpoint")),
		WithSignName(cfg.GetString("sms.aliyun.signName")),
	)
}
