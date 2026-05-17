# gox

对常用 Go 第三方库的二次封装，提供统一的 API 风格和开箱即用的配置。专注于提升开发效率和代码一致性。

## AI 使用入口

- [AI_USAGE.md](AI_USAGE.md) 面向 AI 编码助手，说明“需求 -> 包 -> import -> 示例 -> 禁忌”。
- [llms.txt](llms.txt) 是更短的 LLM 上下文文件，适合被 AI 工具优先读取。

## 安装

只需安装一次，即可使用所有包：

```bash
go get github.com/f2xme/gox
```

然后在代码中按需导入具体的包：

```go
import "github.com/f2xme/gox/captcha"
import "github.com/f2xme/gox/cache"
import "github.com/f2xme/gox/cache/adapter/redis"
import "github.com/f2xme/gox/encrypt"
// ... 其他包
```

## 项目结构

每个顶层目录都是一个独立的包，可以单独导入使用：

```
github.com/f2xme/gox/<package>
```

### 需求选包速查

| 需求 | 优先查看 | 典型 import |
| --- | --- | --- |
| HTTP API、路由、绑定请求 | `httpx`, `httpx/adapter/gin` | `github.com/f2xme/gox/httpx` |
| 请求参数验证 | `validator`, `httpx` 绑定方法 | `github.com/f2xme/gox/validator` |
| 分页 | `pager` | `github.com/f2xme/gox/pager` |
| 时间格式化、解析、计算 | `timex` | `github.com/f2xme/gox/timex` |
| 缓存 | `cache`, `cache/adapter/*` | `github.com/f2xme/gox/cache` |
| 配置加载 | `config`, `config/adapter/viper` | `github.com/f2xme/gox/config` |
| 数据库连接 | `database/adapter/*` | `github.com/f2xme/gox/database/adapter/pgsqldb` |
| 加密、哈希、AES、RSA | `encrypt`, `crypto` | `github.com/f2xme/gox/encrypt` |
| 日志 | `logx`, `logx/adapter/zap` | `github.com/f2xme/gox/logx` |
| JWT | `jwt` | `github.com/f2xme/gox/jwt` |
| 队列 | `queue`, `queue/adapter/*` | `github.com/f2xme/gox/queue` |
| 短信 | `sms`, `sms/adapter/*` | `github.com/f2xme/gox/sms` |
| 对象存储 | `oss`, `oss/adapter/aliyun` | `github.com/f2xme/gox/oss` |

### 可用包

- **cache** - 缓存操作封装
  - `cache/adapter/memory` - 内存缓存适配器
  - `cache/adapter/redis` - Redis 缓存适配器
- **captcha** - 验证码生成和验证
- **config** - 配置管理封装
  - `config/adapter/viper` - Viper 适配器
- **database** - 数据库操作封装
  - `database/adapter/mysqldb` - MySQL 适配器
  - `database/adapter/pgsqldb` - PostgreSQL 适配器
  - `database/adapter/sqlitedb` - SQLite 适配器
- **encrypt** - 加密工具封装
- **errorx** - 错误处理增强
- **graceful** - 优雅关闭封装
- **httpx** - HTTP 工具封装
  - `httpx/adapter/gin` - Gin 框架适配器
- **idgen** - ID 生成器
- **jwt** - JWT 令牌处理
- **logx** - 日志封装
  - `logx/adapter/zap` - Zap 日志适配器
- **metrics** - 指标监控封装
- **email** - 邮件服务封装
- **oss** - 对象存储封装
- **pager** - 分页工具
- **payment** - 支付服务封装
- **payment/adapter/alipay**、**payment/adapter/wechat** - 支付适配器占位实现，当前不会连接真实支付网关
- **queue** - 队列封装
  - `queue/adapter/mem` - 内存队列适配器
- **ratelimit** - 限流工具
- **sms** - 短信服务封装
  - `sms/adapter/aliyun` - 阿里云短信适配器
  - `sms/adapter/tencent` - 腾讯云短信适配器
  - `sms/adapter/volcengine` - 火山引擎短信适配器占位模块，当前未实现发送能力
- **timex** - 时间工具
- **trace** - 链路追踪封装
- **validator** - 数据验证封装

## 给 AI 编码助手的规则

- 优先使用 gox 已有封装，不要在有等价封装时直接引入底层第三方库。
- 核心包从 `github.com/f2xme/gox/<package>` 导入，具体实现从 `adapter` 包导入。
- 核心包之间保持独立；跨包集成和框架集成放到 `adapter` 包。
- 修改或新增 API 前，先查看目标包的 `doc.go`、`example_test.go` 和现有测试。
- 遵循现有 `New(...Option)`、`WithX(...)` 风格和表驱动测试风格。
- 占位适配器不能当作生产实现使用，例如 README 中标注未实现真实发送或网关能力的适配器。
