# SMS 适配器模式重构设计

## 概述

将 `sms` 包重构为适配器模式，每个短信服务提供商（阿里云、腾讯云、火山引擎）作为独立的 Go module，实现按需引入和独立版本管理。

## 设计目标

1. **模块化**：每个适配器是独立的 Go module，用户只引入需要的适配器
2. **解耦依赖**：主 module 不再依赖各厂商的 SDK
3. **统一接口**：保持 `sms.SMS` 接口作为统一抽象层
4. **独立演进**：各适配器可以独立发版，互不影响

## 整体架构

### 模块划分

系统将包含以下组件：

1. **核心接口**（在主 module `github.com/f2xme/gox` 中）
   - 位置：`sms/sms.go`
   - 内容：`SMS` 接口定义
   - 依赖：无外部依赖

2. **阿里云适配器**（独立 module）
   - Module 路径：`github.com/f2xme/gox/sms/adapter/aliyun`
   - 依赖：主 module（获取接口和 config）+ 阿里云 SDK
   - 版本：独立版本号，从 `v0.1.0` 开始

3. **腾讯云适配器**（独立 module）
   - Module 路径：`github.com/f2xme/gox/sms/adapter/tencent`
   - 依赖：主 module + 腾讯云 SDK
   - 版本：独立版本号，从 `v0.1.0` 开始

4. **火山引擎适配器**（独立 module）
   - Module 路径：`github.com/f2xme/gox/sms/adapter/volcengine`
   - 依赖：主 module + 火山引擎 SDK
   - 版本：独立版本号，从 `v0.1.0` 开始

### 依赖关系

```
用户代码
  ↓
adapter/aliyun ──→ github.com/f2xme/gox (主 module)
  ↓                    ↓
阿里云 SDK          sms.SMS 接口 + config.Config
```

每个适配器通过依赖主 module 来获取 `sms.SMS` 接口定义和 `config.Config` 支持。

## 目录结构

### 新的目录布局

```
sms/
├── sms.go                           # SMS 接口定义（主 module）
├── doc.go                           # 包文档
└── adapter/
    ├── aliyun/
    │   ├── go.mod                   # module github.com/f2xme/gox/sms/adapter/aliyun
    │   ├── aliyun.go                # 适配器实现
    │   ├── option.go                # 选项模式
    │   ├── doc.go                   # 包文档
    │   ├── aliyun_test.go           # 单元测试
    │   └── example_test.go          # 示例代码
    ├── tencent/
    │   ├── go.mod                   # module github.com/f2xme/gox/sms/adapter/tencent
    │   ├── tencent.go
    │   ├── option.go
    │   ├── doc.go
    │   ├── tencent_test.go
    │   └── example_test.go
    └── volcengine/
        ├── go.mod                   # module github.com/f2xme/gox/sms/adapter/volcengine
        ├── volcengine.go
        ├── option.go
        ├── doc.go
        └── volcengine_test.go
```

### 迁移策略

1. **删除旧路径**：移除现有的 `sms/aliyun/`、`sms/tencent/`、`sms/volcengine/` 目录
2. **创建新路径**：在 `sms/adapter/` 下重建各适配器，每个带独立 `go.mod`
3. **更新主 module**：从 `gox/go.mod` 中移除 SMS 厂商 SDK 依赖
4. **独立版本**：每个 adapter 从 `v0.1.0` 开始独立版本号
5. **不提供向后兼容**：完全迁移，用户必须更新导入路径

## 代码实现

### 核心接口

`sms/sms.go` 保持现有定义：

```go
package sms

// SMS 定义短信服务提供商接口
type SMS interface {
    // Send 发送短信消息
    //
    // 参数：
    //   - phone: 手机号码
    //   - templateCode: 短信模板代码
    //   - templateParam: 模板参数（JSON 格式），例如 {"code":"1234"}
    Send(phone, templateCode, templateParam string) error
}
```

### 适配器实现模式

每个适配器遵循统一的实现模式：

**1. 构造函数**
- `New(opts ...Option) (sms.SMS, error)` - 使用选项模式创建
- `NewWithConfig(cfg config.Config) (sms.SMS, error)` - 从配置创建

**2. 选项模式**
```go
type Options struct {
    // 厂商特定字段
    AccessKeyID     string
    AccessKeySecret string
    // ...
}

type Option func(*Options)

func WithAccessKeyID(id string) Option {
    return func(o *Options) {
        o.AccessKeyID = id
    }
}
```

**3. 接口实现**
```go
type aliyunSMS struct {
    options Options
    client  *dysmsapi.Client
}

var _ sms.SMS = (*aliyunSMS)(nil)

func (s *aliyunSMS) Send(phone, templateCode, templateParam string) error {
    // 调用厂商 SDK
}
```

### 错误处理

统一的错误格式，带厂商前缀：
```go
return fmt.Errorf("aliyun sms: %w", err)
return fmt.Errorf("tencent sms: %w", err)
return fmt.Errorf("volcengine sms: %w", err)
```

### 导入路径变化

**旧的导入方式（将被废弃）：**
```go
import "github.com/f2xme/gox/sms/adapter/aliyun"

client, err := aliyun.New(...)
```

**新的导入方式：**
```go
import (
    "github.com/f2xme/gox/sms"
    "github.com/f2xme/gox/sms/adapter/aliyun"
)

client, err := aliyun.New(...)
var _ sms.SMS = client  // 实现接口
```

## 版本管理

### 版本号规则

1. **主 module** (`github.com/f2xme/gox`)
   - 保持现有版本号
   - `sms.SMS` 接口变更会影响主 module 版本（应尽量避免破坏性变更）

2. **各适配器 module**
   - 初始版本：`v0.1.0`
   - 独立的语义化版本
   - 适配器实现变更不影响其他 module

### 兼容性保证

- `sms.SMS` 接口一旦稳定，应避免破坏性变更
- 如需扩展功能，考虑新增可选接口（如 `SMSBatch` 用于批量发送）
- 各适配器可以独立演进，不受其他适配器影响

### 发布流程

当需要发布适配器更新时：

1. 修改适配器代码
2. 在适配器目录下打 tag：
   ```bash
   git tag sms/adapter/aliyun/v0.2.0
   ```
3. 推送 tag：
   ```bash
   git push origin sms/adapter/aliyun/v0.2.0
   ```
4. Go module proxy 会自动识别子目录的 module

### go.mod 依赖示例

**aliyun 适配器的 go.mod：**
```go
module github.com/f2xme/gox/sms/adapter/aliyun

go 1.25

require (
    github.com/f2xme/gox v0.x.x  // 获取 sms.SMS 接口和 config.Config
    github.com/alibabacloud-go/darabonba-openapi/v2 v2.x.x
    github.com/alibabacloud-go/dysmsapi-20170525/v4 v4.x.x
    github.com/alibabacloud-go/tea v1.x.x
)
```

**用户项目的 go.mod：**
```go
require (
    github.com/f2xme/gox v0.x.x  // 主 module（包含 sms.SMS 接口）
    github.com/f2xme/gox/sms/adapter/aliyun v0.1.0  // 只引入需要的适配器
)
```

用户不需要的适配器（如 tencent、volcengine）及其依赖不会被引入。

## 测试与文档

### 测试策略

**1. 接口契约测试**（可选）
- 在主 module 的 `sms/` 下可以添加接口契约测试
- 确保所有适配器都正确实现接口

**2. 适配器单元测试**
- 每个适配器独立维护测试
- Mock 厂商 SDK 的网络调用
- 测试错误处理、参数验证、边界情况

**3. 示例代码**
- 每个适配器提供 `example_test.go`
- 展示基本用法和配置方式
- 作为可执行的文档

### 文档更新

**1. 主包文档** (`sms/doc.go`)

```go
/*
Package sms 提供统一的短信服务抽象层。

# 概述

sms 包定义了短信服务的标准接口，支持多种短信服务提供商。
通过这些接口，你可以轻松地在不同的短信服务提供商之间切换，而无需修改业务代码。

# 核心接口

    type SMS interface {
        Send(phone, templateCode, templateParam string) error
    }

# 可用适配器（独立 module）

## 阿里云短信

    import "github.com/f2xme/gox/sms/adapter/aliyun"

## 腾讯云短信

    import "github.com/f2xme/gox/sms/adapter/tencent"

## 火山引擎短信

    import "github.com/f2xme/gox/sms/adapter/volcengine"

# 使用示例

    import (
        "github.com/f2xme/gox/sms"
        "github.com/f2xme/gox/sms/adapter/aliyun"
    )
    
    client, err := aliyun.New(
        aliyun.WithAccessKeyID("your-key-id"),
        aliyun.WithAccessKeySecret("your-key-secret"),
        aliyun.WithSignName("your-sign-name"),
    )
    if err != nil {
        // 处理错误
    }
    
    err = client.Send("13800138000", "SMS_123456789", `{"code":"1234"}`)
*/
package sms
```

**2. 适配器文档**
- 每个适配器的 `doc.go` 说明配置项和使用方式
- 包含厂商特定的注意事项

### 用户迁移指南

**旧代码：**
```go
import "github.com/f2xme/gox/sms/adapter/aliyun"

client, err := aliyun.New(
    aliyun.WithAccessKeyID("..."),
    aliyun.WithAccessKeySecret("..."),
    aliyun.WithSignName("..."),
)
```

**新代码：**
```go
import "github.com/f2xme/gox/sms/adapter/aliyun"

client, err := aliyun.New(
    aliyun.WithAccessKeyID("..."),
    aliyun.WithAccessKeySecret("..."),
    aliyun.WithSignName("..."),
)
```

**迁移步骤：**
1. 更新导入路径：`sms/aliyun` → `sms/adapter/aliyun`
2. 运行 `go mod tidy` 更新依赖
3. API 保持不变，无需修改业务代码

## 实现计划

### 阶段 1：准备工作
1. 创建 `sms/adapter/` 目录结构
2. 为每个适配器创建独立的 `go.mod`

### 阶段 2：迁移代码
1. 将 `sms/aliyun/` 代码迁移到 `sms/adapter/aliyun/`
2. 将 `sms/tencent/` 代码迁移到 `sms/adapter/tencent/`
3. 将 `sms/volcengine/` 代码迁移到 `sms/adapter/volcengine/`
4. 更新各适配器的 import 路径

### 阶段 3：清理
1. 删除旧的 `sms/aliyun/`、`sms/tencent/`、`sms/volcengine/` 目录
2. 从主 module 的 `go.mod` 中移除厂商 SDK 依赖
3. 更新 `sms/doc.go` 文档

### 阶段 4：测试与验证
1. 运行所有适配器的单元测试
2. 验证各适配器的 `go.mod` 依赖正确
3. 测试示例代码可以正常运行

### 阶段 5：提交
1. 提交代码变更
2. 为每个适配器打初始 tag（`v0.1.0`）

## 优势总结

1. **按需引入**：用户只需引入需要的适配器，减少依赖体积
2. **解耦依赖**：主 module 不再依赖各厂商 SDK，保持轻量
3. **独立演进**：各适配器可以独立发版，互不影响
4. **清晰架构**：适配器模式使职责更清晰，易于维护和扩展
5. **统一接口**：保持 `sms.SMS` 接口作为统一抽象，业务代码无需关心底层实现

## 潜在风险

1. **版本管理复杂度**：需要管理多个 module 的版本
   - 缓解：使用清晰的版本号规则和发布流程

2. **用户迁移成本**：需要更新导入路径
   - 缓解：提供清晰的迁移指南，API 保持不变

3. **循环依赖风险**：适配器依赖主 module
   - 缓解：确保依赖关系单向（adapter → main），不反向依赖
