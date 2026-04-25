package tencent

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	txsms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"

	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/sms"
)

type tencentSMS struct {
	options Options
	client  tencentSender
}

var _ sms.SMS = (*tencentSMS)(nil)

type tencentSender interface {
	SendSmsWithContext(ctx context.Context, request *txsms.SendSmsRequest) (*txsms.SendSmsResponse, error)
}

// validateOptions 校验配置选项并设置默认值。
func validateOptions(o *Options) error {
	if o.Region == "" {
		o.Region = "ap-guangzhou"
	}

	if o.SecretID == "" {
		return fmt.Errorf("tencent sms: secret id is required")
	}
	if o.SecretKey == "" {
		return fmt.Errorf("tencent sms: secret key is required")
	}
	if o.AppID == "" {
		return fmt.Errorf("tencent sms: app id is required")
	}
	if o.SignName == "" {
		return fmt.Errorf("tencent sms: sign name is required")
	}

	return nil
}

// createClient 根据已校验的配置创建腾讯云短信客户端。
func createClient(o *Options) (*txsms.Client, error) {
	credential := common.NewCredential(o.SecretID, o.SecretKey)
	cpf := profile.NewClientProfile()
	client, err := txsms.NewClient(credential, o.Region, cpf)
	if err != nil {
		return nil, fmt.Errorf("tencent sms: create client: %w", err)
	}

	return client, nil
}

// New 创建由腾讯云支持的 sms.SMS
//
// 使用选项模式配置客户端：
//
//	client, err := tencent.New(
//		tencent.WithSecretID("your-secret-id"),
//		tencent.WithSecretKey("your-secret-key"),
//		tencent.WithAppID("your-app-id"),
//		tencent.WithSignName("your-sign-name"),
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

	return &tencentSMS{
		options: o,
		client:  client,
	}, nil
}

// Send 发送短信消息
func (s *tencentSMS) Send(ctx context.Context, message sms.Message) error {
	if ctx == nil {
		return fmt.Errorf("tencent sms: context cannot be nil")
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("tencent sms: context error: %w", err)
	}
	if err := validateMessage(message); err != nil {
		return err
	}

	templateParams, err := normalizeTemplateParams(message.TemplateParam)
	if err != nil {
		return err
	}

	request := txsms.NewSendSmsRequest()
	request.SmsSdkAppId = common.StringPtr(s.options.AppID)
	request.SignName = common.StringPtr(s.options.SignName)
	request.TemplateId = common.StringPtr(message.TemplateCode)
	request.PhoneNumberSet = common.StringPtrs([]string{message.Phone})
	request.TemplateParamSet = common.StringPtrs(templateParams)

	response, err := s.client.SendSmsWithContext(ctx, request)
	if err != nil {
		return fmt.Errorf("tencent sms: send sms: %w", err)
	}

	if response == nil || response.Response == nil {
		return fmt.Errorf("tencent sms: send sms failed: empty response")
	}
	if len(response.Response.SendStatusSet) == 0 {
		return fmt.Errorf("tencent sms: send sms failed: no response")
	}

	status := response.Response.SendStatusSet[0]
	if status.Code == nil || *status.Code != "Ok" {
		code := "unknown"
		if status.Code != nil {
			code = *status.Code
		}
		msg := "unknown error"
		if status.Message != nil {
			msg = *status.Message
		}
		requestID := ""
		if response.Response.RequestId != nil {
			requestID = *response.Response.RequestId
		}
		return fmt.Errorf("tencent sms: send sms failed: code=%s message=%s requestID=%s", code, msg, requestID)
	}

	return nil
}

func validateMessage(message sms.Message) error {
	if strings.TrimSpace(message.Phone) == "" {
		return fmt.Errorf("tencent sms: phone is required")
	}
	if strings.TrimSpace(message.TemplateCode) == "" {
		return fmt.Errorf("tencent sms: template code is required")
	}
	return nil
}

func normalizeTemplateParams(param any) ([]string, error) {
	if param == nil {
		return nil, nil
	}

	switch v := param.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return nil, nil
		}
		return []string{v}, nil
	case []string:
		return v, nil
	case []any:
		params := make([]string, len(v))
		for i, item := range v {
			params[i] = fmt.Sprint(item)
		}
		return params, nil
	default:
		return nil, fmt.Errorf("tencent sms: template param must be string, []string, or []any")
	}
}

// MustNew 创建由腾讯云支持的 sms.SMS，如果失败则 log.Fatal
func MustNew(opts ...Option) sms.SMS {
	client, err := New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// NewWithConfig 使用 config.Config 创建由腾讯云支持的 sms.SMS。
// 可选 prefix 参数用于自定义配置键前缀，默认值为 "sms"。
// 配置键：
//   - {prefix}.tencent.secretID：腾讯云密钥 ID，必填
//   - {prefix}.tencent.secretKey：腾讯云密钥 Key，必填
//   - {prefix}.tencent.region：腾讯云地域，选填，默认 ap-guangzhou
//   - {prefix}.tencent.appID：短信应用 ID，必填
//   - {prefix}.tencent.signName：短信签名名称，必填
func NewWithConfig(cfg config.Config, prefix ...string) (sms.SMS, error) {
	p := "sms"
	if len(prefix) > 0 && prefix[0] != "" {
		p = prefix[0]
	}
	return New(
		WithSecretID(cfg.GetString(p+".tencent.secretID")),
		WithSecretKey(cfg.GetString(p+".tencent.secretKey")),
		WithRegion(cfg.GetString(p+".tencent.region")),
		WithAppID(cfg.GetString(p+".tencent.appID")),
		WithSignName(cfg.GetString(p+".tencent.signName")),
	)
}

// MustNewWithConfig 使用 config.Config 创建由腾讯云支持的 sms.SMS。
// 如果创建失败则调用 log.Fatal。
// 可选 prefix 参数用于自定义配置键前缀，默认值为 "sms"。
func MustNewWithConfig(cfg config.Config, prefix ...string) sms.SMS {
	client, err := NewWithConfig(cfg, prefix...)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// NewWithOptions 使用 Options 创建由腾讯云支持的 sms.SMS。
func NewWithOptions(opts *Options) (sms.SMS, error) {
	if opts == nil {
		return nil, fmt.Errorf("tencent sms: options cannot be nil")
	}

	o := *opts
	if err := validateOptions(&o); err != nil {
		return nil, err
	}

	client, err := createClient(&o)
	if err != nil {
		return nil, err
	}

	return &tencentSMS{
		options: o,
		client:  client,
	}, nil
}

// MustNewWithOptions 使用 Options 创建由腾讯云支持的 sms.SMS。
// 如果创建失败则调用 log.Fatal。
func MustNewWithOptions(opts *Options) sms.SMS {
	client, err := NewWithOptions(opts)
	if err != nil {
		log.Fatal(err)
	}
	return client
}
