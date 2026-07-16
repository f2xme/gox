package payment_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/f2xme/gox/payment"
	paymock "github.com/f2xme/gox/payment/adapter/mock"
)

// ExampleParseProvider 演示按环境变量风格装配支付渠道。
// 本地默认 mock；生产显式 wechat / alipay。
func ExampleParseProvider() {
	for _, name := range []string{"", "mock", "wechat", "alipay", "wx"} {
		provider, err := payment.ParseProvider(name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(name, "->", provider)
	}
	// Output:
	//  -> mock
	// mock -> mock
	// wechat -> wechat
	// alipay -> alipay
	// wx -> wechat
}

// ExampleValidateOrder 演示下单前统一校验。
func ExampleValidateOrder() {
	err := payment.ValidateOrder(&payment.Order{
		OrderID:   "order-1001",
		Amount:    9900,
		Subject:   "会员订阅",
		NotifyURL: "https://merchant.example/payment/notify",
	})
	fmt.Println(err == nil)
	// Output:
	// true
}

// Example_paymentNotifyHandler 演示依赖 PaymentNotifier 的通用回调 handler。
// 生产注入 alipay/wechat adapter；本地/CI 注入 mock。
func Example_paymentNotifyHandler() {
	client, err := paymock.New()
	if err != nil {
		log.Fatal(err)
	}

	// 先完成一笔本地支付，得到可解析的回调请求。
	_, notify, _, err := client.PayAndDeliver(context.Background(), &payment.Order{
		OrderID:   "order-1001",
		Amount:    9900,
		Subject:   "会员订阅",
		NotifyURL: "https://merchant.example/payment/notify",
	})
	if err != nil {
		log.Fatal(err)
	}

	// 业务 HTTP 层：验签解析 → 事务入账 → 成功后 ACK。
	handler := paymentNotifyHandler(client)
	rec := httptest.NewRecorder()
	// 重新生成一份回调请求供 handler 消费（Body 只能读一次）。
	req, err := client.PaymentNotificationRequest("order-1001")
	if err != nil {
		log.Fatal(err)
	}
	handler.ServeHTTP(rec, req)

	fmt.Println(notify.Status)
	fmt.Println(rec.Code)
	// Output:
	// success
	// 200
}

// paymentNotifyHandler 是可复用的异步支付回调形态。
// notifier 为 alipay.Alipay、wechat.WechatPay 或 mock.Client。
//
// 仅 PaymentStatusSuccess 走支付入账；refunded 应走 RefundNotifier，不要当支付成功。
// 已验签且业务决定不处理的状态仍须写 SuccessResponse（支付宝 body=success / 微信 JSON SUCCESS），
// 否则平台会按失败持续重试。入账事务失败时不要写 SuccessResponse。
func paymentNotifyHandler(notifier payment.PaymentNotifier) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notify, err := notifier.ParsePaymentNotification(r.Context(), r)
		if err != nil {
			http.Error(w, "bad notification", http.StatusBadRequest)
			return
		}

		switch notify.Status {
		case payment.PaymentStatusSuccess:
			// 在数据库事务内：核对订单号、金额并幂等入账。
			// if err := applyPayment(r.Context(), notify); err != nil {
			// 	http.Error(w, "retry later", http.StatusInternalServerError)
			// 	return
			// }
		case payment.PaymentStatusRefunded:
			// 支付通知里的 refunded 不是支付成功入账路径；退款流水走 RefundNotifier。
			// 此处仍 ACK，避免支付回调通道无意义重试。
		default:
			// pending / failed / closed 等：记录日志后 ACK。
		}

		// 回执写入失败时记录日志；响应可能已部分写出，依赖平台重试策略。
		if err := notifier.SuccessResponse().WriteTo(w); err != nil {
			// log.Printf("payment notify ack write failed: %v", err)
			return
		}
	})
}
