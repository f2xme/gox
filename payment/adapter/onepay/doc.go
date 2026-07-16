// Package onepay 实现支付宝与微信共用的中立支付二维码（一码付）。
//
// 同一 HTTPS 地址与 PNG 二维码，按扫码客户端 User-Agent 路由：
// 微信走 OAuth + JSAPI，支付宝走 WAP 收银台。onepay 不直接向平台下单，
// 收银台创建与幂等复用由业务实现的 CheckoutResolver 负责。
//
// # 功能特性
//
//   - 生成中立 HTTPS 支付 URL 与可扫码 PNG
//   - AES-256-GCM 加密 token（金额、订单号不进二维码）
//   - 微信：OAuth snsapi_base → OpenID → JSAPI bridge HTML（CSP nonce）
//   - 微信调起页文案/模板可配（WechatPage）；零值用 DefaultWechat* 与默认 HTML
//   - 支付宝：WAP 收银台 HTTP 303（仅官方 gateway host）
//   - 未知客户端返回安全提示页，不猜测渠道
//   - 标准库 net/http.Handler，可挂载到任意 HTTP 框架
//
// # 快速开始
//
//	svc, err := onepay.New(onepay.Config{
//		BaseURL:  "https://pay.example.com",
//		TokenKey: tokenKey32Bytes, // 恰好 32 字节
//		Resolver: businessResolver, // 业务实现 CheckoutResolver
//		Wechat:   wechatClient,     // 提供 OAuth 的微信 adapter
//		// 可选：覆盖默认「支付中…」等文案
//		// WechatPage: onepay.WechatPage{LoadingText: "正在支付…", Title: "收银台"},
//	}, onepay.WithQRSize(256))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	code, err := svc.CreateCode(ctx, "intent-1001")
//	// code.URL 写入二维码；code.PNG 可直接下发图片
//
//	http.Handle("/pay/", svc.Handler()) // 或 mux.Handle("/pay/", svc.Handler())
//
// # 微信调起页（WechatPage）
//
// 默认 loading 文案为「支付中…」（旧版「正在调起微信支付…」已变更；E2E/文案断言需同步）。
//
// 自定义整页 Template 时数据为 WechatBridgeData，且必须兼容 WechatBridgeCSP：
//
//	tpl := template.Must(template.New("wx").Parse(`<!doctype html>
//	<html lang="zh-CN"><head><meta charset="utf-8"><title>{{.Title}}</title></head>
//	<body><p id="status">{{.LoadingText}}</p>
//	<script nonce="{{.Nonce}}">
//	const pay = {{.Params}};
//	function invokePay(){
//	  WeixinJSBridge.invoke('getBrandWCPayRequest', pay, function(res){
//	    document.getElementById('status').textContent =
//	      res.err_msg === 'get_brand_wcpay_request:ok' ? {{.SuccessText}} : {{.FailText}};
//	  });
//	}
//	if (typeof WeixinJSBridge === 'undefined') {
//	  document.addEventListener('WeixinJSBridgeReady', invokePay, false);
//	} else { invokePay(); }
//	</script></body></html>`))
//
//	WechatPage: onepay.WechatPage{Template: tpl, LoadingText: "支付中…"}
//
// # CheckoutResolver 契约
//
// Resolver 是业务持久化边界，onepay 不缓存订单。必须满足：
//
//   - 同一 intentID + provider 只允许一个活动平台订单
//   - 支付宝：payerOpenID 为空，返回未过期 WAP
//   - 微信：绑定 OpenID 摘要；同 OpenID 复用未过期 JSAPI；不同 OpenID 不得复用
//   - 不同 provider 使用不同商户订单号，并保留映射到主支付意图
//   - artifact 过期后先查原单状态；仍未支付则关单并换新单号重建
//   - 平台写请求结果不确定时先 Query，不盲目重试创建
//   - 首个成功支付回调由业务 CAS 完成主意图，再关闭另一平台待支付单
//
// 内存骨架见本包 ExampleCheckoutResolver / TestMemoryResolver*：
// 创建中/退役中占位、intent paid 单调、uncertain 仅锁 provider 槽、ctx 取消不落永久墓碑。
// 主意图 CAS 与真实 DB 须由业务补齐，勿直接当生产实现。
//
// # 挂载
//
// 标准库：
//
//	mux := http.NewServeMux()
//	mux.Handle("/pay/", svc.Handler())
//
// Gin（或其它框架）把标准 Handler 包一层即可：
//
//	r.Any("/pay/*path", gin.WrapH(svc.Handler()))
//
// # 注意事项
//
//   - BaseURL 须为 HTTPS origin（localhost/loopback 测试允许 HTTP）
//   - Path 默认 /pay/，须以 / 开头和结尾，禁止 . 与 .. 段
//   - TokenTTL 默认 15 分钟，最大 24 小时
//   - 同步跳转/JS bridge 仅用于 UI；入账只认异步回调验签结果
//   - 请使用微信或支付宝扫码；其它 UA 返回 HTTP 400
package onepay
