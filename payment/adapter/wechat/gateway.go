package wechat

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/f2xme/gox/payment"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/pkg/xhttp"
	wx "github.com/go-pay/gopay/wechat/v3"
)

// 回调补拉抗滥用参数。
// AutoVerifySign 会全量拉取证书，一次成功刷新可供所有 serial 复用；
// 刷新成功但目标 serial 仍缺失时进入负向缓存，避免伪造 serial 打穿出站。
const (
	platformCertRefreshCooldown = 60 * time.Second
	platformCertMissTTL         = 5 * time.Minute
)

type gateway interface {
	native(context.Context, gopay.BodyMap) (*wx.NativeRsp, error)
	jsapi(context.Context, gopay.BodyMap) (*wx.PrepayRsp, error)
	paySign(string, string) (*wx.JSAPIPayParams, error)
	query(context.Context, string) (*wx.QueryOrderRsp, error)
	refund(context.Context, gopay.BodyMap) (*wx.RefundRsp, error)
	close(context.Context, string) (*wx.EmptyRsp, error)
	publicKeys() map[string]*rsa.PublicKey
	// ensurePublicKey 在自动模式下若 serial 未登记则受控补拉平台证书（不启动定时刷新）。
	// 公钥/静态模式为 no-op。补拉带互斥合流、全局冷却与 serial 负向缓存。
	ensurePublicKey(serial string) error
}

type gopayGateway struct {
	client   *wx.ClientV3
	autoMode bool

	// refreshMu 串行化补拉：并发回调在锁上等待，复用同一次网络拉取。
	refreshMu     sync.Mutex
	lastRefreshAt time.Time
	// missUntil 记录「已全量刷新但 serial 仍不存在」的负向缓存截止时间。
	missUntil map[string]time.Time
}

func newGopayGateway(c Config, opts options) (*gopayGateway, error) {
	client, err := wx.NewClientV3(c.MchID, c.MerchantSerialNo, c.APIV3Key, c.MerchantPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("%w: create wechat v3 client: %w", payment.ErrInvalidConfig, err)
	}
	// 平台证书自动拉取会发网关请求，须先配置 HTTP 客户端（超时/传输层）。
	hc := xhttp.NewClient().SetTimeout(opts.timeout)
	if opts.transport != nil {
		hc.SetTransport(opts.transport)
	}
	client.SetHttpClient(hc)
	if opts.logger != nil {
		client.SetLogger(slogAdapter{logger: opts.logger})
	}
	mode, err := c.resolveVerifyMode()
	if err != nil {
		return nil, err
	}
	if err := configureVerifyMode(client, c, mode); err != nil {
		return nil, err
	}
	return &gopayGateway{
		client:    client,
		autoMode:  mode == VerifyModePlatformCertAuto,
		missUntil: make(map[string]time.Time),
	}, nil
}

// configureVerifyMode 按配置选择微信支付公钥或平台证书验签。
//
// PEM/参数类失败返回 payment.ErrInvalidConfig；
// 平台证书自动拉取以网络类为主时返回 payment.ErrGateway，可识别的配置/权限类失败返回 ErrInvalidConfig。
func configureVerifyMode(client *wx.ClientV3, c Config, mode VerifyMode) error {
	switch mode {
	case VerifyModePublicKey:
		// 微信支付公钥验签（商户平台「微信支付公钥」）。
		if err := client.AutoVerifySignByPublicKey([]byte(c.WechatPayPublicKey), c.WechatPayPublicKeyID); err != nil {
			return fmt.Errorf("%w: configure wechat pay public key verify: %w", payment.ErrInvalidConfig, err)
		}
	case VerifyModePlatformCertStatic:
		if err := validatePlatformCertSerial(c.PlatformCert, c.PlatformCertSerialNo); err != nil {
			return err
		}
		// 平台证书静态模式：登记单张证书，不自动刷新。
		// gopay 仍以 AutoVerifySignByCert 作为静态证书/公钥登记入口（符号已 Deprecated，行为仍正确）。
		serial := normalizeCertSerial(c.PlatformCertSerialNo)
		if err := client.AutoVerifySignByCert([]byte(c.PlatformCert), serial); err != nil {
			return fmt.Errorf("%w: configure wechat platform cert verify: %w", payment.ErrInvalidConfig, err)
		}
	case VerifyModePlatformCertAuto:
		// 平台证书自动模式：启动时拉取全量有效证书；默认每 12 小时刷新（后台 goroutine，无法停止）。
		if err := client.AutoVerifySign(c.platformCertShouldAutoRefresh()); err != nil {
			return classifyPlatformCertAutoError(err)
		}
	default:
		return fmt.Errorf("%w: unsupported wechat verify mode %q", payment.ErrInvalidConfig, mode)
	}
	return nil
}

// classifyPlatformCertAutoError 将自动拉证失败粗分为配置/权限 vs 网关。
// gopay 错误类型有限，启发式匹配；无法区分时归 ErrGateway，并在文档中说明可能仍需检查密钥与平台证书权限。
func classifyPlatformCertAutoError(err error) error {
	msg := strings.ToLower(err.Error())
	configHints := []string{
		"decrypt", "apiv3", "api v3", "private key", "privatekey",
		"unauthorized", "forbidden", "permission", "invalid",
		"401", "403",
	}
	for _, h := range configHints {
		if strings.Contains(msg, h) {
			return fmt.Errorf("%w: configure wechat platform cert auto verify: %w", payment.ErrInvalidConfig, err)
		}
	}
	return fmt.Errorf("%w: configure wechat platform cert auto verify: %w", payment.ErrGateway, err)
}

func (g *gopayGateway) native(ctx context.Context, bm gopay.BodyMap) (*wx.NativeRsp, error) {
	return g.client.V3TransactionNative(ctx, bm)
}
func (g *gopayGateway) jsapi(ctx context.Context, bm gopay.BodyMap) (*wx.PrepayRsp, error) {
	return g.client.V3TransactionJsapi(ctx, bm)
}
func (g *gopayGateway) paySign(appID, prepayID string) (*wx.JSAPIPayParams, error) {
	return g.client.PaySignOfJSAPI(appID, prepayID)
}
func (g *gopayGateway) query(ctx context.Context, id string) (*wx.QueryOrderRsp, error) {
	return g.client.V3TransactionQueryOrder(ctx, wx.OutTradeNo, id)
}
func (g *gopayGateway) refund(ctx context.Context, bm gopay.BodyMap) (*wx.RefundRsp, error) {
	return g.client.V3Refund(ctx, bm)
}
func (g *gopayGateway) close(ctx context.Context, id string) (*wx.EmptyRsp, error) {
	return g.client.V3TransactionCloseOrder(ctx, id)
}
func (g *gopayGateway) publicKeys() map[string]*rsa.PublicKey { return g.client.WxPublicKeyMap() }

func (g *gopayGateway) ensurePublicKey(serial string) error {
	if !g.autoMode || serial == "" {
		return nil
	}
	// 公钥模式序列号不会通过本路径补拉。
	if strings.HasPrefix(serial, "PUB_KEY_ID") {
		return nil
	}
	if _, ok := g.client.WxPublicKeyMap()[serial]; ok {
		return nil
	}

	// 持锁补拉：并发回调排队，后到者复用同一次全量拉取结果（合流 + 冷却）。
	g.refreshMu.Lock()
	defer g.refreshMu.Unlock()

	if _, ok := g.client.WxPublicKeyMap()[serial]; ok {
		return nil
	}

	now := time.Now()
	if skipPlatformCertRefresh(now, serial, g.lastRefreshAt, g.missUntil, platformCertRefreshCooldown) {
		return nil
	}

	err := g.client.AutoVerifySign(false)
	g.lastRefreshAt = time.Now()
	pruneMissCache(g.missUntil, g.lastRefreshAt)
	if err != nil {
		return classifyPlatformCertAutoError(err)
	}
	if _, ok := g.client.WxPublicKeyMap()[serial]; !ok {
		// 全量刷新后仍无此 serial：负向缓存，避免伪造 serial 反复打穿。
		g.missUntil[serial] = g.lastRefreshAt.Add(platformCertMissTTL)
	} else {
		delete(g.missUntil, serial)
	}
	return nil
}

// skipPlatformCertRefresh 判断是否应跳过网络补拉（冷却或 serial 负向缓存）。
// hasKey 已在调用方检查；此处只处理时间窗。
func skipPlatformCertRefresh(now time.Time, serial string, lastRefreshAt time.Time, missUntil map[string]time.Time, cooldown time.Duration) bool {
	if until, ok := missUntil[serial]; ok && now.Before(until) {
		return true
	}
	if !lastRefreshAt.IsZero() && now.Sub(lastRefreshAt) < cooldown {
		return true
	}
	return false
}

func pruneMissCache(missUntil map[string]time.Time, now time.Time) {
	for s, until := range missUntil {
		if !now.Before(until) {
			delete(missUntil, s)
		}
	}
}

type slogAdapter struct{ logger *slog.Logger }

func (l slogAdapter) Debug(args ...any)            { l.log(slog.LevelDebug, fmt.Sprint(args...)) }
func (l slogAdapter) Info(args ...any)             { l.log(slog.LevelInfo, fmt.Sprint(args...)) }
func (l slogAdapter) Warn(args ...any)             { l.log(slog.LevelWarn, fmt.Sprint(args...)) }
func (l slogAdapter) Error(args ...any)            { l.log(slog.LevelError, fmt.Sprint(args...)) }
func (l slogAdapter) Debugf(f string, args ...any) { l.log(slog.LevelDebug, fmt.Sprintf(f, args...)) }
func (l slogAdapter) Infof(f string, args ...any)  { l.log(slog.LevelInfo, fmt.Sprintf(f, args...)) }
func (l slogAdapter) Warnf(f string, args ...any)  { l.log(slog.LevelWarn, fmt.Sprintf(f, args...)) }
func (l slogAdapter) Errorf(f string, args ...any) { l.log(slog.LevelError, fmt.Sprintf(f, args...)) }
func (l slogAdapter) log(level slog.Level, msg string) {
	if l.logger != nil {
		l.logger.Log(context.Background(), level, msg)
	}
}
