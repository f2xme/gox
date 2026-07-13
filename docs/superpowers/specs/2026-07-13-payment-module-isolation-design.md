# Payment Adapter 模块隔离设计

## 背景

`payment` 核心仅依赖标准库，但支付宝、微信与一码付 adapter 分别依赖
`go-pay` 和 `go-qrcode`。当前三个 adapter 属于根模块，导致根模块携带支付
SDK 与二维码库，和仓库现有的 adapter 依赖隔离方式不一致。

## 方案比较

1. 仅拆分三个 adapter（采用）：核心继续跟随根模块，第三方依赖按 adapter
   隔离；导入路径不变，影响最小。
2. 将整个 `payment` 拆成独立模块：边界完整，但核心也需要独立版本与标签，
   对仅使用领域类型的调用方增加版本管理成本。
3. 保持现状：无需改动，但所有根模块用户都会下载支付 SDK，依赖污染继续存在。

## 模块边界

- `github.com/f2xme/gox/payment`：保留在根模块，仅含支付领域模型、校验、错误与
  回调协议，不创建独立标签。
- `github.com/f2xme/gox/payment/adapter/alipay`：独立模块，直接依赖根模块、
  `github.com/go-pay/gopay` 与 `github.com/go-pay/crypto`。
- `github.com/f2xme/gox/payment/adapter/wechat`：独立模块，直接依赖根模块与
  `github.com/go-pay/gopay`。
- `github.com/f2xme/gox/payment/adapter/onepay`：独立模块，直接依赖根模块与
  `github.com/yeqown/go-qrcode/v2`。

每个 adapter 的 `go.mod` 使用本地 `replace` 指向仓库根模块，加入 `go.work`。
现有包路径、公开 API、运行时行为保持不变。

## 依赖整理

分别对根模块和三个新模块执行 `go mod tidy`。若根模块没有其他引用，删除
`go-pay`、`go-qrcode` 及其仅由支付 adapter 引入的间接依赖。`go.work` 只登记
三个 adapter，不登记 `payment` 核心。

## 验证

执行以下检查：

1. 根模块 `go test ./...`，确认 `payment` 核心仍由根模块发布。
2. 三个 adapter 分别执行 `go test ./...`。
3. `git diff --check`，确认模块文件与依赖锁文件干净。
4. 从根模块测试结果确认三个 adapter 不再被根模块递归测试，避免模块边界失效。

## 发布

- 三个新 adapter 首次发布：
  - `payment/adapter/alipay/v0.1.0`
  - `payment/adapter/wechat/v0.1.0`
  - `payment/adapter/onepay/v0.1.0`
- 根模块使用补丁版本发布模块边界修正；已发布标签保持不可变，不移动旧标签。

## 非目标

- 不修改支付 API、支付流程或回调语义。
- 不拆分 `payment` 核心模块。
- 不引入新的支付 SDK。
