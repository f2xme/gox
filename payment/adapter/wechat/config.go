package wechat

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/f2xme/gox/payment"
)

const (
	defaultOAuthAuthURL  = "https://open.weixin.qq.com/connect/oauth2/authorize"
	defaultOAuthTokenURL = "https://api.weixin.qq.com/sns/oauth2/access_token"
)

// VerifyMode 指定商户平台「验证微信支付身份」所用验签方式。
//
// 与商户平台两种方式对齐（只能启用一种）；平台证书再细分为静态与自动拉取。
type VerifyMode string

const (
	// VerifyModePublicKey 使用微信支付公钥验签。
	// 须配置 WechatPayPublicKey + WechatPayPublicKeyID。
	VerifyModePublicKey VerifyMode = "public_key"
	// VerifyModePlatformCertStatic 使用配置中的平台证书 PEM 验签。
	// 须配置 PlatformCert + PlatformCertSerialNo。
	// 仅登记单张证书序列号；证书轮换窗口请优先自动模式。
	VerifyModePlatformCertStatic VerifyMode = "platform_cert_static"
	// VerifyModePlatformCertAuto 启动时通过 API 拉取平台证书并（默认定时）刷新。
	// 须显式设置；不配置公钥与静态证书字段。
	VerifyModePlatformCertAuto VerifyMode = "platform_cert_auto"
)

// Config 定义微信支付 V3 配置。
//
// 支持两种「验证微信支付身份」方式（二选一，与商户平台一致）：
//
//  1. 微信支付公钥：WechatPayPublicKey + WechatPayPublicKeyID
//  2. 平台证书：
//     - 静态：PlatformCert + PlatformCertSerialNo（离线/固定证书，单序列号）
//     - 自动：VerifyMode=platform_cert_auto，启动时 API 拉取并默认每 12 小时刷新
//
// VerifyMode 为空时仅按材料推断公钥或静态证书模式；材料皆空不会静默进入自动拉证，
// 须显式设置 VerifyModePlatformCertAuto（fail-closed，与支付宝 adapter 一致）。
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
	// VerifyMode 指定验签方式。
	// 为空时：公钥齐全 → public_key，静态证书齐全 → platform_cert_static；
	// 材料皆空须显式 platform_cert_auto，否则配置错误。
	// 合法值：public_key、platform_cert_static、platform_cert_auto。
	VerifyMode VerifyMode
	// WechatPayPublicKey 是微信支付公钥 PEM 内容（公钥模式）。
	// 与平台证书模式二选一。
	WechatPayPublicKey string
	// WechatPayPublicKeyID 是微信支付公钥 ID（公钥模式，含 PUB_KEY_ID_ 前缀时请保留）。
	WechatPayPublicKeyID string
	// PlatformCert 是微信平台证书 PEM 内容（平台证书静态模式）。
	// 与 PlatformCertSerialNo 配套。若 PEM 为 CERTIFICATE，初始化时会与序列号交叉校验。
	// 轮换期间微信可能并存两张有效证书；静态模式只登记单序列号，请优先自动模式。
	PlatformCert string
	// PlatformCertSerialNo 是微信平台证书序列号（平台证书静态模式）。
	// 与商户平台展示一致：十六进制大写、可带或不带冒号，比较前会规范化。
	PlatformCertSerialNo string
	// PlatformCertAutoRefresh 控制平台证书自动模式是否定时刷新。
	// 仅平台证书自动模式生效；nil 或 true 为默认开启，false 关闭。
	// 开启时 gopay 会启动后台 goroutine（约 12h），与 client 同生命周期且无法停止；
	// 短生命周期/测试场景请设为 false。在公钥或静态模式下设置会返回配置错误。
	PlatformCertAutoRefresh *bool
}

// ResolveVerifyMode 返回最终生效的验签模式（含空 VerifyMode 时的字段推断）。
// 非法或材料不足无法推断时返回空字符串；经 New 创建时 validateConfig 会拒绝。
func (c Config) ResolveVerifyMode() VerifyMode {
	mode, err := c.resolveVerifyMode()
	if err != nil {
		return ""
	}
	return mode
}

func (c Config) resolveVerifyMode() (VerifyMode, error) {
	switch c.VerifyMode {
	case VerifyModePublicKey, VerifyModePlatformCertStatic, VerifyModePlatformCertAuto:
		return c.VerifyMode, nil
	case "":
		if c.hasPublicKeyPair() {
			return VerifyModePublicKey, nil
		}
		if c.hasPlatformCertPair() {
			return VerifyModePlatformCertStatic, nil
		}
		// fail-closed：禁止材料皆空时静默进入自动拉证。
		return "", fmt.Errorf("%w: wechat verify materials missing: provide public key pair, platform cert pair, or set VerifyMode=%q",
			payment.ErrInvalidConfig, VerifyModePlatformCertAuto)
	default:
		return "", fmt.Errorf("%w: wechat verify mode must be %q, %q or %q",
			payment.ErrInvalidConfig, VerifyModePublicKey, VerifyModePlatformCertStatic, VerifyModePlatformCertAuto)
	}
}

func (c Config) hasPublicKeyPair() bool {
	return c.WechatPayPublicKey != "" && c.WechatPayPublicKeyID != ""
}

func (c Config) hasPlatformCertPair() bool {
	return c.PlatformCert != "" && c.PlatformCertSerialNo != ""
}

// platformCertShouldAutoRefresh 返回平台证书自动模式是否应开启定时刷新。
func (c Config) platformCertShouldAutoRefresh() bool {
	if c.PlatformCertAutoRefresh == nil {
		return true
	}
	return *c.PlatformCertAutoRefresh
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
	}

	hasPubKey := c.WechatPayPublicKey != ""
	hasPubKeyID := c.WechatPayPublicKeyID != ""
	hasPlatCert := c.PlatformCert != ""
	hasPlatSerial := c.PlatformCertSerialNo != ""

	// 公钥字段只配了一部分时给出明确错误。
	if hasPubKey != hasPubKeyID {
		if !hasPubKey {
			return fmt.Errorf("%w: wechat pay public key cannot be empty when public key ID is set", payment.ErrInvalidConfig)
		}
		return fmt.Errorf("%w: wechat pay public key ID cannot be empty when public key is set", payment.ErrInvalidConfig)
	}

	// 平台证书静态字段只配了一部分时给出明确错误。
	if hasPlatCert != hasPlatSerial {
		if !hasPlatCert {
			return fmt.Errorf("%w: wechat platform cert cannot be empty when platform cert serial is set", payment.ErrInvalidConfig)
		}
		return fmt.Errorf("%w: wechat platform cert serial cannot be empty when platform cert is set", payment.ErrInvalidConfig)
	}

	// 公钥材料与平台证书静态材料不可同时完整配置（商户平台亦二选一）。
	if c.hasPublicKeyPair() && c.hasPlatformCertPair() {
		return fmt.Errorf("%w: wechat public key mode and platform cert mode are mutually exclusive", payment.ErrInvalidConfig)
	}

	mode, err := c.resolveVerifyMode()
	if err != nil {
		return err
	}

	switch mode {
	case VerifyModePublicKey:
		if !c.hasPublicKeyPair() {
			return fmt.Errorf("%w: wechat pay public key and public key ID are required in public_key mode", payment.ErrInvalidConfig)
		}
		if c.PlatformCertAutoRefresh != nil {
			return fmt.Errorf("%w: PlatformCertAutoRefresh is only valid in platform_cert_auto mode", payment.ErrInvalidConfig)
		}
	case VerifyModePlatformCertStatic:
		if !c.hasPlatformCertPair() {
			return fmt.Errorf("%w: wechat platform cert and serial are required in platform_cert_static mode", payment.ErrInvalidConfig)
		}
		if c.PlatformCertAutoRefresh != nil {
			return fmt.Errorf("%w: PlatformCertAutoRefresh is only valid in platform_cert_auto mode", payment.ErrInvalidConfig)
		}
	case VerifyModePlatformCertAuto:
		if c.hasPublicKeyPair() {
			return fmt.Errorf("%w: wechat pay public key fields must be empty in platform_cert_auto mode", payment.ErrInvalidConfig)
		}
		if c.hasPlatformCertPair() {
			return fmt.Errorf("%w: wechat platform cert fields must be empty in platform_cert_auto mode (use platform_cert_static for fixed cert)", payment.ErrInvalidConfig)
		}
	}

	return nil
}

// normalizeCertSerial 规范化微信平台证书序列号（去冒号/空白、大写十六进制）。
func normalizeCertSerial(serial string) string {
	s := strings.TrimSpace(serial)
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, " ", "")
	return strings.ToUpper(s)
}

// platformCertSerialFromPEM 从 CERTIFICATE PEM 提取序列号；PUBLIC KEY 等返回空串与 nil。
func platformCertSerialFromPEM(pemContent string) (string, error) {
	block, _ := pem.Decode([]byte(pemContent))
	if block == nil {
		return "", fmt.Errorf("pem decode platform cert failed")
	}
	if block.Type != "CERTIFICATE" {
		// 仅公钥 PEM 时无法交叉校验序列号，交由 gopay 登记。
		return "", nil
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("parse platform certificate: %w", err)
	}
	if cert.SerialNumber == nil {
		return "", fmt.Errorf("platform certificate serial number is empty")
	}
	return normalizeCertSerial(cert.SerialNumber.Text(16)), nil
}

// validatePlatformCertSerial 在静态模式下交叉校验 PEM 证书序列号与配置序列号。
func validatePlatformCertSerial(certPEM, configuredSerial string) error {
	fromPEM, err := platformCertSerialFromPEM(certPEM)
	if err != nil {
		return fmt.Errorf("%w: %v", payment.ErrInvalidConfig, err)
	}
	if fromPEM == "" {
		return nil
	}
	want := normalizeCertSerial(configuredSerial)
	if fromPEM != want {
		return fmt.Errorf("%w: wechat platform cert serial mismatch: pem=%s config=%s",
			payment.ErrInvalidConfig, fromPEM, want)
	}
	return nil
}
