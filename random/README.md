# random

安全的随机字符生成包，基于 `crypto/rand` 提供密码学安全的随机性。

## 功能特性

- 密码学安全的随机生成
- 多种预定义字符集
- 支持自定义字符集
- 并发安全
- 零依赖

## 安装

```bash
go get github.com/f2xme/gox/random
```

## 快速开始

```go
package main

import (
    "fmt"
    "github.com/f2xme/gox/random"
)

func main() {
    // 生成 10 位随机数字
    code, _ := random.Numeric(10)
    fmt.Println(code) // 例如: 8472951063

    // 生成 16 位字母数字混合
    token, _ := random.AlphaNumeric(16)
    fmt.Println(token) // 例如: aB3xK9mP2qR7sT4u

    // 生成 8 位小写字母
    id, _ := random.AlphaLower(8)
    fmt.Println(id) // 例如: xkpmqrst

    // 使用自定义字符集
    custom, _ := random.String(12, "ABCDEF0123456789")
    fmt.Println(custom) // 例如: A3F0B2E1D4C5
}
```

## API 文档

### 预定义字符集

```go
const (
    Digits       = "0123456789"
    LowerLetters = "abcdefghijklmnopqrstuvwxyz"
    UpperLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    Letters      = LowerLetters + UpperLetters
    Alphanumeric = Letters + Digits
)
```

### 函数

#### String

```go
func String(length int, charset string) (string, error)
```

生成指定长度的随机字符串，使用给定的字符集。

#### Numeric

```go
func Numeric(length int) (string, error)
```

生成指定长度的随机数字字符串（0-9）。

#### Alpha

```go
func Alpha(length int) (string, error)
```

生成指定长度的随机字母字符串（大小写混合）。

#### AlphaLower

```go
func AlphaLower(length int) (string, error)
```

生成指定长度的随机小写字母字符串。

#### AlphaUpper

```go
func AlphaUpper(length int) (string, error)
```

生成指定长度的随机大写字母字符串。

#### AlphaNumeric

```go
func AlphaNumeric(length int) (string, error)
```

生成指定长度的随机字母数字字符串。

## 使用场景

- 验证码生成
- 临时密码
- API Token
- 会话 ID
- 邀请码
- 短链接 ID

## 注意事项

- 所有函数使用 `crypto/rand` 提供密码学安全的随机性
- 适用于安全敏感场景（密码、令牌等）
- 如果 `length <= 0` 或 `charset` 为空，返回空字符串
- 错误通常来自底层的 `crypto/rand`，在正常系统上极少发生
