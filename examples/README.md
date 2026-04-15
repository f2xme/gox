# gox 使用示例

本目录包含 gox 项目各个包的使用示例，帮助您快速理解和上手各个模块。

## 可用示例

### 核心包

- **httpx** - HTTP 服务器示例（使用 Gin 适配器）
  - 路径: `examples/httpx/main.go`
  - 功能: 演示路由定义、请求处理、参数绑定、统一响应格式、优雅关闭
  - 特点: 完整的 REST API 示例，包含路由组、中间件使用

- **logx** - 日志记录示例（使用 Zap 适配器）
  - 路径: `examples/logx/main.go`
  - 功能: 演示不同日志级别、结构化日志、字段记录、业务流程日志
  - 特点: 展示如何在实际业务中使用结构化日志

- **jwt** - JWT 令牌示例
  - 路径: `examples/jwt/main.go`
  - 功能: 演示 token 生成、解析、验证、刷新、多种算法
  - 特点: 包含 HMAC 算法示例和完整的错误处理

### 工具包

- **cache** - 缓存操作示例
  - 路径: `examples/cache/main.go`
  - 功能: 演示 Set/Get/Delete 操作、过期时间、批量操作
  - 特点: 使用内存适配器，无需外部依赖

- **config** - 配置管理示例
  - 路径: `examples/config/main.go`
  - 功能: 演示配置文件加载、多种数据类型读取、实际应用场景
  - 特点: 自动创建示例配置文件，展示完整的配置管理流程

- **database** - 数据库操作示例
  - 路径: `examples/database/main.go`
  - 功能: 演示 CRUD 操作、事务处理、自动迁移
  - 特点: 使用 SQLite，无需外部数据库服务

- **errorx** - 错误处理示例
  - 路径: `examples/errorx/main.go`
  - 功能: 演示错误创建、包装、类型判断、元数据附加
  - 特点: 展示多种错误类型和实际应用场景

- **idgen** - ID 生成示例
  - 路径: `examples/idgen/main.go`
  - 功能: 演示 UUID/ULID/Snowflake ID 生成、解析、并发安全测试
  - 特点: 对比不同 ID 类型的特点和适用场景

- **validator** - 数据验证示例
  - 路径: `examples/validator/main.go`
  - 功能: 演示结构体验证、自定义验证规则、错误消息翻译
  - 特点: 包含多种验证场景和自定义验证器实例

- **captcha** - 验证码示例
  - 路径: `examples/captcha/main.go`
  - 功能: 演示数字、字母、算术、音频验证码生成和验证
  - 特点: 展示多种验证码类型和完整的验证流程

- **encrypt** - 加密工具示例
  - 路径: `examples/encrypt/main.go`
  - 功能: 演示哈希（MD5/SHA256/Blake3）、AES-GCM 加密、RSA 加密
  - 特点: 涵盖常用加密算法和实际应用场景

- **graceful** - 优雅关闭示例
  - 路径: `examples/graceful/main.go`
  - 功能: 演示 HTTP 服务器、数据库、定时任务等资源的优雅关闭
  - 特点: 展示优先级控制和超时管理

- **metrics** - 指标监控示例
  - 路径: `examples/metrics/main.go`
  - 功能: 演示 Counter、Gauge、Histogram 指标类型
  - 特点: Mock 实现，展示指标监控的基本用法

- **sms** - 短信服务示例
  - 路径: `examples/sms/main.go`
  - 功能: 演示短信发送接口（验证码、订单通知）
  - 特点: Mock 实现，展示短信服务的抽象层

- **oss** - 对象存储示例
  - 路径: `examples/oss/main.go`
  - 功能: 说明对象存储的典型用法和可用适配器
  - 特点: 简化示例，需要配置真实的存储服务

- **pager** - 分页工具示例
  - 路径: `examples/pager/main.go`
  - 功能: 演示 Offset、Page、Cursor 三种分页策略
  - 特点: 模拟数据查询，展示不同分页方式的使用

- **payment** - 支付服务示例
  - 路径: `examples/payment/main.go`
  - 功能: 说明支付流程和可用适配器
  - 特点: 简化示例，需要配置真实的支付服务商

- **queue** - 消息队列示例
  - 路径: `examples/queue/main.go`
  - 功能: 演示消息发布、订阅、处理流程
  - 特点: 使用内存适配器，无需外部依赖

- **ratelimit** - 限流示例
  - 路径: `examples/ratelimit/main.go`
  - 功能: 演示令牌桶和滑动窗口限流算法
  - 特点: 展示不同限流策略的使用和效果

- **timex** - 时间工具示例
  - 路径: `examples/timex/main.go`
  - 功能: 演示时间格式化、解析、计算、工作日判断
  - 特点: 涵盖常用时间操作场景

- **trace** - 链路追踪示例
  - 路径: `examples/trace/main.go`
  - 功能: 演示分布式链路追踪的基本用法
  - 特点: 展示 Span 创建、上下文传递、追踪信息记录

## 运行示例

### 运行单个示例

```bash
# 运行 httpx 示例（HTTP 服务器，按 Ctrl+C 停止）
go run examples/httpx/main.go

# 运行 logx 示例
go run examples/logx/main.go

# 运行 jwt 示例
go run examples/jwt/main.go

# 运行 cache 示例
go run examples/cache/main.go

# 运行 config 示例
go run examples/config/main.go

# 运行 database 示例
go run examples/database/main.go

# 运行 errorx 示例
go run examples/errorx/main.go

# 运行 idgen 示例
go run examples/idgen/main.go

# 运行 validator 示例
go run examples/validator/main.go

# 运行 captcha 示例
go run examples/captcha/main.go

# 运行 encrypt 示例
go run examples/encrypt/main.go

# 运行 graceful 示例
go run examples/graceful/main.go

# 运行 metrics 示例
go run examples/metrics/main.go

# 运行 sms 示例
go run examples/sms/main.go

# 运行 oss 示例
go run examples/oss/main.go

# 运行 pager 示例
go run examples/pager/main.go

# 运行 payment 示例
go run examples/payment/main.go

# 运行 queue 示例
go run examples/queue/main.go

# 运行 ratelimit 示例
go run examples/ratelimit/main.go

# 运行 timex 示例
go run examples/timex/main.go

# 运行 trace 示例
go run examples/trace/main.go
```

### 编译所有示例

```bash
# 编译所有示例（验证代码正确性）
go build ./examples/...

# 或者逐个编译
go build ./examples/httpx
go build ./examples/logx
go build ./examples/jwt
go build ./examples/cache
go build ./examples/config
go build ./examples/database
go build ./examples/errorx
go build ./examples/idgen
go build ./examples/validator
go build ./examples/captcha
go build ./examples/encrypt
go build ./examples/graceful
go build ./examples/metrics
go build ./examples/sms
go build ./examples/oss
go build ./examples/pager
go build ./examples/payment
go build ./examples/queue
go build ./examples/ratelimit
go build ./examples/timex
go build ./examples/trace
```

### 批量运行示例

```bash
# 运行所有非服务器示例（快速验证）
for dir in logx jwt cache config database errorx idgen validator; do
  echo "=== Running $dir example ==="
  go run examples/$dir/main.go
  echo ""
done
```

## 依赖安装

示例代码依赖 gox 项目的各个包，运行前请确保依赖已安装：

```bash
# 在项目根目录执行
go mod tidy
```

## 注意事项

- 所有示例都是独立的可运行程序
- 示例使用内存实现或模拟数据，不依赖外部服务
- 示例代码包含详细注释，便于理解
- 建议按照示例代码学习各包的使用方法
- httpx 示例会启动 HTTP 服务器，需要手动停止（Ctrl+C）
- database 和 config 示例会创建临时文件，运行后会自动清理

## 常见问题

**Q: 示例无法编译？**  
A: 确保在项目根目录执行 `go mod tidy` 安装所有依赖。

**Q: httpx 示例端口被占用？**  
A: 修改 `examples/httpx/main.go` 中的端口号（默认 8080）。

**Q: 如何在自己的项目中使用这些包？**  
A: 参考示例代码，使用 `go get github.com/f2xme/gox/<package>` 安装需要的包。

**Q: 示例代码可以直接用于生产环境吗？**  
A: 示例代码仅用于演示，生产环境需要根据实际需求调整配置和错误处理。

## 更多信息

详细的 API 文档和使用说明，请参考项目根目录的 README.md 文件。
