package alipay_test

import (
	"fmt"

	"github.com/f2xme/gox/payment/adapter/alipay"
)

// ExampleNew 演示密钥模式 + 沙箱构造（凭证请替换为真实 PEM / 沙箱 AppID）。
// 本示例不连接网关，仅展示推荐配置形态。
func ExampleNew() {
	fmt.Println("Environment:", alipay.EnvSandbox)
	fmt.Println("alipay.New(Config{AppID, SellerID, PrivateKey, AlipayPublicKey, Environment})")
	// Output:
	// Environment: sandbox
	// alipay.New(Config{AppID, SellerID, PrivateKey, AlipayPublicKey, Environment})
}

// ExampleNew_certMode 演示证书模式 + 正式环境构造。
func ExampleNew_certMode() {
	fmt.Println("Environment:", alipay.EnvProduction)
	fmt.Println("cert mode: PrivateKey + AppPublicCert + AlipayRootCert + AlipayPublicCert")
	// Output:
	// Environment: production
	// cert mode: PrivateKey + AppPublicCert + AlipayRootCert + AlipayPublicCert
}

// Example_payAndNotify 演示当面付下单与异步回调处理步骤。
// 生产代码将 client 换为 alipay.New(...) 的返回值。
// 可运行的通用回调形态见 payment 包 Example_paymentNotifyHandler。
func Example_payAndNotify() {
	fmt.Println("1. result, err := client.Pay(ctx, order)  // PayURL = qr_code")
	fmt.Println("2. notify, err := client.ParsePaymentNotification(ctx, r)")
	fmt.Println("3. // 仅 success：事务内核对 OrderID/Amount 并幂等入账")
	fmt.Println("4. client.SuccessResponse().WriteTo(w)  // body: success")
	// Output:
	// 1. result, err := client.Pay(ctx, order)  // PayURL = qr_code
	// 2. notify, err := client.ParsePaymentNotification(ctx, r)
	// 3. // 仅 success：事务内核对 OrderID/Amount 并幂等入账
	// 4. client.SuccessResponse().WriteTo(w)  // body: success
}

// Example_notifyHTTPHandler 演示把支付回调接到标准库 HTTP 的步骤。
// 完整可运行 handler 见 payment.Example_paymentNotifyHandler（注入 alipay.Alipay 即可）。
func Example_notifyHTTPHandler() {
	fmt.Println("POST /payment/alipay/notify")
	fmt.Println("ParsePaymentNotification → ledger(success only) → SuccessResponse")
	fmt.Println("runnable: payment.Example_paymentNotifyHandler")
	// Output:
	// POST /payment/alipay/notify
	// ParsePaymentNotification → ledger(success only) → SuccessResponse
	// runnable: payment.Example_paymentNotifyHandler
}
