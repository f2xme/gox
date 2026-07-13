package onepay

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/f2xme/gox/payment"
)

const (
	defaultPath = "/pay/"
	defaultTTL  = 15 * time.Minute
	maxTTL      = 24 * time.Hour
)

// Checkout 表示 Resolver 持久化并可复用的平台收银台结果。
type Checkout struct {
	// Provider 是收银台所属支付平台。
	Provider payment.Provider
	// OrderID 是平台商户订单号。
	OrderID string
	// WAP 是支付宝手机网站收银台结果。
	WAP *payment.WAPResult
	// JSAPI 是微信 JSAPI 调起参数。
	JSAPI *payment.JSAPIResult
	// ExpiresAt 是收银台结果过期时间。
	ExpiresAt time.Time
}

// CheckoutResolver 在业务持久层幂等地解析或创建完整平台收银台。
type CheckoutResolver interface {
	// ResolveOrCreate 原子创建或复用指定支付意图、平台和付款人的收银台。
	ResolveOrCreate(ctx context.Context, intentID string, provider payment.Provider, payerOpenID string) (*Checkout, error)
}

// WechatOAuth 定义一码付需要的微信网页授权能力。
type WechatOAuth interface {
	// OAuthURL 创建微信网页授权地址。
	OAuthURL(redirectURL, state string) (string, error)
	// ExchangeOAuthCode 使用授权 code 换取 openid。
	ExchangeOAuthCode(ctx context.Context, code string) (string, error)
}

// Config 定义一码付服务配置。
type Config struct {
	// BaseURL 是二维码访问使用的 HTTPS origin。
	BaseURL string
	// Path 是 handler 路径前缀，默认 /pay/。
	Path string
	// TokenKey 是 32 字节 AES-256 密钥。
	TokenKey []byte
	// TokenTTL 是二维码默认有效期，默认 15 分钟。
	TokenTTL time.Duration
	// Resolver 是业务收银台的幂等持久化边界。
	Resolver CheckoutResolver
	// Wechat 提供微信 OAuth 能力。
	Wechat WechatOAuth
}

type options struct{ qrSize int }

// Option 定义一码付服务选项。
type Option func(*options)

// WithQRSize 设置 PNG 宽高，范围 128–1024。
func WithQRSize(size int) Option { return func(o *options) { o.qrSize = size } }

type codeOptions struct{ expiresAt time.Time }

// CodeOption 定义单个码选项。
type CodeOption func(*codeOptions)

// WithExpiresAt 覆盖单个码的过期时间。
func WithExpiresAt(expiresAt time.Time) CodeOption {
	return func(o *codeOptions) { o.expiresAt = expiresAt }
}

func normalizeConfig(c Config) (Config, *url.URL, error) {
	if c.Path == "" {
		c.Path = defaultPath
	}
	if c.TokenTTL == 0 {
		c.TokenTTL = defaultTTL
	}
	u, err := url.Parse(c.BaseURL)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return c, nil, fmt.Errorf("%w: onepay base URL must be an origin", payment.ErrInvalidConfig)
	}
	if u.Scheme != "https" && !(u.Scheme == "http" && isLocalhost(u.Hostname())) {
		return c, nil, fmt.Errorf("%w: onepay base URL must use HTTPS", payment.ErrInvalidConfig)
	}
	if !strings.HasPrefix(c.Path, "/") || !strings.HasSuffix(c.Path, "/") {
		return c, nil, fmt.Errorf("%w: onepay path must start and end with slash", payment.ErrInvalidConfig)
	}
	for _, segment := range strings.Split(c.Path, "/") {
		if segment == "." || segment == ".." {
			return c, nil, fmt.Errorf("%w: unsafe onepay path", payment.ErrInvalidConfig)
		}
	}
	if len(c.TokenKey) != 32 {
		return c, nil, fmt.Errorf("%w: onepay token key must be 32 bytes", payment.ErrInvalidConfig)
	}
	if c.TokenTTL <= 0 || c.TokenTTL > maxTTL {
		return c, nil, fmt.Errorf("%w: onepay token TTL must be within 24 hours", payment.ErrInvalidConfig)
	}
	if c.Resolver == nil || c.Wechat == nil {
		return c, nil, fmt.Errorf("%w: onepay dependencies cannot be nil", payment.ErrInvalidConfig)
	}
	u.Path = ""
	c.TokenKey = append([]byte(nil), c.TokenKey...)
	return c, u, nil
}

func isLocalhost(host string) bool {
	ip := net.ParseIP(host)
	return strings.EqualFold(host, "localhost") || (ip != nil && ip.IsLoopback())
}
