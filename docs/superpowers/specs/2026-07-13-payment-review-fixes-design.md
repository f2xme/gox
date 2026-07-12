# 支付代码 Review 修复设计

日期：2026-07-13

## 目标

修复 2026-07-13 payment review 发现的四项问题：

- 微信网页授权 `state` 超出协议长度与字符集限制。
- 支付宝公钥无效时构造仍成功，可能静默关闭同步响应验签。
- 相同支付意图重复微信扫码时再次下单，触发 `OUT_TRADE_NO_USED`。
- 小尺寸 PNG 无法容纳 QR 矩阵时仍返回不可扫描图片。

允许破坏 `onepay` 公共 API。不增加进程内缓存，不把分布式幂等伪装成单机锁。

## Checkout 持久化边界

`onepay` 不再直接创建支付宝或微信支付订单。业务 Resolver 负责调用 adapter、持久化 provider 订单与 checkout artifact，并原子复用有效结果。

```go
type Checkout struct {
	Provider  payment.Provider
	OrderID   string
	WAP       *payment.WAPResult
	JSAPI     *payment.JSAPIResult
	ExpiresAt time.Time
}

type CheckoutResolver interface {
	ResolveOrCreate(
		ctx context.Context,
		intentID string,
		provider payment.Provider,
		payerOpenID string,
	) (*Checkout, error)
}

type WechatOAuth interface {
	OAuthURL(redirectURL, state string) (string, error)
	ExchangeOAuthCode(ctx context.Context, code string) (string, error)
}

type Config struct {
	BaseURL  string
	Path     string
	TokenKey []byte
	TokenTTL time.Duration
	Resolver CheckoutResolver
	Wechat   WechatOAuth
}
```

支付宝调用 Resolver 时 `payerOpenID` 为空，返回非空 `WAP`。微信在 OAuth 换得 OpenID 后调用 Resolver，返回非空 `JSAPI`。onepay 校验 provider、artifact 类型和过期时间，不接收不匹配结果。

Resolver 必须满足：

- 同一 `intentID + provider` 只允许一个活动 provider 订单。
- 微信首次创建时持久化 OpenID 的不可逆摘要、`JSAPIResult` 与 artifact 过期时间。
- 相同 OpenID 重复扫码时复用未过期 `JSAPIResult`，不得再次调用微信下单接口。
- 不同 OpenID 扫描已有活动微信订单时返回业务错误，不复用绑定其他付款人的订单。
- artifact 过期后先确认原订单状态；仍未支付时关闭原订单，使用新商户订单号创建新订单。
- provider 网络结果不确定时先查询，不盲目重试创建。
- 首个成功支付回调仍由业务事务 compare-and-swap 完成主支付意图，随后关闭另一 provider 待支付订单。

## 微信 OAuth State

state 使用固定二进制布局，不再编码 JSON：

| 字段 | 字节数 |
| --- | ---: |
| 版本 | 1 |
| token SHA-256 前缀 | 12 |
| 随机 cookie nonce | 12 |
| 过期 Unix 秒，大端序 | 8 |
| HMAC-SHA256 截断 tag | 16 |

总计 49 字节，使用无 padding Base32 编码后为 79 个字符；字符仅为 `A-Z2-7`，小于 128 字节。

HMAC key 继续从 `TokenKey` 加域分隔字符串派生。验证时先 Base32 解码，再校验固定长度、版本、HMAC、token hash、cookie nonce 和过期时间。hash、nonce 与 tag 比较使用常量时间比较。cookie 保持 `Secure`、`HttpOnly`、`SameSite=Lax` 和 token 专属 path。

## 支付宝公钥

`newGopayGateway` 在调用 `AutoVerifySign` 前，使用与 gopay 一致的 PEM 公钥解析器主动解析 `AlipayPublicKey`。解析失败立即返回错误，`New` 包装为 `ErrInvalidConfig`。只有解析成功才开启自动验签。

回调继续使用同一配置公钥验签。测试增加“有效私钥 + 无效支付宝公钥”用例，构造必须失败。

## QR 渲染

QR 渲染保持精确输出尺寸和四模块白边：

1. `totalModules = matrix.Width() + 8`。
2. `moduleSize = size / totalModules`。
3. `moduleSize < 1` 时返回错误，不生成损坏 PNG。
4. 使用统一整数 `moduleSize` 绘制全部模块。
5. 将实际 QR 居中放入 `size × size` 白色画布，剩余像素成为额外白边。

不使用按列取整的非均匀模块宽度。正常码测试保留 PNG 尺寸校验；增加 `WithQRSize(128)` 与长 intent 的失败测试。

## Handler 流程

支付宝：解密 token → `Resolver.ResolveOrCreate(..., ProviderAlipay, "")` → 校验 `Checkout.WAP` 与官方 gateway allowlist → HTTP 303。

微信：解密 token → OAuth state/cookie → code 换 OpenID → `Resolver.ResolveOrCreate(..., ProviderWechat, openID)` → 校验 `Checkout.JSAPI` 与过期时间 → 输出 CSP nonce bridge HTML。

Resolver 或 artifact 校验失败统一返回 HTTP 502 安全错误页，不暴露内部错误。token、过期、UA、OAuth state 的既有 HTTP 状态保持不变。

## 测试与完成标准

- state 长度不超过 128，匹配 `^[A-Z2-7]+$`，支持 round trip，拒绝篡改、错误 cookie、错误 token 与过期值。
- 支付宝无效公钥在 `New` 阶段失败，同步自动验签不会被静默关闭。
- 同一 intent、provider、OpenID 两次微信扫码只产生一次 provider checkout；第二次返回缓存 JSAPI 参数。
- 不同 OpenID 由 Resolver fake 返回冲突，onepay 输出安全 502。
- 128 像素无法容纳长 QR 时 `CreateCode` 返回错误；普通二维码仍为可解析 PNG。
- `go test -race ./payment/...`、`go vet ./payment/...`、`go build ./payment/...` 通过。
- 使用不含未完成 OAuth2 workspace 条目的临时 workspace 执行 `go test ./...` 通过。

## 迁移

原 `OrderResolver`、`AlipayCheckout`、`WechatCheckout` 删除。调用方把支付宝/微信 adapter 注入自己的 `CheckoutResolver`，在业务持久层完成创建与 artifact 缓存。`onepay.Config` 不再包含 `Alipay`；`Wechat` 仅提供 OAuth 能力。
