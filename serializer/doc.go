// Package serializer 提供统一的序列化接口和多种格式适配器。
//
// 本包封装了常见的序列化格式(JSON、XML、Protobuf、MessagePack),
// 提供统一的 API 接口,支持字节数组和流式操作。
//
// # 功能特性
//
//   - 统一的 Serializer 接口,支持多种序列化格式
//   - JSON 序列化器,基于标准库 encoding/json
//   - XML 序列化器,基于标准库 encoding/xml
//   - Protobuf 序列化器,基于 google.golang.org/protobuf
//   - MessagePack 序列化器,基于 github.com/vmihailenco/msgpack/v5
//   - 支持字节数组操作(Marshal/Unmarshal)
//   - 支持流式操作(Encode/Decode)
//   - 返回标准 MIME 类型(ContentType)
//
// # 快速开始
//
// JSON 序列化示例:
//
//	package main
//
//	import (
//		"fmt"
//		"github.com/f2xme/gox/serializer"
//	)
//
//	type User struct {
//		Name string `json:"name"`
//		Age  int    `json:"age"`
//	}
//
//	func main() {
//		s := serializer.NewJSON()
//
//		// 序列化
//		user := User{Name: "Alice", Age: 30}
//		data, _ := s.Marshal(user)
//		fmt.Println(string(data)) // {"name":"Alice","age":30}
//
//		// 反序列化
//		var u User
//		s.Unmarshal(data, &u)
//		fmt.Printf("%+v\n", u) // {Name:Alice Age:30}
//	}
//
// # 使用其他格式
//
// XML 序列化:
//
//	s := serializer.NewXML()
//	data, _ := s.Marshal(user)
//
// MessagePack 序列化:
//
//	s := serializer.NewMsgPack()
//	data, _ := s.Marshal(user)
//
// Protobuf 序列化(需要实现 proto.Message 接口):
//
//	s := serializer.NewProtobuf()
//	data, _ := s.Marshal(protoMessage)
//
// # 流式操作
//
// 使用 Encode/Decode 进行流式操作:
//
//	var buf bytes.Buffer
//	s := serializer.NewJSON()
//
//	// 编码到流
//	s.Encode(&buf, user)
//
//	// 从流解码
//	var result User
//	s.Decode(&buf, &result)
//
// # 注意事项
//
//   - Protobuf 序列化器要求对象实现 proto.Message 接口
//   - Protobuf Decode 限制最大读取 10MB,防止内存耗尽
//   - 所有序列化器都是无状态的,可以安全地并发使用
//   - 使用 ContentType() 方法获取对应的 MIME 类型
package serializer
