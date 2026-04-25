// Package alioss 提供阿里云 OSS adapter。
//
// alioss 包实现 github.com/f2xme/gox/oss 的 Storage 和 BucketStorage 接口，
// 负责把统一的对象存储 API 映射到阿里云 OSS Go SDK。
//
// # 功能特性
//
//   - 对象操作：上传、下载、删除、存在性检查和元信息读取
//   - 对象列表：支持前缀、分隔符、分页令牌和最大返回数量
//   - 预签名 URL：支持 GET、PUT 和 DELETE 方法
//   - Bucket 管理：支持创建、删除和列出存储桶
//   - 统一错误：将阿里云 OSS 错误转换为 oss.Error
//   - Options 模式：构造函数和操作选项都使用函数式选项
//
// # 快速开始
//
// 基本用法：
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
//		err = storage.Put(ctx, "hello.txt", strings.NewReader("hello"),
//			oss.WithContentType("text/plain"),
//		)
//		if err != nil {
//			return
//		}
//
//		body, err := storage.Get(ctx, "hello.txt")
//		if err == nil {
//			defer body.Close()
//		}
//	}
//
// # 构造选项
//
// New 必须提供 Endpoint、访问凭证和默认 Bucket：
//
//	storage, err := alioss.New(
//		alioss.WithEndpoint("oss-cn-hangzhou.aliyuncs.com"),
//		alioss.WithCredentials("access-key-id", "access-key-secret"),
//		alioss.WithBucket("my-bucket"),
//		alioss.WithSecurityToken("sts-token"),
//		alioss.WithEnableCRC(true),
//		alioss.WithTimeout(60),
//	)
//
// 也可以从 config.Config 读取配置，默认配置前缀为 oss：
//
//	storage, err := alioss.NewWithConfig(cfg)
//
// 初始化阶段可以使用 MustNew、MustNewWithOptions 或 MustNewWithConfig，在创建失败时直接终止程序。
//
// # 对象列表
//
//	result, err := storage.List(ctx,
//		oss.WithPrefix("images/"),
//		oss.WithDelimiter("/"),
//		oss.WithLimit(100),
//	)
//	if err != nil {
//		return
//	}
//	for _, obj := range result.Objects {
//		_ = obj.Key
//	}
//
// # 预签名 URL
//
//	url, err := storage.SignURL(ctx, "uploads/avatar.png",
//		oss.WithMethod(oss.MethodPut),
//		oss.WithExpires(15*time.Minute),
//		oss.WithSignContentType("image/png"),
//	)
//
// # 错误处理
//
//	_, err := storage.Get(ctx, "missing.txt")
//	if oss.IsNotFound(err) {
//		// 对象不存在
//	}
//
// # 集成测试
//
// 默认测试不会访问真实阿里云 OSS。需要运行集成测试时设置：
//
//   - GOX_ALIOSS_ENDPOINT
//   - GOX_ALIOSS_ACCESS_KEY_ID
//   - GOX_ALIOSS_ACCESS_KEY_SECRET
//   - GOX_ALIOSS_BUCKET
//   - GOX_ALIOSS_SECURITY_TOKEN（可选）
package alioss
