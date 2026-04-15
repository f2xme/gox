#!/bin/bash
# Release 说明生成脚本

set -e

VERSION=${1:-""}
PREVIOUS_VERSION=${2:-""}

if [ -z "$VERSION" ]; then
    echo "用法: $0 <version> [previous_version]"
    exit 1
fi

# 如果没有提供上一个版本，自动获取
if [ -z "$PREVIOUS_VERSION" ]; then
    PREVIOUS_VERSION=$(git describe --tags --abbrev=0 "$VERSION^" 2>/dev/null || echo "")
fi

# 生成变更日志
if [ -n "$PREVIOUS_VERSION" ]; then
    COMMITS=$(git log --pretty=format:"- %s (%h)" "$PREVIOUS_VERSION".."$VERSION" 2>/dev/null || git log --pretty=format:"- %s (%h)" "$VERSION")
    COMPARE_LINK="https://github.com/f2xme/gox/compare/$PREVIOUS_VERSION...$VERSION"
    IS_INITIAL=false
else
    COMMITS=$(git log --pretty=format:"- %s (%h)" "$VERSION" 2>/dev/null || echo "")
    COMPARE_LINK="https://github.com/f2xme/gox/commits/$VERSION"
    IS_INITIAL=true
fi

# 分类变更
FEATURES=$(echo "$COMMITS" | grep -iE "^- (feat|feature|add)" || echo "")
FIXES=$(echo "$COMMITS" | grep -iE "^- (fix|bug)" || echo "")
IMPROVEMENTS=$(echo "$COMMITS" | grep -iE "^- (refactor|perf|optimize|improve)" || echo "")
DOCS=$(echo "$COMMITS" | grep -iE "^- (docs|doc)" || echo "")
BREAKING=$(echo "$COMMITS" | grep -iE "^- (breaking|BREAKING)" || echo "")

# 生成 Release 说明
cat <<EOF
## 📋 变更说明

EOF

# 如果是初始版本，显示特殊说明
if [ "$IS_INITIAL" = true ]; then
    cat <<EOF
🎉 **gox - Go eXtended utilities 初始发布**

这是 gox 的首个版本，提供了 20+ 个常用 Go 工具包的统一封装。

### ✨ 核心特性

- 统一的 API 风格和开箱即用的配置
- 独立的包设计，按需导入
- 完整的测试覆盖和丰富的示例代码
- 基于成熟开源库的二次封装

### 📦 包含的工具包

- **cache** - 缓存操作（内存/Redis）
- **captcha** - 验证码生成
- **config** - 配置管理（Viper）
- **database** - 数据库操作（MySQL/PostgreSQL/SQLite）
- **email** - 邮件服务
- **encrypt** - 加密工具（AES/RSA/Hash）
- **errorx** - 错误处理增强
- **graceful** - 优雅关闭
- **httpx** - HTTP 工具（Gin + 中间件）
- **idgen** - ID 生成器（UUID/Snowflake/ShortID）
- **jwt** - JWT 令牌处理
- **logx** - 日志封装（Zap）
- **metrics** - 指标监控（Prometheus）
- **oss** - 对象存储（阿里云 OSS）
- **pager** - 分页工具
- **payment** - 支付服务（支付宝/微信）
- **queue** - 队列封装
- **ratelimit** - 限流工具
- **sms** - 短信服务（阿里云/腾讯云/火山引擎）
- **timex** - 时间工具
- **trace** - 链路追踪
- **validator** - 数据验证

EOF
else
    if [ -n "$FEATURES" ]; then
        echo "### ✨ 新增功能"
        echo ""
        echo "$FEATURES"
        echo ""
    fi
fi

if [ -n "$FIXES" ]; then
    echo "### 🐛 Bug 修复"
    echo ""
    echo "$FIXES"
    echo ""
fi

if [ -n "$IMPROVEMENTS" ]; then
    echo "### 🔧 优化改进"
    echo ""
    echo "$IMPROVEMENTS"
    echo ""
fi

if [ -n "$DOCS" ]; then
    echo "### 📚 文档更新"
    echo ""
    echo "$DOCS"
    echo ""
fi

if [ -n "$BREAKING" ]; then
    echo "### ⚠️ 破坏性变更"
    echo ""
    echo "$BREAKING"
    echo ""
else
    echo "### ⚠️ 破坏性变更"
    echo ""
    echo "- 无"
    echo ""
fi

cat <<EOF
---

## 📦 安装

\`\`\`bash
go get github.com/f2xme/gox@$VERSION
\`\`\`

## 📖 使用

\`\`\`go
import "github.com/f2xme/gox/<package>"
\`\`\`

## 🔗 相关链接

- [完整变更日志]($COMPARE_LINK)
- [文档](https://github.com/f2xme/gox/blob/$VERSION/README.md)
- [示例代码](https://github.com/f2xme/gox/tree/$VERSION/examples)

## 📊 包含的工具包

cache, captcha, config, database, email, encrypt, errorx, graceful, httpx, idgen, jwt, logx, metrics, oss, pager, payment, queue, ratelimit, sms, timex, trace, validator
EOF
