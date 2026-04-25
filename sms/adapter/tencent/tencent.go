package tencent

import (
	"fmt"

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
	o := Options{
		Region: "ap-guangzhou",
	}
	for _, opt := range opts {
		opt(&o)
	}

	if o.SecretID == "" {
		return nil, fmt.Errorf("tencent sms: secret id is required")
	}
	if o.SecretKey == "" {
		return nil, fmt.Errorf("tencent sms: secret key is required")
	}
	if o.Region == "" {
		return nil, fmt.Errorf("tencent sms: region is required")
	}
	if o.AppID == "" {
		return nil, fmt.Errorf("tencent sms: app id is required")
	}
	if o.SignName == "" {
		return nil, fmt.Errorf("tencent sms: sign name is required")
	}

	credential := common.NewCredential(o.SecretID, o.SecretKey)
	cpf := profile.NewClientProfile()
	client, err := txsms.NewClient(credential, o.Region, cpf)
	if err != nil {
		return nil, fmt.Errorf("tencent sms: create client: %w", err)
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

// NewWithConfig creates a sms.SMS backed by Tencent with configuration from config.Config.
// Configuration keys:
//   - sms.tencent.secretID (string): Tencent secret ID (required)
//   - sms.tencent.secretKey (string): Tencent secret key (required)
//   - sms.tencent.region (string): Tencent region (optional, default: ap-guangzhou)
//   - sms.tencent.appID (string): SMS application ID (required)
//   - sms.tencent.signName (string): SMS signature name (required)
func NewWithConfig(cfg goxconfig.Config) (sms.SMS, error) {
	return New(
		WithSecretID(cfg.GetString("sms.tencent.secretID")),
		WithSecretKey(cfg.GetString("sms.tencent.secretKey")),
		WithRegion(cfg.GetString("sms.tencent.region")),
		WithAppID(cfg.GetString("sms.tencent.appID")),
		WithSignName(cfg.GetString("sms.tencent.signName")),
	)
}
