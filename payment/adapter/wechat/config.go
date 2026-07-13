package wechat

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/f2xme/gox/payment"
)

const (
	defaultOAuthAuthURL  = "https://open.weixin.qq.com/connect/oauth2/authorize"
	defaultOAuthTokenURL = "https://api.weixin.qq.com/sns/oauth2/access_token"
)

// Config 定义微信支付 V3 配置。
type Config struct {
	// AppID 是公众号或应用 ID。
	AppID string
	// OAuthAppSecret 是公众号网页授权密钥。
	OAuthAppSecret string
	// MchID 是微信支付商户号。
	MchID string
	// MerchantSerialNo 是商户 API 证书序列号。
	MerchantSerialNo string
	// MerchantPrivateKey 是商户 API 私钥 PEM 内容。
	MerchantPrivateKey string
	// APIV3Key 是 32 字节 API v3 密钥。
	APIV3Key string
	// WechatPayPublicKey 是微信支付公钥 PEM 内容。
	WechatPayPublicKey string
	// WechatPayPublicKeyID 是微信支付公钥 ID。
	WechatPayPublicKeyID string
}

type options struct {
	timeout       time.Duration
	transport     http.RoundTripper
	logger        *slog.Logger
	oauthAuthURL  string
	oauthTokenURL string
}

// Option 定义微信适配器选项。
type Option func(*options)

func defaultOptions() options {
	return options{timeout: 10 * time.Second, oauthAuthURL: defaultOAuthAuthURL, oauthTokenURL: defaultOAuthTokenURL}
}

// WithTimeout 设置 HTTP 超时。
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) {
		if timeout > 0 {
			o.timeout = timeout
		}
	}
}

// WithHTTPTransport 设置 HTTP transport。
func WithHTTPTransport(transport http.RoundTripper) Option {
	return func(o *options) { o.transport = transport }
}

// WithLogger 设置结构化日志器。
func WithLogger(logger *slog.Logger) Option {
	return func(o *options) { o.logger = logger }
}

func (o options) oauthClient() *http.Client {
	return &http.Client{Timeout: o.timeout, Transport: o.transport}
}

func validateConfig(c Config) error {
	switch {
	case c.AppID == "":
		return fmt.Errorf("%w: wechat app ID cannot be empty", payment.ErrInvalidConfig)
	case c.MchID == "":
		return fmt.Errorf("%w: wechat merchant ID cannot be empty", payment.ErrInvalidConfig)
	case c.MerchantSerialNo == "":
		return fmt.Errorf("%w: wechat merchant serial number cannot be empty", payment.ErrInvalidConfig)
	case c.MerchantPrivateKey == "":
		return fmt.Errorf("%w: wechat merchant private key cannot be empty", payment.ErrInvalidConfig)
	case c.APIV3Key == "":
		return fmt.Errorf("%w: wechat API v3 key cannot be empty", payment.ErrInvalidConfig)
	case c.WechatPayPublicKey == "":
		return fmt.Errorf("%w: wechat pay public key cannot be empty", payment.ErrInvalidConfig)
	case c.WechatPayPublicKeyID == "":
		return fmt.Errorf("%w: wechat pay public key ID cannot be empty", payment.ErrInvalidConfig)
	default:
		return nil
	}
}
