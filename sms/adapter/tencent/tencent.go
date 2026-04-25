package tencent

import (
	"fmt"
	"log"

	goxconfig "github.com/f2xme/gox/config"
	"github.com/f2xme/gox/sms"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	txsms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type tencentSMS struct {
	options Options
	client  *txsms.Client
}

var _ sms.SMS = (*tencentSMS)(nil)

// validateOptions validates the options and sets defaults
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

// createClient creates a new Tencent SMS client from validated options
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

	return &tencentSMS{
		options: o,
		client:  client,
	}, nil
}

// Send 发送短信消息
func (s *tencentSMS) Send(phone, templateCode, templateParam string) error {
	request := txsms.NewSendSmsRequest()
	request.SmsSdkAppId = common.StringPtr(s.options.AppID)
	request.SignName = common.StringPtr(s.options.SignName)
	request.TemplateId = common.StringPtr(templateCode)
	request.PhoneNumberSet = common.StringPtrs([]string{phone})
	request.TemplateParamSet = common.StringPtrs([]string{templateParam})

	response, err := s.client.SendSms(request)
	if err != nil {
		return fmt.Errorf("send sms: %w", err)
	}

	if len(response.Response.SendStatusSet) == 0 {
		return fmt.Errorf("send sms failed: no response")
	}

	status := response.Response.SendStatusSet[0]
	if status.Code == nil || *status.Code != "Ok" {
		msg := "unknown error"
		if status.Message != nil {
			msg = *status.Message
		}
		return fmt.Errorf("send sms failed: %s", msg)
	}

	return nil
}

// MustNew 创建由腾讯云支持的 sms.SMS，如果失败则 log.Fatal
func MustNew(opts ...Option) sms.SMS {
	client, err := New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// NewWithConfig creates a sms.SMS backed by Tencent with configuration from config.Config.
// The optional prefix parameter allows customizing the configuration key prefix (default: "sms").
// Configuration keys:
//   - {prefix}.tencent.secretID (string): Tencent secret ID (required)
//   - {prefix}.tencent.secretKey (string): Tencent secret key (required)
//   - {prefix}.tencent.region (string): Tencent region (optional, default: ap-guangzhou)
//   - {prefix}.tencent.appID (string): SMS application ID (required)
//   - {prefix}.tencent.signName (string): SMS signature name (required)
func NewWithConfig(cfg goxconfig.Config, prefix ...string) (sms.SMS, error) {
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

// MustNewWithConfig creates a sms.SMS backed by Tencent with configuration from config.Config.
// Calls log.Fatal if creation fails.
// The optional prefix parameter allows customizing the configuration key prefix (default: "sms").
func MustNewWithConfig(cfg goxconfig.Config, prefix ...string) sms.SMS {
	client, err := NewWithConfig(cfg, prefix...)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// NewWithOptions creates a sms.SMS backed by Tencent using an Options struct.
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

// MustNewWithOptions creates a sms.SMS backed by Tencent using an Options struct.
// Calls log.Fatal if creation fails.
func MustNewWithOptions(opts *Options) sms.SMS {
	client, err := NewWithOptions(opts)
	if err != nil {
		log.Fatal(err)
	}
	return client
}
