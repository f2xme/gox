package alipay

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/f2xme/gox/payment"
)

// Environment 表示支付宝开放平台网关环境。
type Environment string

const (
	// EnvProduction 正式环境，网关为 https://openapi.alipay.com/gateway.do。
	EnvProduction Environment = "production"
	// EnvSandbox 沙箱环境，网关为 https://openapi-sandbox.dl.alipaydev.com/gateway.do。
	// 须使用开放平台新版沙箱应用及对应密钥/证书，勿与正式环境凭证混用。
	EnvSandbox Environment = "sandbox"
)

// 网关地址为与 go-pay 内部 baseUrl / sandboxBaseUrl 对齐的只读镜像，
// 并非底层 client 的实时取值；升级 go-pay 时需核对是否漂移。
const (
	gatewayURLProduction = "https://openapi.alipay.com/gateway.do"
	gatewayURLSandbox    = "https://openapi-sandbox.dl.alipaydev.com/gateway.do"
)

// Config 定义支付宝适配器配置。
//
// 支持两种加签/验签模式（二选一）：
//
//  1. 密钥模式：配置 PrivateKey + AlipayPublicKey
//  2. 证书模式：配置 PrivateKey + AppPublicCert + AlipayRootCert + AlipayPublicCert
//
// 若证书三件套齐全，优先使用证书模式；AlipayPublicKey 可同时保留但不会参与证书模式验签。
//
// 环境选择：
//
//   - 推荐设置 Environment 为 EnvProduction 或 EnvSandbox
//   - Environment 为空时回退到 Production：true 正式环境，false 沙箱
//   - 零值配置默认走沙箱，避免误连正式网关
type Config struct {
	// AppID 是支付宝应用 ID。
	AppID string
	// SellerID 是支付宝商户 PID，用于校验回调收款方。
	SellerID string
	// PrivateKey 是应用 RSA 私钥 PEM 内容。
	PrivateKey string
	// AlipayPublicKey 是支付宝 RSA 公钥 PEM 内容（密钥模式）。
	// 与证书模式二选一；证书三件套齐全时优先使用证书模式。
	AlipayPublicKey string
	// AppPublicCert 是应用公钥证书内容（证书模式，appCertPublicKey_*.crt）。
	AppPublicCert string
	// AlipayRootCert 是支付宝根证书内容（证书模式，alipayRootCert.crt）。
	AlipayRootCert string
	// AlipayPublicCert 是支付宝公钥证书内容（证书模式，alipayCertPublicKey_RSA2.crt）。
	AlipayPublicCert string
	// AESKey 是透传给 gopay 的接口内容加密密钥（可选）。
	//
	// 非空时调用 gopay Client.SetAESKey；gopay 将字符串**原样**作为 AES 密钥字节
	// （不做 Base64 解码），长度须为 16 / 24 / 32。
	// 开放平台下发的 Base64 串常见为 24 字符（对应 16 字节密钥的编码），可直接填入，
	// 但与官方「先 Base64 解码再作 AES key」的语义不同，且 gopay v1.5.x 标注
	// SetAESKey「目前不可用，设置后会报错」（缺少 encrypt_type、响应不自动解密等）。
	//
	// 因此本字段为实验性透传：勿默认等同于已支持开放平台「接口内容加密」生产能力。
	// 未开启内容加密时请留空。使用前须在目标 gopay 版本上自测。
	AESKey string
	// Environment 指定网关环境，推荐使用 EnvProduction 或 EnvSandbox。
	// 为空时回退到 Production 字段。
	Environment Environment
	// Production 表示是否访问生产环境。
	//
	// Deprecated: 请使用 Environment（EnvProduction / EnvSandbox）。
	// 仅当 Environment 为空时生效：true 正式环境，false 沙箱。
	Production bool
}

// ResolveEnvironment 返回最终生效的网关环境。
//
// 非法 Environment 值回退到 Production 字段（与 GatewayBaseURL 语义一致）；
// 经 New 创建时 validateConfig 会直接拒绝非法值。
func (c Config) ResolveEnvironment() Environment {
	switch c.Environment {
	case EnvProduction, EnvSandbox:
		return c.Environment
	default:
		// 含空字符串与非法值：统一按 Production 回退，避免
		// IsProduction/IsSandbox/GatewayBaseURL 语义不一致。
		if c.Production {
			return EnvProduction
		}
		return EnvSandbox
	}
}

// IsProduction 返回是否使用正式环境网关。
func (c Config) IsProduction() bool {
	return c.ResolveEnvironment() == EnvProduction
}

// IsSandbox 返回是否使用沙箱环境网关。
func (c Config) IsSandbox() bool {
	return c.ResolveEnvironment() == EnvSandbox
}

// GatewayBaseURL 返回当前配置对应的支付宝网关地址。
// 该值为与 go-pay 约定一致的配置层镜像，非底层 HTTP client 的实时 URL。
func (c Config) GatewayBaseURL() string {
	if c.IsProduction() {
		return gatewayURLProduction
	}
	return gatewayURLSandbox
}

// useCertMode 判断是否启用证书模式（三件证书均已配置）。
func (c Config) useCertMode() bool {
	return c.AppPublicCert != "" && c.AlipayRootCert != "" && c.AlipayPublicCert != ""
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
	case config.Environment != "" && config.Environment != EnvProduction && config.Environment != EnvSandbox:
		return fmt.Errorf("%w: alipay environment must be %q or %q", payment.ErrInvalidConfig, EnvProduction, EnvSandbox)
	}

	hasAppCert := config.AppPublicCert != ""
	hasRootCert := config.AlipayRootCert != ""
	hasPublicCert := config.AlipayPublicCert != ""
	certCount := 0
	if hasAppCert {
		certCount++
	}
	if hasRootCert {
		certCount++
	}
	if hasPublicCert {
		certCount++
	}

	// 证书字段只配了一部分时给出明确错误，避免静默回退到密钥模式。
	if certCount > 0 && certCount < 3 {
		switch {
		case !hasAppCert:
			return fmt.Errorf("%w: alipay app public cert cannot be empty in cert mode", payment.ErrInvalidConfig)
		case !hasRootCert:
			return fmt.Errorf("%w: alipay root cert cannot be empty in cert mode", payment.ErrInvalidConfig)
		default:
			return fmt.Errorf("%w: alipay public cert cannot be empty in cert mode", payment.ErrInvalidConfig)
		}
	}

	if err := validateAESKey(config.AESKey); err != nil {
		return err
	}

	if config.useCertMode() {
		return nil
	}
	if config.AlipayPublicKey == "" {
		return fmt.Errorf("%w: alipay public key or full certificate set is required", payment.ErrInvalidConfig)
	}
	return nil
}

// validateAESKey 校验 AESKey：空表示不启用；非空须为 gopay 可用的 16/24/32 字节密钥串。
func validateAESKey(aesKey string) error {
	key := strings.TrimSpace(aesKey)
	if key == "" {
		return nil
	}
	// gopay encryptBizContent 使用 []byte(aesKey) 直接作为 cipher key，须满足 AES 密钥长度。
	switch len(key) {
	case 16, 24, 32:
		return nil
	default:
		return fmt.Errorf("%w: alipay AESKey length must be 16, 24 or 32 (gopay uses raw string bytes as AES key, no Base64 decode)", payment.ErrInvalidConfig)
	}
}
