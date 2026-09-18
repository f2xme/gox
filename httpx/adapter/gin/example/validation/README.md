# Gin 多语言校验

使用当前仓库代码（包含 `FieldError.StructNamespace`）：

```sh
# 仓库根目录
go run ./httpx/adapter/gin/example/validation
```

中文必填提示：

```sh
curl -i http://127.0.0.1:8080/users \
  -H 'Content-Type: application/json' \
  -H 'Accept-Language: zh-CN' \
  -d '{}'
# HTTP 400，Content-Language: zh
# {"message":"名称为必填字段"}
```

按权重选英文，并插入校验规则参数：

```sh
curl -i http://127.0.0.1:8080/users \
  -H 'Content-Type: application/json' \
  -H 'Accept-Language: zh-CN;q=0.1,en-US;q=0.9' \
  -d '{"name":"A"}'
# HTTP 400，Content-Language: en
# {"message":"Name must contain at least 2 characters"}
```

`{"name":"张三"}` 返回 HTTP 204；示例不实际创建用户。

## 接入方式

1. 请求字段使用 `validate` 标签。`BindJSON` 自动校验，错误原样返回或使用 `httpx.ErrBadRequest("请求参数无效", err)` 包装，保留错误链。
2. `SetErrorHandler` 内通过 `validator.AsValidationError(err)` 提取字段错误，其他错误交给 `httpx.DefaultErrorHandler`。
3. 使用 `language.MatchStrings` 解析请求的 `Accept-Language`，支持地区标签和 `q` 权重；缺失或无匹配时默认中文。读取请求头用 `c.Header`，写响应头用 `c.SetHeader`。
4. 使用 `StructNamespace + "." + Tag` 查找业务语言包，例如 `CreateUserRequest.Name.required`；用 `Param` 替换 `{param}`。不需要再次调用 `ValidateWithLang`。

语言包见 [messages.json](messages.json)，程序启动时加载，此后只读。增加语言时同步增加语言包、`languages` 和 `matcher` 中对应项，保持两者顺序一致；业务语言不受 validator 内置语言范围限制。示例仅翻译校验错误，JSON 解析错误及其他业务错误沿用默认错误处理。

缺少字段翻译时使用所选语言的 `validation.failed`，每种语言都应保留该项。`StructNamespace` 使用原始 Go 类型和字段名，不受 `json`、`label` 标签影响；重命名类型或字段后需同步修改语言包。嵌套字段会得到 `CreateUserRequest.Billing.Name` 这样的路径；集合路径带索引或键，如 `CreateUserRequest.Items[0].Name`，业务按需归一化后再查语言包。

## 验证

```sh
go test ./httpx/adapter/gin/example/validation
```

测试覆盖实际 HTTP 请求的语言匹配、规则参数、缺失翻译、JSON 解析失败、服务错误及校验通过。
