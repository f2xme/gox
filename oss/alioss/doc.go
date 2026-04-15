/*
Package alioss 提供阿里云 OSS（Object Storage Service）的适配器实现。

# 概述

alioss 包实现了 github.com/f2xme/gox/oss 包定义的统一对象存储接口，
提供对阿里云 OSS 的完整支持，包括对象上传、下载、删除、列表、预签名 URL 等功能。

# 快速开始

基本用法：

	import (
		"context"
		"strings"
		"github.com/f2xme/gox/oss/alioss"
	)

	func main() {
		// 创建 OSS 存储实例
		storage, err := alioss.New(
			"oss-cn-hangzhou.aliyuncs.com",
			"your-access-key-id",
			"your-access-key-secret",
			"your-bucket-name",
		)
		if err != nil {
			panic(err)
		}

		ctx := context.Background()

		// 上传对象
		reader := strings.NewReader("Hello, OSS!")
		err = storage.Put(ctx, "hello.txt", reader)

		// 下载对象
		body, err := storage.Get(ctx, "hello.txt")
		defer body.Close()

		// 删除对象
		err = storage.Delete(ctx, "hello.txt")
	}

# 配置选项

## 基本配置

创建存储实例时需要提供四个必需参数：

	storage, err := alioss.New(
		"oss-cn-hangzhou.aliyuncs.com",  // Endpoint
		"LTAI5t...",                      // Access Key ID
		"abc123...",                      // Access Key Secret
		"my-bucket",                      // Bucket 名称
	)

## 可选配置

使用 Option 函数配置额外选项：

	storage, err := alioss.New(
		endpoint, keyID, keySecret, bucket,
		alioss.WithSecurityToken("STS-token"),  // STS 临时凭证
		alioss.WithEnableCRC(true),             // 启用 CRC 校验
		alioss.WithTimeout(60),                 // 超时时间（秒）
	)

## 从配置文件加载

使用 NewWithConfig 从配置文件加载：

	import "github.com/f2xme/gox/config/adapter/viper"

	cfg := viper.New("config.yaml")
	storage, err := alioss.NewWithConfig(cfg)

配置文件示例（YAML）：

	oss:
	  alioss:
	    endpoint: oss-cn-hangzhou.aliyuncs.com
	    accessKeyID: LTAI5t...
	    accessKeySecret: abc123...
	    bucket: my-bucket
	    securityToken: ""           # 可选
	    enableCRC: false            # 可选
	    timeout: 60                 # 可选，单位：秒

# 核心功能

## 对象操作

上传对象：

	reader := strings.NewReader("content")
	err := storage.Put(ctx, "path/to/file.txt", reader,
		oss.WithContentType("text/plain"),
		oss.WithMetadata(map[string]string{
			"author": "alice",
			"version": "1.0",
		}),
	)

下载对象：

	// 完整下载
	body, err := storage.Get(ctx, "path/to/file.txt")
	defer body.Close()

	// 范围下载
	body, err := storage.Get(ctx, "path/to/file.txt",
		oss.WithRange(0, 1023), // 下载前 1KB
	)

删除对象：

	err := storage.Delete(ctx, "path/to/file.txt")

## 对象元信息

获取对象元信息：

	info, err := storage.Head(ctx, "path/to/file.txt")
	fmt.Println("Size:", info.Size)
	fmt.Println("Content-Type:", info.ContentType)
	fmt.Println("Last-Modified:", info.LastModified)
	fmt.Println("Metadata:", info.Metadata)

检查对象是否存在：

	exists, err := storage.Exists(ctx, "path/to/file.txt")
	if exists {
		fmt.Println("Object exists")
	}

## 列出对象

列出指定前缀的对象：

	objects, err := storage.List(ctx,
		oss.WithPrefix("images/"),
		oss.WithMaxKeys(100),
	)

	for _, obj := range objects {
		fmt.Printf("%s - %d bytes\n", obj.Key, obj.Size)
	}

## 预签名 URL

生成预签名 URL 用于临时访问：

	// GET 预签名 URL（下载）
	url, err := storage.PresignedURL(ctx, "path/to/file.txt",
		oss.WithMethod(oss.MethodGet),
		oss.WithExpires(15*time.Minute),
	)

	// PUT 预签名 URL（上传）
	url, err := storage.PresignedURL(ctx, "path/to/file.txt",
		oss.WithMethod(oss.MethodPut),
		oss.WithExpires(15*time.Minute),
	)

## 存储桶操作

创建存储桶：

	err := storage.CreateBucket(ctx, "new-bucket")

删除存储桶：

	err := storage.DeleteBucket(ctx, "old-bucket")

列出所有存储桶：

	buckets, err := storage.ListBuckets(ctx)
	for _, bucket := range buckets {
		fmt.Printf("%s - %s\n", bucket.Name, bucket.CreationDate)
	}

# 错误处理

alioss 包将阿里云 OSS 错误转换为统一的 oss.Error：

	_, err := storage.Get(ctx, "nonexistent.txt")
	if ossErr, ok := err.(*oss.Error); ok {
		switch ossErr.Code {
		case oss.ErrCodeNotFound:
			fmt.Println("Object not found")
		case oss.ErrCodeAccessDenied:
			fmt.Println("Access denied")
		case oss.ErrCodeInvalidArgument:
			fmt.Println("Invalid argument")
		default:
			fmt.Println("Other error:", ossErr.Message)
		}
	}

# 最佳实践

## 1. 使用 context 控制超时

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := storage.Put(ctx, "large-file.bin", reader)

## 2. 合理设置对象元数据

	err := storage.Put(ctx, "document.pdf", reader,
		oss.WithContentType("application/pdf"),
		oss.WithMetadata(map[string]string{
			"author": "alice",
			"department": "engineering",
			"version": "2.0",
		}),
	)

## 3. 使用预签名 URL 实现安全的临时访问

	// 生成 15 分钟有效的下载链接
	url, err := storage.PresignedURL(ctx, "private/report.pdf",
		oss.WithMethod(oss.MethodGet),
		oss.WithExpires(15*time.Minute),
	)

	// 将 URL 返回给客户端，客户端可直接下载

## 4. 使用 STS 临时凭证提高安全性

对于客户端直传场景，使用 STS 临时凭证：

	storage, err := alioss.New(
		endpoint, keyID, keySecret, bucket,
		alioss.WithSecurityToken(stsToken),
	)

## 5. 启用 CRC 校验确保数据完整性

	storage, err := alioss.New(
		endpoint, keyID, keySecret, bucket,
		alioss.WithEnableCRC(true),
	)

# 性能考虑

  - 上传大文件时考虑使用分片上传（未来版本支持）
  - 使用范围下载减少带宽消耗
  - 合理设置超时时间避免长时间阻塞
  - 使用预签名 URL 让客户端直接访问 OSS，减轻服务器压力

# 线程安全

Storage 实例是线程安全的，可以在多个 goroutine 中并发使用。

# 区域和 Endpoint

阿里云 OSS 支持多个区域，常用 Endpoint：

  - 华东1（杭州）：oss-cn-hangzhou.aliyuncs.com
  - 华东2（上海）：oss-cn-shanghai.aliyuncs.com
  - 华北1（青岛）：oss-cn-qingdao.aliyuncs.com
  - 华北2（北京）：oss-cn-beijing.aliyuncs.com
  - 华南1（深圳）：oss-cn-shenzhen.aliyuncs.com

完整列表请参考阿里云官方文档。
*/
package alioss
