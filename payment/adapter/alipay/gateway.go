package alipay

import (
	"context"
	"fmt"
	"log/slog"

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
	client, err := aliyun.NewClient(config.AppID, config.PrivateKey, config.Production)
	if err != nil {
		return nil, err
	}
	publicKey, err := xpem.DecodePublicKey([]byte(config.AlipayPublicKey))
	if err != nil {
		return nil, fmt.Errorf("decode alipay public key: %w", err)
	}
	if publicKey == nil {
		return nil, fmt.Errorf("decode alipay public key: empty public key")
	}
	client.AutoVerifySign([]byte(config.AlipayPublicKey))
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
