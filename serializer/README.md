# serializer

统一的序列化接口和多种格式适配器。

## 特性

- 统一的 `Serializer` 接口
- 支持 JSON、XML、Protobuf、MessagePack
- 流式编码/解码支持
- 返回标准 MIME 类型

## 安装

```bash
go get github.com/f2xme/gox/serializer
```

## 使用示例

### JSON 序列化

```go
package main

import (
    "fmt"
    "github.com/f2xme/gox/serializer"
)

type User struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func main() {
    s := serializer.NewJSON()
    
    user := User{Name: "Alice", Age: 30}
    data, _ := s.Marshal(user)
    fmt.Println(string(data)) // {"name":"Alice","age":30}
    
    var u User
    s.Unmarshal(data, &u)
    fmt.Printf("%+v\n", u) // {Name:Alice Age:30}
}
```

### XML 序列化

```go
s := serializer.NewXML()
data, _ := s.Marshal(user)
```

### MessagePack 序列化

```go
s := serializer.NewMsgPack()
data, _ := s.Marshal(user)
```

### Protobuf 序列化

```go
// 需要实现 proto.Message 接口
s := serializer.NewProtobuf()
data, _ := s.Marshal(protoMessage)
```

### 流式编码

```go
var buf bytes.Buffer
s := serializer.NewJSON()
s.Encode(&buf, user)
```

## 依赖

- `google.golang.org/protobuf` - Protobuf 支持
- `github.com/vmihailenco/msgpack/v5` - MessagePack 支持
