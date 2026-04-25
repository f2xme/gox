package uni

import (
	"context"
	"fmt"
	"strings"

	uniapi "github.com/apistd/uni-go-sdk"
	unisms "github.com/apistd/uni-go-sdk/sms"

	"github.com/f2xme/gox/sms"
)

type uniSMS struct {
	options Options
	client  uniSender
}

var _ sms.SMS = (*uniSMS)(nil)

type uniSender interface {
	Send(message *unisms.UniMessage) (*uniapi.UniResponse, error)
}

// createClient 根据已校验的配置创建 UniSMS 客户端。
func createClient(o *Options) *unisms.UniSMSClient {
	if strings.TrimSpace(o.AccessKeySecret) == "" {
		return unisms.NewClient(o.AccessKeyID)
	}
	return unisms.NewClient(o.AccessKeyID, o.AccessKeySecret)
}

// Send 发送短信消息。
func (s *uniSMS) Send(ctx context.Context, message sms.Message) error {
	if ctx == nil {
		return fmt.Errorf("uni sms: context cannot be nil")
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("uni sms: context error: %w", err)
	}
	if err := validateMessage(message); err != nil {
		return err
	}

	templateData, err := normalizeTemplateData(message.TemplateParam)
	if err != nil {
		return err
	}

	request := unisms.BuildMessage()
	request.SetTo(message.Phone)
	request.SetSignature(s.options.SignName)
	request.SetTemplateId(message.TemplateCode)
	if len(templateData) > 0 {
		request.SetTemplateData(templateData)
	}

	response, err := s.client.Send(request)
	if err != nil {
		return fmt.Errorf("uni sms: send sms: %w", err)
	}
	if response == nil {
		return fmt.Errorf("uni sms: send sms failed: empty response")
	}
	if response.Code != "" && response.Code != "0" {
		return fmt.Errorf("uni sms: send sms failed: code=%s message=%s requestID=%s", response.Code, response.Message, response.RequestId)
	}

	return nil
}

func validateMessage(message sms.Message) error {
	if strings.TrimSpace(message.Phone) == "" {
		return fmt.Errorf("uni sms: phone is required")
	}
	if strings.TrimSpace(message.TemplateCode) == "" {
		return fmt.Errorf("uni sms: template code is required")
	}
	return nil
}

func normalizeTemplateData(param any) (map[string]string, error) {
	if param == nil {
		return nil, nil
	}

	switch v := param.(type) {
	case map[string]string:
		return v, nil
	case map[string]any:
		data := make(map[string]string, len(v))
		for key, value := range v {
			data[key] = fmt.Sprint(value)
		}
		return data, nil
	default:
		return nil, fmt.Errorf("uni sms: template param must be map[string]string or map[string]any")
	}
}
