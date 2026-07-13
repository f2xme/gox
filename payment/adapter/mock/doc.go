// Package mock 提供并发安全、状态可控的内存支付测试实现。
//
// # 功能特性
//
//   - 实现支付、查询、退款和关单接口
//   - 支持状态切换、错误注入和 context 感知延迟
//   - 保存订单、退款和调用记录的独立副本
//   - 生成并解析 mock 支付与退款回调
//
// # 快速开始
//
// 创建自动成功的测试支付服务：
//
//	client, err := mock.New(
//		mock.WithPaymentStatus(payment.PaymentStatusSuccess),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	result, err := client.Pay(context.Background(), &payment.Order{
//		OrderID:   "order-1",
//		Amount:    100,
//		Subject:   "测试商品",
//		NotifyURL: "https://example.com/payment/notify",
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Print(result.PayURL)
//
// mock 仅用于测试，不模拟支付宝或微信的真实签名协议，也不得用于生产收款。
package mock
