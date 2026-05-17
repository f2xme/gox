/*
Package payment 提供统一的支付操作抽象层。

payment 包只定义支付领域的公共接口和数据结构。具体支付服务商适配器位于
payment/adapter/alipay 和 payment/adapter/wechat 子包中；当前仓库内的
支付宝和微信支付适配器仍是占位实现，不会连接真实支付网关，调用支付操作
会返回 ErrNotImplemented。

# 功能特性

  - 定义统一的支付、查询、退款和关单接口
  - 使用分为单位的 int64 金额，避免浮点精度问题
  - 通过 Extra 字段保留服务商专有参数
  - 提供订单和退款请求校验函数
  - 将真实服务商接入隔离到 adapter 子包

# 快速开始

业务代码依赖 Payment 接口即可隔离具体支付服务商：

	type Payment interface {
		Pay(order *Order) (*PaymentResult, error)
		Query(orderID string) (*QueryResult, error)
		Refund(req *RefundRequest) (*RefundResult, error)
		Close(orderID string) error
	}

	var p Payment
	result, err := p.Pay(&Order{
		OrderID: "order-1001",
		Amount:  9900,
		Subject: "会员订阅",
	})
	_ = result
	_ = err

# 使用建议

业务代码可以依赖 Payment 接口隔离支付提供商差异；生产环境需要接入真实
支付网关时，应确认所选 adapter 已完成签名、请求、回调验签、状态映射和
退款证书等能力。
*/
package payment
