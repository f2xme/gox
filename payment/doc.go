/*
Package payment 提供统一的支付操作抽象层。

# 概述

payment 包定义了支付操作的标准接口，支持多种支付提供商（微信支付、支付宝等）。
通过统一的 API，你可以轻松地在不同的支付服务之间切换，而无需修改业务代码。

# 核心功能

  - 统一的支付接口
  - 支持支付、查询、退款、关闭订单
  - 金额使用分（int64）避免浮点精度问题
  - 灵活的 Extra 字段支持提供商特定参数

# 核心接口

## Payment - 支付操作接口

所有支付实现都必须实现此接口：

	type Payment interface {
		Pay(order *Order) (*PaymentResult, error)
		Query(orderID string) (*QueryResult, error)
		Refund(req *RefundRequest) (*RefundResult, error)
		Close(orderID string) error
	}

# 使用示例

## 创建支付订单

	order := &payment.Order{
		OrderID:     "ORDER20240101001",
		Amount:      9900, // 99.00 元（单位：分）
		Subject:     "商品名称",
		Description: "商品描述",
		NotifyURL:   "https://example.com/notify",
		Extra: map[string]interface{}{
			"product_id": "12345",
		},
	}

	result, err := pay.Pay(order)
	if err != nil {
		log.Fatal(err)
	}

	// 返回支付参数给前端
	fmt.Println(result.TransactionID) // 支付平台的交易 ID
	fmt.Println(result.PayURL)        // 支付 URL（扫码支付）
	fmt.Println(result.Extra)         // 其他支付参数

## 查询支付状态

	result, err := pay.Query("ORDER20240101001")
	if err != nil {
		log.Fatal(err)
	}

	switch result.Status {
	case payment.StatusSuccess:
		// 支付成功
	case payment.StatusPending:
		// 支付中
	case payment.StatusFailed:
		// 支付失败
	case payment.StatusClosed:
		// 订单已关闭
	}

## 申请退款

	req := &payment.RefundRequest{
		OrderID:       "ORDER20240101001",
		RefundID:      "REFUND20240101001",
		RefundAmount:  5000, // 退款 50.00 元
		RefundReason:  "用户申请退款",
		NotifyURL:     "https://example.com/refund-notify",
	}

	result, err := pay.Refund(req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.RefundID)      // 退款单号
	fmt.Println(result.Status)        // 退款状态

## 关闭订单

	err := pay.Close("ORDER20240101001")
	if err != nil {
		log.Fatal(err)
	}

# 可用适配器

## 微信支付

	import "github.com/f2xme/gox/payment/adapter/wechatadapter"

	pay := wechatadapter.New(
		wechatadapter.WithAppID("your-app-id"),
		wechatadapter.WithMchID("your-mch-id"),
		wechatadapter.WithAPIKey("your-api-key"),
	)

## 支付宝

	import "github.com/f2xme/gox/payment/adapter/alipayadapter"

	pay := alipayadapter.New(
		alipayadapter.WithAppID("your-app-id"),
		alipayadapter.WithPrivateKey("your-private-key"),
		alipayadapter.WithPublicKey("alipay-public-key"),
	)

# 支付状态

	StatusPending  // 支付中
	StatusSuccess  // 支付成功
	StatusFailed   // 支付失败
	StatusClosed   // 订单已关闭

# 最佳实践

## 1. 金额使用分（int64）

	// 推荐：使用分
	amount := int64(9900) // 99.00 元

	// 不推荐：使用浮点数
	amount := 99.00 // 可能有精度问题

## 2. 生成唯一的订单号

	// 使用时间戳 + 随机数
	orderID := fmt.Sprintf("ORDER%d%06d", time.Now().Unix(), rand.Intn(1000000))

	// 或使用 UUID
	orderID := "ORDER" + uuid.New().String()

## 3. 处理异步通知

	func handleNotify(c *gin.Context) {
		// 验证签名
		if !verifySignature(c.Request) {
			c.Status(400)
			return
		}

		// 解析通知数据
		var notify NotifyData
		c.BindJSON(&notify)

		// 更新订单状态
		updateOrderStatus(notify.OrderID, notify.Status)

		// 返回成功响应
		c.JSON(200, map[string]string{"code": "SUCCESS"})
	}

## 4. 实现幂等性

	// 使用订单号作为幂等键
	if isOrderPaid(orderID) {
		return errors.New("订单已支付")
	}

	result, err := pay.Pay(order)

## 5. 处理退款

	// 检查退款金额
	if refundAmount > order.Amount {
		return errors.New("退款金额不能大于订单金额")
	}

	// 生成唯一的退款单号
	refundID := generateRefundID()

	req := &payment.RefundRequest{
		OrderID:      orderID,
		RefundID:     refundID,
		RefundAmount: refundAmount,
	}

	result, err := pay.Refund(req)

## 6. 超时处理

	// 设置支付超时时间
	order.Extra = map[string]interface{}{
		"timeout_express": "30m", // 30 分钟后自动关闭
	}

	// 定时检查未支付订单
	go func() {
		time.Sleep(30 * time.Minute)
		if !isOrderPaid(orderID) {
			pay.Close(orderID)
		}
	}()

# 安全考虑

  - 验证异步通知的签名
  - 使用 HTTPS 传输敏感信息
  - 不要在客户端暴露 API 密钥
  - 实现订单金额校验
  - 记录所有支付操作日志

# 线程安全

所有支付实现都应该是线程安全的，可以在多个 goroutine 中并发使用。
*/
package payment
