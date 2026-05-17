/*
Package payment 提供统一的支付操作抽象层。

payment 包只定义支付领域的公共接口和数据结构。具体支付服务商适配器
位于 adapter 子包中；当前仓库内的支付宝和微信支付适配器仍是占位实现，
不会连接真实支付网关，调用支付操作会返回 ErrNotImplemented。

# 核心接口

	type Payment interface {
		Pay(order *Order) (*PaymentResult, error)
		Query(orderID string) (*QueryResult, error)
		Refund(req *RefundRequest) (*RefundResult, error)
		Close(orderID string) error
	}

# 使用建议

业务代码可以依赖 Payment 接口隔离支付提供商差异；生产环境需要接入真实
支付网关时，应确认所选 adapter 已完成签名、请求、回调验签、状态映射和
退款证书等能力。
*/
package payment
