// Package alipay 为 payment 包提供支付宝当面付、WAP、查询、退款、关单与回调验签实现。
//
// # 功能特性
//
//   - 当面付预创建（TradePrecreate），返回 qr_code
//   - 手机网站支付（TradeWapPay），返回收银台 URL
//   - 订单查询、退款、关单
//   - 异步支付通知解析与验签
//   - 密钥模式与公钥证书模式两种加签/验签方式
//   - 可选 AESKey 透传（实验性；受 gopay 限制，非完整接口内容加密）
//   - 正式环境与新版沙箱环境网关切换
//
// # 快速开始
//
// 密钥模式 + 沙箱：
//
//	client, err := alipay.New(alipay.Config{
//		AppID:           "sandbox_app_id",
//		SellerID:        "sandbox_pid",
//		PrivateKey:      appPrivateKeyPEM,
//		AlipayPublicKey: alipayPublicKeyPEM,
//		// AESKey: 仅实验性透传 gopay SetAESKey，见下方「接口内容加密」
//		Environment:     alipay.EnvSandbox,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	result, err := client.Pay(ctx, &payment.Order{
//		OrderID:   "order-1001",
//		Amount:    100, // 单位：分
//		Subject:   "测试商品",
//		NotifyURL: "https://merchant.example/payment/alipay/notify",
//	})
//
// 证书模式 + 正式环境：
//
//	client, err := alipay.New(alipay.Config{
//		AppID:            "app_id",
//		SellerID:         "pid",
//		PrivateKey:       appPrivateKeyPEM,
//		AppPublicCert:    appCertPublicKeyContent,    // appCertPublicKey_*.crt
//		AlipayRootCert:   alipayRootCertContent,      // alipayRootCert.crt
//		AlipayPublicCert: alipayPublicCertContent,    // alipayCertPublicKey_RSA2.crt
//		Environment:      alipay.EnvProduction,
//	})
//
// # 加签模式
//
// 二选一：
//
//  1. 密钥模式：PrivateKey + AlipayPublicKey
//  2. 证书模式：PrivateKey + AppPublicCert + AlipayRootCert + AlipayPublicCert
//
// 证书三件套齐全时优先使用证书模式。只配置部分证书字段会返回配置错误，不会静默回退密钥模式。
//
// # 环境
//
// 推荐使用 Environment：
//
//   - alipay.EnvSandbox：https://openapi-sandbox.dl.alipaydev.com/gateway.do
//   - alipay.EnvProduction：https://openapi.alipay.com/gateway.do
//
// Environment 为空时回退 Production 字段（true 正式，false 沙箱）。
// 零值配置默认沙箱，避免误连正式网关。沙箱须使用开放平台新版沙箱应用凭证。
//
// 兼容字段 Production 已标记 Deprecated，新代码请使用 Environment。
//
// # 回调
//
// ParsePaymentNotification 会验签并校验 app_id / seller_id。
// 业务侧仍须在事务内核对订单、金额、状态并幂等入账；仅在业务成功后返回 SuccessResponse。
//
// # 接口内容加密（实验性）
//
// Config.AESKey 会透传给 gopay Client.SetAESKey。请注意：
//
//   - gopay v1.5.x 将密钥字符串**原样**作为 AES 字节（不 Base64 解码），长度须为 16/24/32
//   - 上游注释写明 SetAESKey「目前不可用，设置后会报错」（如缺少 encrypt_type、响应不自动解密）
//   - 本适配器不做完整开放平台「接口内容加密」兼容实现；未开启内容加密时请留空 AESKey
//   - 若必须使用，请在目标 gopay 版本上自行联调，勿默认视为生产可用
//
// # 注意事项
//
//   - 金额在 payment.Order 中使用分（int64），适配器内部转换为元字符串
//   - 证书/密钥内容以 PEM 字符串传入，由调用方负责从文件或密钥管理服务读取
//   - AESKey 为可选实验性字段；空表示不调用 SetAESKey
//   - GatewayBaseURL 是与 go-pay 约定一致的配置层镜像，非底层 client 实时 URL
package alipay
