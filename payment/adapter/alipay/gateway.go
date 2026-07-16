package alipay

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-pay/crypto/xpem"
	"github.com/go-pay/gopay"
	aliyun "github.com/go-pay/gopay/alipay"
	"github.com/go-pay/gopay/pkg/xhttp"
)

type gateway interface {
	precreate(context.Context, gopay.BodyMap) (*aliyun.TradePrecreateResponse, error)
	wapPay(context.Context, gopay.BodyMap) (string, error)
	query(context.Context, gopay.BodyMap) (*aliyun.TradeQueryResponse, error)
	refund(context.Context, gopay.BodyMap) (*aliyun.TradeRefundResponse, error)
	close(context.Context, gopay.BodyMap) (*aliyun.TradeCloseResponse, error)
}

type gopayGateway struct{ client *aliyun.Client }

func newGopayGateway(config Config, opts options) (*gopayGateway, error) {
	client, err := aliyun.NewClient(config.AppID, config.PrivateKey, config.IsProduction())
	if err != nil {
		return nil, err
	}

	verifyMaterial, err := configureSignMode(client, config)
	if err != nil {
		return nil, err
	}
	// 预解码校验材料，尽早暴露错误的公钥/证书内容。
	// DecodePublicKey 在 PEM 类型不受支持时会返回 (nil, nil)，须显式拒绝，
	// 否则 AutoVerifySign 会静默关闭同步验签。
	pubKey, err := xpem.DecodePublicKey(verifyMaterial)
	if err != nil {
		return nil, fmt.Errorf("decode alipay verify material: %w", err)
	}
	if pubKey == nil {
		return nil, fmt.Errorf("decode alipay verify material: empty public key")
	}
	client.AutoVerifySign(verifyMaterial)
	// 可选透传：gopay SetAESKey（字符串原样作密钥字节）。上游标注内容加密仍不完整，见 Config.AESKey。
	if aesKey := strings.TrimSpace(config.AESKey); aesKey != "" {
		client.SetAESKey(aesKey)
	}

	httpClient := xhttp.NewClient().SetTimeout(opts.timeout)
	if opts.transport != nil {
		httpClient.SetTransport(opts.transport)
	}
	client.SetHttpClient(httpClient)
	if opts.logger != nil {
		client.SetLogger(slogAdapter{logger: opts.logger})
	}
	return &gopayGateway{client: client}, nil
}

// configureSignMode 按配置选择密钥或证书模式，返回用于同步验签的材料。
func configureSignMode(client *aliyun.Client, config Config) ([]byte, error) {
	if config.useCertMode() {
		if err := client.SetCertSnByContent(
			[]byte(config.AppPublicCert),
			[]byte(config.AlipayRootCert),
			[]byte(config.AlipayPublicCert),
		); err != nil {
			return nil, fmt.Errorf("set alipay cert sn: %w", err)
		}
		return []byte(config.AlipayPublicCert), nil
	}
	return []byte(config.AlipayPublicKey), nil
}

func (g *gopayGateway) precreate(ctx context.Context, bm gopay.BodyMap) (*aliyun.TradePrecreateResponse, error) {
	return g.client.TradePrecreate(ctx, bm)
}
func (g *gopayGateway) wapPay(ctx context.Context, bm gopay.BodyMap) (string, error) {
	return g.client.TradeWapPay(ctx, bm)
}
func (g *gopayGateway) query(ctx context.Context, bm gopay.BodyMap) (*aliyun.TradeQueryResponse, error) {
	return g.client.TradeQuery(ctx, bm)
}
func (g *gopayGateway) refund(ctx context.Context, bm gopay.BodyMap) (*aliyun.TradeRefundResponse, error) {
	return g.client.TradeRefund(ctx, bm)
}
func (g *gopayGateway) close(ctx context.Context, bm gopay.BodyMap) (*aliyun.TradeCloseResponse, error) {
	return g.client.TradeClose(ctx, bm)
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
