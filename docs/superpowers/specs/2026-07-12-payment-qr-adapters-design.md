# 支付宝与微信扫码支付 Adapter 设计

日期：2026-07-12

## 目标

把 `payment/adapter/alipay` 与 `payment/adapter/wechat` 从占位实现升级为可用于生产扫码支付的 adapter。

- 支付宝使用当面付预创建接口 `alipay.trade.precreate`。
- 微信使用支付 API v3 Native 下单接口。
- 两个平台统一通过 `payment.Payment` 暴露下单、查询、退款、关单。
- 支付回调完成解析、验签；微信回调额外完成 AES-GCM 解密。
- 微信退款回调完成解析、验签、解密。
- 底层统一使用 `github.com/go-pay/gopay`，初始目标版本为 `v1.5.122`；实现时若该版本不能通过兼容性验证，只允许升级到更高补丁版本并在文档记录原因。

## 非目标

- 不生成二维码 PNG、SVG 或其他图片，只返回二维码内容字符串。
- 不实现 APP、H5、JSAPI、小程序、付款码支付。
- 不保存订单，不提供幂等数据库，不执行对账。
- 不自动重试退款或网关写操作。
- 不提供 HTTP webhook server；adapter 只解析调用方传入的请求。

## 破坏性变更

本次允许破坏性升级，不保留旧构造函数与无 `context.Context` 的接口。

`payment.Payment` 改为：

```go
type Payment interface {
	Pay(ctx context.Context, order *Order) (*PaymentResult, error)
	Query(ctx context.Context, orderID string) (*QueryResult, error)
	Refund(ctx context.Context, req *RefundRequest) (*RefundResult, error)
	Close(ctx context.Context, orderID string) error
}
```

所有网络请求把调用方 context 直接传给 gopay。adapter 的 10 秒默认 HTTP client timeout 与 context deadline 同时生效，以更早到期者为准。nil context 返回 `ErrInvalidRequest`。

## 核心领域模型

### Order

保留现有字段，增加：

```go
ExpireAt *time.Time
```

- `Amount` 单位为分，必须大于零。
- 币种固定为 `CNY`。
- `Subject`、`OrderID`、`NotifyURL` 必填。
- `NotifyURL` 必须是包含 host 的绝对 HTTP 或 HTTPS URL。
- `ExpireAt` 非空时必须晚于当前时间。
- `Extra` 保留，但本次扫码支付不依赖未文档化的 `Extra` 键。

`PaymentResult.PayURL` 统一承载支付宝 `qr_code` 或微信 `code_url`。`TransactionID` 在预创建阶段可能为空。

### RefundRequest

增加：

```go
OriginalAmount int64
```

- `Amount` 与 `OriginalAmount` 单位均为分。
- 二者必须大于零。
- `Amount` 不得大于 `OriginalAmount`。
- 微信退款使用 `OriginalAmount` 组装必需的原订单总金额。
- 支付宝退款忽略网关不需要的 `OriginalAmount`，但仍执行一致的金额校验。

### 状态

增加 `PaymentStatusRefunded`，避免把已全额退款订单错误映射为失败或关闭。

已知状态映射：

| 平台 | 平台状态 | gox 状态 |
| --- | --- | --- |
| 支付宝 | `WAIT_BUYER_PAY` | `pending` |
| 支付宝 | `TRADE_SUCCESS`、`TRADE_FINISHED` | `success` |
| 支付宝 | `TRADE_CLOSED` | `closed` |
| 微信 | `NOTPAY`、`USERPAYING`、`ACCEPT` | `pending` |
| 微信 | `SUCCESS` | `success` |
| 微信 | `CLOSED`、`REVOKED` | `closed` |
| 微信 | `PAYERROR` | `failed` |
| 微信 | `REFUND` | `refunded` |

未知状态返回状态映射错误，不猜测结果。

## 回调 API

核心包新增两个可选能力接口。业务可按需要依赖，不扩大 `Payment` 基础接口。

```go
type PaymentNotifier interface {
	ParsePaymentNotification(ctx context.Context, req *http.Request) (*PaymentNotification, error)
	SuccessResponse() NotifyResponse
}

type RefundNotifier interface {
	ParseRefundNotification(ctx context.Context, req *http.Request) (*RefundNotification, error)
	SuccessResponse() NotifyResponse
}
```

- 支付宝 adapter 实现 `PaymentNotifier`。
- 微信 adapter 实现 `PaymentNotifier` 与 `RefundNotifier`。
- 支付宝当前扫码支付退款为同步接口，本次不声明实现 `RefundNotifier`。

统一通知至少包含：平台、商户订单号、平台交易号、状态、金额、发生时间和复制后的 provider 扩展字段。退款通知另含商户退款单号与退款金额；adapter 不把内部可变 map 直接暴露给调用方。

`NotifyResponse` 包含 HTTP status、content type、body，并提供写入 `http.ResponseWriter` 的辅助方法。调用方必须先完成业务校验和幂等落库，再写成功 ACK：

- 支付宝：HTTP 200，`Content-Type: text/plain; charset=utf-8`，正文 `success`。
- 微信：HTTP 200，`Content-Type: application/json; charset=utf-8`，JSON `{"code":"SUCCESS","message":"成功"}`。

解析成功不代表业务已入账。adapter 不自动发送 ACK。

## 支付宝 Adapter

### 配置

```go
type Config struct {
	AppID           string
	SellerID        string
	PrivateKey      string
	AlipayPublicKey string
	Production      bool
}

func New(config Config, opts ...Option) (*Alipay, error)
```

私钥与公钥参数接收 PEM 内容，不接收文件路径。`New` 校验必填项并初始化 gopay 支付宝 v3 client。生产/沙箱由 `Production` 明确控制。

### 操作映射

- `Pay`：`TradePrecreate`，分转换为两位小数元，返回 `qr_code`。
- `Query`：`TradeQuery`，映射订单、交易号、金额、状态、支付时间。
- `Refund`：`TradeRefund`，返回退款状态与完成时间。
- `Close`：`TradeClose`。

响应同时检查 HTTP 状态与支付宝业务码。业务失败返回统一 provider error。

### 支付回调

1. 解析 form body。
2. 使用支付宝公钥执行 RSA2 验签。
3. 校验通知 `app_id == Config.AppID`。
4. 校验通知 `seller_id == Config.SellerID`。
5. 映射通知字段与状态。
6. 返回统一 `PaymentNotification`，不自动 ACK。

解析函数消费 `http.Request.Body` 一次；nil request、nil body、重复消费导致的空 body 均返回 `ErrInvalidRequest`。

## 微信 Adapter

### 配置

```go
type Config struct {
	AppID                  string
	MchID                  string
	MerchantSerialNo       string
	MerchantPrivateKey     string
	APIV3Key               string
	WechatPayPublicKey     string
	WechatPayPublicKeyID   string
}

func New(config Config, opts ...Option) (*WechatPay, error)
```

密钥参数接收内容，不接收文件路径。微信 v3 不支持沙箱。构造时使用 gopay `NewClientV3`，并以微信支付公钥与公钥 ID 开启自动验签。

### 操作映射

- `Pay`：`V3TransactionNative`，返回 `code_url`。
- `Query`：按商户订单号调用 v3 查询接口。
- `Refund`：`V3Refund`，传入退款金额、原订单金额与可选退款通知 URL。
- `Close`：`V3TransactionCloseOrder`。

同步响应必须通过签名验证；验签失败不得返回业务结果。

### 支付与退款回调

1. 读取微信签名 headers 与 JSON body。
2. 使用已配置微信支付公钥验签。
3. 使用 API v3 Key 解密 `resource`。
4. 支付通知校验 `appid` 与 `mchid`。
5. 退款通知校验 `mchid`。
6. 映射支付或退款通知。
7. 返回统一通知，不自动 ACK。

解析函数消费 `http.Request.Body` 一次；nil request、nil body、重复消费导致的空 body 均返回 `ErrInvalidRequest`。

## Options

两个 adapter 提供同名、各自定义的 options：

```go
WithTimeout(timeout time.Duration)
WithHTTPTransport(transport http.RoundTripper)
WithLogger(logger *slog.Logger)
```

- 默认超时为 10 秒。
- 默认 transport 使用 Go 安全 TLS 校验；禁止默认设置 `InsecureSkipVerify`。
- logger 默认静默，不记录请求 body、私钥、API key、签名或解密材料。
- 公共 API 不暴露 gopay、xhttp、xlog 类型。

## 错误模型

核心包新增可用于 `errors.Is` 的分类错误：

```go
ErrInvalidConfig
ErrInvalidRequest
ErrGateway
ErrInvalidSignature
ErrUnknownStatus
```

adapter 使用 `ProviderError` 保留：

- `Provider`
- `Operation`
- provider `Code`
- 安全的 `Message`
- 可 `Unwrap` 的 cause

错误字符串不得包含私钥、API v3 key、完整签名、完整回调密文或完整请求 body。

context 取消与 deadline 错误保留原始 cause，使调用方仍可使用 `errors.Is(err, context.Canceled)` 和 `errors.Is(err, context.DeadlineExceeded)`。

## 幂等与安全边界

- 商户订单号与退款单号承担 provider 侧幂等键职责。
- adapter 不因网络错误自动重放下单、退款或关单。
- 调用方收到不确定网络结果后，应先 `Query`，再决定后续动作。
- 回调可能重复、乱序；调用方必须在数据库事务内按订单号或平台交易号幂等处理。
- 回调验签通过后，调用方仍必须核对订单存在、金额一致、当前状态允许迁移。
- 私钥与 API key 由调用方从 secret manager 或安全配置加载；示例不包含真实凭据。

## 内部结构

每个 adapter 内部定义最小 gateway 接口，只包含本次需要的 gopay 调用。生产实现包装 gopay client；测试使用 fake gateway。gateway 接口不导出，避免第三方 SDK 成为 gox 公共 API。

实现文件布局：

```text
payment/
  payment.go
  notification.go
  error.go
  validation.go
  doc.go
  adapter/alipay/
    alipay.go
    config.go
    gateway.go
    notification.go
    mapping.go
    example_test.go
  adapter/wechat/
    wechat.go
    config.go
    gateway.go
    notification.go
    mapping.go
    example_test.go
```

## 测试策略

默认测试不访问真实支付网关，不产生真实交易。

### 核心包

- 表驱动测试订单、订单号、退款请求、URL、金额、过期时间校验。
- 测试 `NotifyResponse` 输出。
- 测试 provider error 的分类与 unwrap。

### 支付宝

- 编译期断言实现 `payment.Payment`、`payment.PaymentNotifier`。
- fake gateway 覆盖下单、查询、退款、关单与 provider 错误。
- 使用测试 RSA key 动态生成合法 RSA2 通知签名。
- 覆盖签名错误、AppID 错误、SellerID 错误、未知状态。
- 验证分到元的精确转换，不使用浮点数。

### 微信

- 编译期断言实现 `payment.Payment`、`payment.PaymentNotifier`、`payment.RefundNotifier`。
- fake gateway 覆盖下单、查询、退款、关单与 provider 错误。
- 使用测试 RSA key 和 AES-GCM fixture 覆盖支付、退款通知。
- 覆盖签名错误、解密错误、AppID/MchID 错误、未知状态。
- 覆盖 context 取消与超时传播。

### 验证命令

```bash
go test ./payment/...
go test ./...
go build ./...
```

真实一分钱测试只作为人工验收说明，依赖环境变量与商户账号，不进入默认 CI。

## 文档与迁移

- 更新 `payment/doc.go`，删除占位实现说明。
- 增加支付宝、微信扫码下单与 HTTP 回调示例。
- 更新 `AI_USAGE.md` 与 `llms.txt` 的 package map。
- 明确旧 `NewAlipay`、`NewWechatPay` 和无 context 方法已删除。
- 文档强调：二维码字符串需要调用方渲染；ACK 只能在业务事务成功后返回。

## 完成标准

- 两个平台扫码下单返回可渲染的二维码内容。
- 四个支付操作均调用真实 gopay gateway，不再返回 `ErrNotImplemented`。
- 支付回调均验签；微信支付与退款回调均完成验签和解密。
- 所有已知状态按表映射，未知状态显式报错。
- 默认测试不联网且稳定通过。
- payment 包文档、示例、AI 使用指南与新 API 一致。
