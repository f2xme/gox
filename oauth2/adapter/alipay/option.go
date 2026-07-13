package alipay

import (
	"net/http"
	"time"
)

const (
	defaultAuthURL     = "https://openauth.alipay.com/oauth2/publicAppAuthorize.htm"
	defaultGatewayURL  = "https://openapi.alipay.com/gateway.do"
	defaultFormat      = "JSON"
	defaultCharset     = "UTF-8"
	defaultSignType    = "RSA2"
	defaultVersion     = "1.0"
	defaultHTTPTimeout = 10 * time.Second
)

// Options 定义支付宝登录适配器配置。
type Options struct {
	// ClientID 支付宝应用 ID。
	ClientID string
	// PrivateKey 应用私钥 PEM 内容，用于支付宝网关请求签名。
	PrivateKey string
	// AlipayPublicKey 支付宝公钥 PEM 内容，用于验证网关响应签名。
	AlipayPublicKey string
	// RedirectURL 授权回调地址。
	RedirectURL string
	// HTTPClient 用于请求支付宝接口的 HTTP 客户端。
	HTTPClient *http.Client
	// AuthURL 授权地址。
	AuthURL string
	// GatewayURL 支付宝开放平台网关地址。
	GatewayURL string
	// Format 响应格式，默认 JSON。
	Format string
	// Charset 请求字符集，默认 UTF-8。
	Charset string
	// SignType 签名类型，默认 RSA2。
	SignType string
	// Version API 版本，默认 1.0。
	Version string
}

// Option 定义支付宝登录适配器配置函数。
type Option func(*Options)

// defaultOptions 返回默认配置。
func defaultOptions() Options {
	return Options{
		HTTPClient: &http.Client{Timeout: defaultHTTPTimeout},
		AuthURL:    defaultAuthURL,
		GatewayURL: defaultGatewayURL,
		Format:     defaultFormat,
		Charset:    defaultCharset,
		SignType:   defaultSignType,
		Version:    defaultVersion,
	}
}

// WithClientID 设置支付宝应用 ID。
func WithClientID(clientID string) Option {
	return func(o *Options) {
		o.ClientID = clientID
	}
}

// WithPrivateKey 设置应用私钥 PEM 内容。
func WithPrivateKey(privateKey string) Option {
	return func(o *Options) {
		o.PrivateKey = privateKey
	}
}

// WithAlipayPublicKey 设置支付宝公钥 PEM 内容。
func WithAlipayPublicKey(publicKey string) Option {
	return func(o *Options) {
		o.AlipayPublicKey = publicKey
	}
}

// WithRedirectURL 设置授权回调地址。
func WithRedirectURL(redirectURL string) Option {
	return func(o *Options) {
		o.RedirectURL = redirectURL
	}
}

// WithHTTPClient 设置 HTTP 客户端。
func WithHTTPClient(client *http.Client) Option {
	return func(o *Options) {
		if client != nil {
			o.HTTPClient = client
		}
	}
}

// WithEndpoints 设置支付宝授权地址和网关地址，主要用于测试或私有代理。
func WithEndpoints(authURL, gatewayURL string) Option {
	return func(o *Options) {
		if authURL != "" {
			o.AuthURL = authURL
		}
		if gatewayURL != "" {
			o.GatewayURL = gatewayURL
		}
	}
}

// WithSignType 设置签名类型，支持 RSA 和 RSA2。
func WithSignType(signType string) Option {
	return func(o *Options) {
		if signType != "" {
			o.SignType = signType
		}
	}
}
