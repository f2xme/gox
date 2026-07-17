# idverify

身份证姓名 + 证件号二要素核验抽象。

## 安装

```bash
go get github.com/f2xme/gox@latest
# 阿里云适配器为独立 module：
go get github.com/f2xme/gox/idverify/adapter/aliyun@latest
```

## 接口

```go
type Verifier interface {
    Provider() string
    Verify(ctx context.Context, req Request) (Result, error)
}
```

- 业务不匹配：`Result.Matched == false` 且 `error == nil`
- 系统错误：返回 `error`（可用 `errors.Is` 判断 `ErrNotConfigured` / `ErrUnavailable` 等）

## 适配器

| 包 | 说明 |
| --- | --- |
| `idverify/adapter/mock` | 内存实现，默认通过 |
| `idverify/adapter/baidu` | 百度 `person/idmatch` |
| `idverify/adapter/aliyun` | 阿里云 `Id2MetaVerify`（独立 go.mod） |

## 示例

```go
import (
    "context"
    "github.com/f2xme/gox/idverify"
    "github.com/f2xme/gox/idverify/adapter/mock"
)

v, _ := mock.New(mock.WithMismatchNames("bad-name"))
res, err := v.Verify(context.Background(), idverify.Request{
    Name: "张三", IDNumber: "110101199001011234",
})
```
