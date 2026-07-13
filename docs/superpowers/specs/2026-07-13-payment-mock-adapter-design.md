# Payment Mock Adapter 设计

## 背景

业务测试目前需要自行实现 `payment.Payment` fake。仓库已有支付宝、微信真实
adapter，但缺少可复用、并发安全、可控制状态的支付测试实现。

## 方案比较

1. 状态化内存 mock adapter（采用）：提供真实接口实现、订单状态与调用记录，
   适合 service、HTTP handler 和集成测试。
2. 仅提供函数字段 fake：实现短，但每个调用方仍需重复维护状态和参数断言。
3. 模拟支付宝/微信协议服务：最接近真实网关，但维护成本高，且和现有 adapter
   单元测试重复。

## 包与模块

- 新包：`github.com/f2xme/gox/payment/adapter/mock`
- 包名：`mock`
- 属于根模块，不创建独立 `go.mod` 或标签。
- 新增 `payment.ProviderMock`，默认 provider 为 `mock`；可通过 option 改为微信或
  支付宝，用于业务分支测试。
- 不引入第三方依赖。

## API

`Client` 实现：

- `payment.Payment`
- `payment.PaymentNotifier`
- `payment.RefundNotifier`

构造采用 Options 模式：

`New(opts ...Option) (*Client, error)` 校验 provider、默认状态、延迟和操作类型；
配置无效返回 `payment.ErrInvalidConfig`。

- `WithProvider`：设置 provider。
- `WithClock`：注入确定性时间。
- `WithDelay`：设置尊重 context 取消的固定延迟。
- `WithPaymentStatus`：设置新支付的默认状态，默认 `pending`。
- `WithRefundStatus`：设置新退款的默认状态，默认 `pending`。
- `WithOperationError`：为指定操作注入固定错误。

运行期控制：

- `SetPaymentStatus(orderID, status)`：修改支付状态；成功时写入支付时间。
- `SetRefundStatus(refundID, status)`：修改退款状态；成功时写入退款时间。
- `SetOperationError(operation, err)`：设置或清除操作错误。
- `Payments()`、`Refunds()`、`Calls()`：返回深拷贝快照。
- `Reset()`：清空订单、退款和调用记录，保留构造配置与操作错误。

操作常量覆盖 `pay`、`query`、`refund`、`close`、支付回调解析和退款回调解析。

## 支付语义

- `Pay` 复用核心校验，创建待支付记录并返回确定性
  `mock://pay/<escaped-order-id>`。
- 支付与退款流水号分别使用 `mock-pay-<order-id>` 和
  `mock-refund-<refund-id>`，便于稳定断言。
- 重复订单号返回 `payment.ErrInvalidRequest`，避免隐藏业务幂等问题。
- `Query` 返回当前快照；不存在的订单返回 `payment.ErrInvalidRequest`。
- `Close` 只关闭 pending 订单；重复关闭幂等成功；成功或已退款订单不可关闭。
- `Refund` 要求原订单已成功、原金额匹配，且累计退款不超过订单金额。
- `SetRefundStatus` 把退款切换为 `pending` 或 `success` 前，排除当前记录
  重算有效退款总额；超过订单金额返回 `payment.ErrInvalidRequest`，不修改
  状态。`failed` 和 `closed` 退款不占用可退金额。
- 所有输入和输出中的 map、slice、array、pointer、interface 和 struct
  均递归复制，调用方不能修改内部状态。具体 Go 类型应保留，不使用
  JSON round-trip 或 `unsafe`。
- `Extra` 拒绝循环引用、func、chan、unsafe pointer，以及含无法安全
  复制的非导出引用字段的 struct；`time.Time` 按值复制。无效 `Extra`
  返回 `payment.ErrInvalidRequest`。
- 所有公开方法并发安全。

## 回调测试

Client 提供：

- `PaymentNotificationRequest(orderID)`
- `RefundNotificationRequest(refundID)`

两者根据当前内存状态生成 mock JSON HTTP 请求。对应 `Parse*Notification` 只解析
mock 自有格式，不模拟支付宝或微信签名协议；请求体限制为 64 KiB，拒绝尾随
JSON。`SuccessResponse` 返回 HTTP 200 JSON 成功回执。

## 错误与延迟

- 参数错误使用 `payment.ErrInvalidRequest`。
- `Extra` 含不支持的类型或循环引用时使用 `payment.ErrInvalidRequest`。
- 注入错误原样返回，便于 `errors.Is` 断言。
- 延迟等待监听 `ctx.Done()`；取消时返回包装后的 context 错误，不修改状态。
- 调用记录包含操作名、订单号、退款号和调用时间；有效请求即记录，包括注入错误。

## 文件结构

- `doc.go`：包文档、功能特性和快速开始。
- `option.go`：Options、Operation 与函数式选项。
- `mock.go`：Client、支付状态与记录。
- `notification.go`：mock 回调请求与解析。
- `clone.go`：防止测试数据别名泄漏的深拷贝。
- `mock_test.go`：状态、错误、延迟、并发、退款状态回切与深拷贝测试。
- `notification_test.go`：回调生成、解析、大小限制和尾随 JSON 测试。

## 验证

1. `go test ./payment/adapter/mock`。
2. `go test -race ./payment/adapter/mock`。
3. 根模块 `go test ./...`。
4. `git diff --check`。

## 非目标

- 不复刻支付宝或微信请求、签名和错误码。
- 不提供生产环境开关或持久化。
- 不模拟一码付的 `CheckoutResolver`。
