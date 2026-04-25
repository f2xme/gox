// Package oss 提供统一的对象存储抽象。
//
// oss 包只定义对象存储的通用接口、选项、错误模型和数据结构，不绑定具体云厂商。
// 具体实现放在 adapter 子包中，例如 github.com/f2xme/gox/oss/adapter/alioss。
//
// # 功能特性
//
//   - 统一接口：抽象上传、下载、删除、元信息、列表和预签名 URL
//   - 能力分层：基础对象操作与存储桶管理分离
//   - 分页列表：List 返回对象、公共前缀、下一页令牌和截断状态
//   - 错误模型：使用统一错误码，并提供 IsNotFound、IsAccessDenied 等判断函数
//   - Options 模式：上传、下载、列表、预签名和存储桶操作都使用函数式选项
//   - Adapter 扩展：新增云厂商实现时只需实现 Storage 接口
//
// # 快速开始
//
// 使用阿里云 OSS adapter：
//
//	package main
//
//	import (
//		"context"
//		"strings"
//		"time"
//
//		"github.com/f2xme/gox/oss"
//		"github.com/f2xme/gox/oss/adapter/alioss"
//	)
//
//	func main() {
//		storage, err := alioss.New(
//			alioss.WithEndpoint("oss-cn-hangzhou.aliyuncs.com"),
//			alioss.WithCredentials("access-key-id", "access-key-secret"),
//			alioss.WithBucket("my-bucket"),
//		)
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
//
//		_, _ = storage.SignURL(ctx, "hello.txt",
//			oss.WithMethod(oss.MethodGet),
//			oss.WithExpires(time.Hour),
//		)
//	}
//
// # 核心接口
//
// Storage 定义对象级操作：
//
//	type Storage interface {
//		Put(ctx context.Context, key string, reader io.Reader, opts ...PutOption) error
//		Get(ctx context.Context, key string, opts ...GetOption) (io.ReadCloser, error)
//		Delete(ctx context.Context, key string) error
//		Stat(ctx context.Context, key string) (*ObjectInfo, error)
//		Exists(ctx context.Context, key string) (bool, error)
//		List(ctx context.Context, opts ...ListOption) (*ListResult, error)
//		SignURL(ctx context.Context, key string, opts ...SignOption) (string, error)
//	}
//
// BucketStorage 定义可选的存储桶管理能力。业务代码只在确实需要创建、删除或列出
// Bucket 时依赖该接口，普通对象读写只依赖 Storage。
//
// # 列表分页
//
// List 返回 ListResult：
//
//	result, err := storage.List(ctx,
//		oss.WithPrefix("photos/"),
//		oss.WithDelimiter("/"),
//		oss.WithLimit(100),
//	)
//	if err != nil {
//		return
//	}
//
//	for _, obj := range result.Objects {
//		_ = obj.Key
//	}
//	if result.Truncated {
//		next, _ := storage.List(ctx, oss.WithToken(result.NextToken))
//		_ = next
//	}
//
// # 错误处理
//
// Adapter 会尽量把云厂商错误转换为 *oss.Error：
//
//	_, err := storage.Get(ctx, "missing.txt")
//	if oss.IsNotFound(err) {
//		// 对象不存在
//	} else if oss.IsAccessDenied(err) {
//		// 权限不足
//	}
//
// # Adapter 约定
//
// 新增 adapter 时建议放在 oss/adapter/<provider> 下，使用纯 Options 模式创建实例，
// 不依赖 gox 内部其他包，并实现 Storage。若 provider 支持 Bucket 管理，再额外实现
// BucketStorage。
//
// # 注意事项
//
//   - Storage 实现应当是并发安全的
//   - key 的组织方式由业务决定，oss 包不会修改 key
//   - 预签名 URL 的方法必须使用 MethodGet、MethodPut 或 MethodDelete
//   - 集成测试应显式依赖环境变量，避免默认访问真实云服务
package oss
