/*
Package payment 提供统一支付领域接口、校验、状态与回调模型。

具体实现位于 adapter 子包：

  - payment/adapter/alipay：支付宝当面付、WAP、查询、退款、关单、支付回调
  - payment/adapter/wechat：微信 Native、JSAPI、OAuth、查询、退款、关单、支付与退款回调
  - payment/adapter/onepay：同一中立二维码按扫码客户端路由微信或支付宝
  - payment/adapter/mock：并发安全、状态可控的内存支付测试实现

金额一律使用分为单位的 int64，不使用浮点数。所有网络操作接收
context.Context。调用方应给 context 设置 deadline，并对不确定的网关写请求先
查询、后决定是否重试。

# 基础支付

	result, err := provider.Pay(ctx, &payment.Order{
		OrderID:   "order-1001",
		Amount:    9900,
		Subject:   "会员订阅",
		NotifyURL: "https://merchant.example/payment/notify",
	})

PaymentResult.PayURL 是支付宝 qr_code 或微信 code_url；二维码图片由调用方渲染。

# 回调

adapter 先验签并解析通知。调用方仍须在数据库事务内核对订单、金额、状态并
幂等入账；只有业务事务成功后，才能写入 SuccessResponse。解析成功本身不代表
业务处理成功。

# 一码付

onepay 创建的是中立 HTTPS URL 与 PNG。微信扫码后执行 OAuth snsapi_base 与
JSAPI 支付；支付宝扫码后进入 WAP 收银台。业务必须实现 CheckoutResolver，
持久化并复用完整 WAP 或 JSAPI artifact。同一 OpenID 重复微信扫码复用未过期
JSAPI 参数；不同 provider 使用不同订单号。首个成功回调应原子完成主支付意图，
随后关闭另一平台仍待支付的订单。
*/
package payment
