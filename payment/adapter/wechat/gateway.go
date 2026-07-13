package wechat

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log/slog"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/pkg/xhttp"
	wx "github.com/go-pay/gopay/wechat/v3"
)

type gateway interface {
	native(context.Context, gopay.BodyMap) (*wx.NativeRsp, error)
	jsapi(context.Context, gopay.BodyMap) (*wx.PrepayRsp, error)
	paySign(string, string) (*wx.JSAPIPayParams, error)
	query(context.Context, string) (*wx.QueryOrderRsp, error)
	refund(context.Context, gopay.BodyMap) (*wx.RefundRsp, error)
	close(context.Context, string) (*wx.EmptyRsp, error)
	publicKeys() map[string]*rsa.PublicKey
}

type gopayGateway struct{ client *wx.ClientV3 }

func newGopayGateway(c Config, opts options) (*gopayGateway, error) {
	client, err := wx.NewClientV3(c.MchID, c.MerchantSerialNo, c.APIV3Key, c.MerchantPrivateKey)
	if err != nil {
		return nil, err
	}
	if err := client.AutoVerifySignByPublicKey([]byte(c.WechatPayPublicKey), c.WechatPayPublicKeyID); err != nil {
		return nil, err
	}
	hc := xhttp.NewClient().SetTimeout(opts.timeout)
	if opts.transport != nil {
		hc.SetTransport(opts.transport)
	}
	client.SetHttpClient(hc)
	if opts.logger != nil {
		client.SetLogger(slogAdapter{logger: opts.logger})
	}
	return &gopayGateway{client: client}, nil
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
