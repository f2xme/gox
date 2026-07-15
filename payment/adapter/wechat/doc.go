// Package wechat 为 payment 包提供微信支付 V3 实现。
//
// # 功能特性
//
//   - Native 支付（扫码 code_url）
//   - JSAPI 预支付与前端调起参数
//   - 公众号 OAuth（snsapi_base）换取 openid
//   - 订单查询、退款、关单
//   - 支付与退款异步通知验签、解密与解析
//   - 微信支付公钥 / 平台证书两种身份验签方式
//
// # 快速开始
//
// 验签方式须与商户平台「验证微信支付身份」当前启用项一致（二选一）。
//
// 微信支付公钥模式：
//
//	client, err := wechat.New(wechat.Config{
//		AppID:                "wx_app_id",
//		OAuthAppSecret:       "app_secret", // JSAPI OAuth 需要
//		MchID:                "1600000000",
//		MerchantSerialNo:     "merchant_api_cert_serial",
//		MerchantPrivateKey:   merchantPrivateKeyPEM,
//		APIV3Key:             "32-byte-api-v3-key................",
//		VerifyMode:           wechat.VerifyModePublicKey, // 可选，有公钥材料时可推断
//		WechatPayPublicKey:   wechatPayPublicKeyPEM,
//		WechatPayPublicKeyID: "PUB_KEY_ID_xxxxxxxx", // 保留 PUB_KEY_ID_ 前缀
//	})
//
// 平台证书自动模式（须显式 VerifyMode；服务可访问微信 API）：
//
//	// 启动拉取全量有效平台证书；默认定时刷新（后台 goroutine，与 client 同生命周期、无法停止）。
//	// 短生命周期/测试请设 PlatformCertAutoRefresh=false。
//	// 回调未知 serial 会受控补拉：并发合流、冷却约 60s、miss 负向缓存约 5min。
//	autoRefresh := false // 测试示例
//	client, err := wechat.New(wechat.Config{
//		AppID:                   "wx_app_id",
//		MchID:                   "1600000000",
//		MerchantSerialNo:        "merchant_api_cert_serial",
//		MerchantPrivateKey:      merchantPrivateKeyPEM,
//		APIV3Key:                "32-byte-api-v3-key................",
//		VerifyMode:              wechat.VerifyModePlatformCertAuto,
//		PlatformCertAutoRefresh: &autoRefresh,
//	})
//
// 平台证书静态模式（离线或可手动更换的固定证书）：
//
//	// 仅登记单张序列号；CERTIFICATE PEM 会与配置序列号交叉校验。
//	// 轮换窗口请改用自动模式。
//	client, err := wechat.New(wechat.Config{
//		AppID:                "wx_app_id",
//		MchID:                "1600000000",
//		MerchantSerialNo:     "merchant_api_cert_serial",
//		MerchantPrivateKey:   merchantPrivateKeyPEM,
//		APIV3Key:             "32-byte-api-v3-key................",
//		VerifyMode:           wechat.VerifyModePlatformCertStatic,
//		PlatformCert:         platformCertPEM,
//		PlatformCertSerialNo: "platform_cert_serial",
//	})
//
// # 验签模式
//
// 商户平台「验证微信支付身份」两种方式只能使用一种：
//
//  1. 微信支付公钥：WechatPayPublicKey + WechatPayPublicKeyID
//  2. 平台证书：
//     - 自动：须 VerifyModePlatformCertAuto（拉取全量证书；回调 serial miss 会补拉一次）
//     - 静态：PlatformCert + PlatformCertSerialNo
//
// VerifyMode 为空时：公钥齐全 → public_key，静态证书齐全 → platform_cert_static；
// 材料皆空返回配置错误（fail-closed），不会静默自动拉证。
//
// 公钥与平台证书材料不可同时完整配置。只配置某一模式的一半字段会返回配置错误。
// PlatformCertAutoRefresh 仅平台证书自动模式合法（默认开启定时刷新）。
//
// # 错误分类
//
//   - payment.ErrInvalidConfig：缺字段、模式与材料不匹配、PEM/序列号校验失败、
//     自动拉证中可识别的配置/权限类失败等
//   - payment.ErrGateway：平台证书自动拉取的网络/网关类失败
//
// 自动拉证错误类型依赖 gopay 错误文本启发式分类；若对 ErrGateway 重试，
// 仍应排查 APIv3Key、商户平台证书权限与出网，而非仅当瞬时抖动。
//
// # 回调
//
// ParsePaymentNotification / ParseRefundNotification 会用已加载的公钥/平台证书验签，
// 再用 APIV3Key 解密 resource，并校验 appid / mchid。
// 自动模式下未知 Wechatpay-Serial 会补拉平台证书再验签（与 gopay 同步验签行为对齐）：
// 并发请求合流为一次全量拉取，成功刷新后全局冷却约 60s；
// 刷新后仍无目标 serial 则负向缓存约 5min，降低伪造 serial 的出站放大。
// 序列号已存在仍失败则不再补拉。补拉失败时错误同时可 errors.Is 到
// ErrInvalidSignature 与 ErrGateway/ErrInvalidConfig。
// 业务侧仍须在事务内核对订单、金额、状态并幂等入账；仅在业务成功后返回 SuccessResponse。
// 回调 URL 仍建议在网关层做 IP/频率限制。
//
// # 注意事项
//
//   - 金额在 payment.Order 中使用分（int64）
//   - 证书/密钥内容以 PEM 字符串传入，由调用方负责从文件或密钥管理服务读取
//   - 平台证书自动模式在 New 时会请求微信证书接口，需保证出网可达
//   - 默认定时刷新 goroutine 无法停止；进程内反复 New 可能叠加刷新任务，生产宜单例
//   - 公钥 ID / 证书序列号须与商户平台展示一致（公钥 ID 勿删 PUB_KEY_ID_ 前缀）
package wechat
