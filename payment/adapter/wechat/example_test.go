package wechat_test

import (
	"fmt"

	"github.com/f2xme/gox/payment"
	"github.com/f2xme/gox/payment/adapter/wechat"
)

// ExampleNew 演示微信支付公钥验签模式构造（凭证请替换为真实材料）。
// 本示例不连接网关，仅展示推荐配置形态。
func ExampleNew() {
	_ = wechat.VerifyModePublicKey
	fmt.Println("wechat.New(wechat.Config{")
	fmt.Println("  AppID, MchID, MerchantSerialNo, MerchantPrivateKey, APIV3Key,")
	fmt.Println("  WechatPayPublicKey, WechatPayPublicKeyID,")
	fmt.Println("})")
	// Output:
	// wechat.New(wechat.Config{
	//   AppID, MchID, MerchantSerialNo, MerchantPrivateKey, APIV3Key,
	//   WechatPayPublicKey, WechatPayPublicKeyID,
	// })
}

// ExampleNew_platformCertAuto 演示平台证书自动拉证模式。
func ExampleNew_platformCertAuto() {
	_ = wechat.VerifyModePlatformCertAuto
	fmt.Println("VerifyMode: wechat.VerifyModePlatformCertAuto")
	fmt.Println("New 时会请求微信证书接口；生产建议单例")
	// Output:
	// VerifyMode: wechat.VerifyModePlatformCertAuto
	// New 时会请求微信证书接口；生产建议单例
}

// Example_payAndNotify 演示 Native 下单与支付/退款回调步骤。
// 可运行的通用支付回调形态见 payment.Example_paymentNotifyHandler。
func Example_payAndNotify() {
	fmt.Println("1. result, err := client.Pay(ctx, order)  // PayURL = code_url")
	fmt.Println("2. notify, err := client.ParsePaymentNotification(ctx, r)")
	fmt.Println("3. // 仅 success：事务内核对 OrderID/Amount 并幂等入账")
	fmt.Println("4. client.SuccessResponse().WriteTo(w)  // JSON SUCCESS")
	fmt.Println("5. refundNotify, err := client.ParseRefundNotification(ctx, r)")
	fmt.Println("6. // 退款流水入账（非支付成功账）")
	fmt.Println("7. client.SuccessResponse().WriteTo(w)  // 与支付回调同一回执格式")
	// Output:
	// 1. result, err := client.Pay(ctx, order)  // PayURL = code_url
	// 2. notify, err := client.ParsePaymentNotification(ctx, r)
	// 3. // 仅 success：事务内核对 OrderID/Amount 并幂等入账
	// 4. client.SuccessResponse().WriteTo(w)  // JSON SUCCESS
	// 5. refundNotify, err := client.ParseRefundNotification(ctx, r)
	// 6. // 退款流水入账（非支付成功账）
	// 7. client.SuccessResponse().WriteTo(w)  // 与支付回调同一回执格式
}

// Example_jsapiForOnepay 演示一码付微信链路需要的 OAuth + JSAPI。
func Example_jsapiForOnepay() {
	_ = payment.ProviderWechat
	fmt.Println("1. url, err := client.OAuthURL(redirectURL, state)")
	fmt.Println("2. openID, err := client.ExchangeOAuthCode(ctx, code)")
	fmt.Println("3. jsapi, err := client.JSAPIPay(ctx, order, openID)")
	// Output:
	// 1. url, err := client.OAuthURL(redirectURL, state)
	// 2. openID, err := client.ExchangeOAuthCode(ctx, code)
	// 3. jsapi, err := client.JSAPIPay(ctx, order, openID)
}

// Example_notifyHTTPHandler 演示把支付与退款回调接到标准库 HTTP。
// 支付侧可运行 handler 见 payment.Example_paymentNotifyHandler；退款侧同理注入 RefundNotifier。
func Example_notifyHTTPHandler() {
	fmt.Println("POST /payment/wechat/notify → ParsePayment → ledger(success only) → SuccessResponse")
	fmt.Println("POST /payment/wechat/refund-notify → ParseRefund → refund ledger → SuccessResponse")
	// Output:
	// POST /payment/wechat/notify → ParsePayment → ledger(success only) → SuccessResponse
	// POST /payment/wechat/refund-notify → ParseRefund → refund ledger → SuccessResponse
}
