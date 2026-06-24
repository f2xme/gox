// Package memory 提供基于内存的 oss.Storage 和 oss.BucketStorage 实现。
//
// memory 包用于单进程测试和本地开发，不访问真实云服务。它实现对象上传、
// 下载、删除、元信息、存在性检查、列表、Bucket 管理和测试用预签名 URL。
//
// # 功能特性
//
//   - 对象操作：支持上传、下载、删除、存在性检查和元信息读取
//   - 列表查询：支持前缀、分隔符、分页令牌和最大返回数量
//   - Bucket 管理：支持创建、删除和列出内存 Bucket
//   - 测试友好：进程内存储，零外部依赖，不访问真实云服务
//   - 统一错误：使用 oss.Error 返回 NotFound、InvalidArgument 等错误码
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import (
//		"context"
//		"strings"
//
//		"github.com/f2xme/gox/oss"
//		"github.com/f2xme/gox/oss/adapter/memory"
//	)
//
//	func main() {
//		storage, err := memory.New()
//		if err != nil {
//			return
//		}
//
//		ctx := context.Background()
//		_ = storage.Put(ctx, "hello.txt", strings.NewReader("hello"),
//			oss.WithContentType("text/plain"),
//		)
//
//		body, err := storage.Get(ctx, "hello.txt")
//		if err == nil {
//			defer body.Close()
//		}
//	}
//
// # 注意事项
//
//   - 仅适用于测试、本地开发和单进程场景
//   - 进程退出后数据会丢失
//   - 对象操作只作用于默认 Bucket，可通过 WithBucketName 设置
//   - SignURL 返回测试用 URL，不具备真实签名能力
package memory
