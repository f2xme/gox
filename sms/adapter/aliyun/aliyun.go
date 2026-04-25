package aliyun

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/sms"
)

type aliyunSMS struct {
	options Options
	client  aliyunSender
}

var _ sms.SMS = (*aliyunSMS)(nil)

type aliyunSender interface {
	SendSms(request *dysmsapi.SendSmsRequest) (*dysmsapi.SendSmsResponse, error)
}

// validateOptions 校验配置选项并设置默认值。
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

// createClient 根据已校验的配置创建阿里云短信客户端。
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
	o := defaultOptions()
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
func (s *aliyunSMS) Send(ctx context.Context, message sms.Message) error {
	if ctx == nil {
		return fmt.Errorf("aliyun sms: context cannot be nil")
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("aliyun sms: context error: %w", err)
	}
	if err := validateMessage(message); err != nil {
		return err
	}

	templateParam, err := encodeTemplateParam(message.TemplateParam)
	if err != nil {
		return err
	}

	request := &dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(message.Phone),
		SignName:      tea.String(s.options.SignName),
		TemplateCode:  tea.String(message.TemplateCode),
		TemplateParam: tea.String(templateParam),
	}

	resp, err := s.client.SendSms(request)
	if err != nil {
		return fmt.Errorf("aliyun sms: send sms: %w", err)
	}

	if resp == nil || resp.Body == nil {
		return fmt.Errorf("aliyun sms: send sms failed: empty response")
	}

	if resp.Body.Code == nil || *resp.Body.Code != "OK" {
		code := "unknown"
		if resp.Body.Code != nil {
			code = *resp.Body.Code
		}
		msg := "unknown error"
		if resp.Body.Message != nil {
			msg = *resp.Body.Message
		}
		requestID := ""
		if resp.Body.RequestId != nil {
			requestID = *resp.Body.RequestId
		}
		return fmt.Errorf("aliyun sms: send sms failed: code=%s message=%s requestID=%s", code, msg, requestID)
	}

	return nil
}

func validateMessage(message sms.Message) error {
	if strings.TrimSpace(message.Phone) == "" {
		return fmt.Errorf("aliyun sms: phone is required")
	}
	if strings.TrimSpace(message.TemplateCode) == "" {
		return fmt.Errorf("aliyun sms: template code is required")
	}
	return nil
}

func encodeTemplateParam(param any) (string, error) {
	if param == nil {
		return "", nil
	}

	switch v := param.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return "", nil
		}
		if !json.Valid([]byte(v)) {
			return "", fmt.Errorf("aliyun sms: template param must be valid json")
		}
		return v, nil
	case []byte:
		if len(v) == 0 {
			return "", nil
		}
		if !json.Valid(v) {
			return "", fmt.Errorf("aliyun sms: template param must be valid json")
		}
		return string(v), nil
	default:
		b, err := json.Marshal(param)
		if err != nil {
			return "", fmt.Errorf("aliyun sms: marshal template param: %w", err)
		}
		return string(b), nil
	}
}

// MustNew 创建由阿里云支持的 sms.SMS，如果失败则 log.Fatal
func MustNew(opts ...Option) sms.SMS {
	client, err := New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// NewWithConfig 使用 config.Config 创建由阿里云支持的 sms.SMS。
// 可选 prefix 参数用于自定义配置键前缀，默认值为 "sms"。
// 配置键：
//   - {prefix}.aliyun.accessKeyID：阿里云访问密钥 ID，必填
//   - {prefix}.aliyun.accessKeySecret：阿里云访问密钥 Secret，必填
//   - {prefix}.aliyun.endpoint：阿里云短信服务端点，选填，默认 dysmsapi.aliyuncs.com
//   - {prefix}.aliyun.signName：短信签名名称，必填
func NewWithConfig(cfg config.Config, prefix ...string) (sms.SMS, error) {
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

// MustNewWithConfig 使用 config.Config 创建由阿里云支持的 sms.SMS。
// 如果创建失败则调用 log.Fatal。
// 可选 prefix 参数用于自定义配置键前缀，默认值为 "sms"。
func MustNewWithConfig(cfg config.Config, prefix ...string) sms.SMS {
	client, err := NewWithConfig(cfg, prefix...)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// NewWithOptions 使用 Options 创建由阿里云支持的 sms.SMS。
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

// MustNewWithOptions 使用 Options 创建由阿里云支持的 sms.SMS。
// 如果创建失败则调用 log.Fatal。
func MustNewWithOptions(opts *Options) sms.SMS {
	client, err := NewWithOptions(opts)
	if err != nil {
		log.Fatal(err)
	}
	return client
}
