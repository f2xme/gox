package mock_test

import (
	"context"
	"fmt"
	"log"

	"github.com/f2xme/gox/payment"
	paymock "github.com/f2xme/gox/payment/adapter/mock"
)

// ExampleClient_PayAndDeliver 演示无真实支付配置时的本地完整链路：下单 → 模拟付款 → 解析回调。
func ExampleClient_PayAndDeliver() {
	client, err := paymock.New()
	if err != nil {
		log.Fatal(err)
	}

	result, notify, resp, err := client.PayAndDeliver(context.Background(), &payment.Order{
		OrderID:   "order-1001",
		Amount:    9900,
		Subject:   "会员订阅",
		NotifyURL: "https://merchant.example/payment/notify",
	})
	if err != nil {
		log.Fatal(err)
	}

	// 业务侧：在事务内核对订单金额并幂等入账，成功后再使用 resp。
	fmt.Println(result.OrderID)
	fmt.Println(notify.Status)
	fmt.Println(resp.StatusCode)
	// Output:
	// order-1001
	// success
	// 200
}

// ExampleNewForProvider 演示本地默认装配 mock 渠道。
// 业务侧可用 payment.ParseProvider(os.Getenv("PAYMENT_PROVIDER")) 做 switch；
// 示例写死 mock，保证 go test 不依赖环境变量。
func ExampleNewForProvider() {
	// 业务代码示例：provider, err := payment.ParseProvider(os.Getenv("PAYMENT_PROVIDER"))
	provider, err := payment.ParseProvider("mock")
	if err != nil {
		log.Fatal(err)
	}

	switch provider {
	case payment.ProviderMock:
		client, err := paymock.NewForProvider(provider)
		if err != nil {
			log.Fatal(err)
		}
		var _ payment.Payment = client
		fmt.Println(provider)
	case payment.ProviderWechat, payment.ProviderAlipay:
		// return wechat.New(...) / alipay.New(...) with real credentials
		fmt.Println("wire real adapter")
	}
	// Output:
	// mock
}
