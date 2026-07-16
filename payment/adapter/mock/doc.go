// Package mock 提供并发安全、状态可控的内存支付测试实现。
//
// # 功能特性
//
//   - 实现支付、查询、退款和关单接口
//   - 支持状态切换、错误注入和 context 感知延迟
//   - 保存订单、退款和调用记录的独立副本
//   - 生成并解析 mock 支付与退款回调
//   - 本地一键完成支付/退款并投递回调（无需真实商户配置与出网）
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
// # 本地无真实配置时怎么测
//
// 推荐分层：
//
//  1. 业务流 / CI / 本地服务：使用本 mock，依赖 payment.Payment 等接口注入
//  2. adapter 协议细节：在 wechat/alipay 包内用自签密钥与假 HTTP 单测（不要求出网）
//  3. 联调验收：再使用沙箱或测试商户 + 公网回调
//
// 本地完整链路（下单 → 模拟用户付款 → 业务入账）：
//
//	client, _ := mock.New() // 默认 pending，贴近异步回调时序
//	result, notify, resp, err := client.PayAndDeliver(ctx, order)
//	// 1) 持久化 result.PayURL 供前端展示
//	// 2) 用 notify 做金额核对与幂等入账
//	// 3) 入账成功后 WriteTo(resp)
//
// 分步控制：
//
//	_, _ = client.Pay(ctx, order)
//	req, _ := client.CompletePayment(order.OrderID)          // pending→success；已 refunded 保持 refunded
//	notify, _ := client.ParsePaymentNotification(ctx, req) // 或 DeliverPaymentNotification
//
// 环境切换（业务装配示例；解析渠道用核心包，生产不必 import mock）：
//
//	provider, _ := payment.ParseProvider(os.Getenv("PAYMENT_PROVIDER"))
//	switch provider {
//	case payment.ProviderMock:
//		return mock.NewForProvider(provider)
//	case payment.ProviderWechat:
//		return wechat.New(wechat.Config{...}) // 真实密钥
//	case payment.ProviderAlipay:
//		return alipay.New(alipay.Config{...})
//	}
//
// 本地默认 PAYMENT_PROVIDER=mock 或留空；生产显式 wechat/alipay。
//
// # 注意事项
//
//   - mock 不模拟支付宝/微信真实签名协议，不得用于生产收款
//   - 回调解析成功不代表业务入账成功；须在事务内核对订单与金额并幂等处理
//   - CompletePayment 对 refunded 订单不会改回 success；通知 Status 为当前订单状态
//   - SuccessResponse 是 HTTP 回执成功，与通知业务 Status 无关
//   - NewForProvider 仅支持 mock，且强制 Provider=mock；误传 wechat/alipay 返回 ErrInvalidConfig
package mock
