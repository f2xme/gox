package alipay

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/f2xme/gox/payment"
)

// Config 定义支付宝适配器配置。
type Config struct {
	// AppID 是支付宝应用 ID。
	AppID string
	// SellerID 是支付宝商户 PID，用于校验回调收款方。
	SellerID string
	// PrivateKey 是应用 RSA 私钥 PEM 内容。
	PrivateKey string
	// AlipayPublicKey 是支付宝 RSA 公钥 PEM 内容。
	AlipayPublicKey string
	// Production 表示是否访问生产环境。
	Production bool
}

type options struct {
	timeout   time.Duration
	transport http.RoundTripper
	logger    *slog.Logger
}

// Option 定义支付宝适配器选项。
type Option func(*options)

func defaultOptions() options {
	return options{timeout: 10 * time.Second}
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

func validateConfig(config Config) error {
	switch {
	case config.AppID == "":
		return fmt.Errorf("%w: alipay app ID cannot be empty", payment.ErrInvalidConfig)
	case config.SellerID == "":
		return fmt.Errorf("%w: alipay seller ID cannot be empty", payment.ErrInvalidConfig)
	case config.PrivateKey == "":
		return fmt.Errorf("%w: alipay private key cannot be empty", payment.ErrInvalidConfig)
	case config.AlipayPublicKey == "":
		return fmt.Errorf("%w: alipay public key cannot be empty", payment.ErrInvalidConfig)
	default:
		return nil
	}
}
