/*
Package oss 提供统一的对象存储抽象层。

oss 包定义了对象存储的标准接口，支持多种对象存储服务（阿里云 OSS、腾讯云 COS、AWS S3、MinIO 等）。
通过统一的 API，你可以轻松地在不同的对象存储服务之间切换，而无需修改业务代码。

# 功能特性

  - 统一接口：支持多种对象存储服务的统一抽象
  - 基础操作：上传、下载、删除、检查存在性
  - 元信息管理：获取对象详细信息和自定义元数据
  - 列表功能：支持前缀过滤和分页的对象列表
  - 预签名 URL：生成临时访问链接，支持直传
  - Bucket 管理：创建、删除和列出存储桶
  - Options 模式：灵活的函数式配置选项

# 快速开始

基本使用：

	package main

	import (
		"context"
		"strings"

		"github.com/f2xme/gox/oss/alioss"
	)

	func main() {
		// 创建存储实例
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

# 核心接口

## Storage - 对象存储接口

所有对象存储实现都必须实现此接口：

	type Storage interface {
		// 基础对象操作
		Put(ctx context.Context, key string, reader io.Reader, opts ...PutOption) error
		Get(ctx context.Context, key string, opts ...GetOption) (io.ReadCloser, error)
		Delete(ctx context.Context, key string) error
		Head(ctx context.Context, key string) (*ObjectInfo, error)
		Exists(ctx context.Context, key string) (bool, error)

		// 对象列表
		List(ctx context.Context, opts ...ListOption) ([]*Object, error)

		// Bucket 操作
		CreateBucket(ctx context.Context, bucket string) error
		DeleteBucket(ctx context.Context, bucket string) error
		ListBuckets(ctx context.Context) ([]*Bucket, error)

		// 预签名 URL
		PresignedURL(ctx context.Context, key string, opts ...PresignedOption) (string, error)
	}

# 使用示例

## 上传文件

	// 从文件上传
	file, _ := os.Open("photo.jpg")
	defer file.Close()

	err := storage.Put(ctx, "photos/2024/photo.jpg", file,
		oss.WithContentType("image/jpeg"),
		oss.WithACL("public-read"),
	)

	// 从字节上传
	data := []byte("hello world")
	err := storage.Put(ctx, "files/hello.txt", bytes.NewReader(data),
		oss.WithContentType("text/plain"),
	)

## 下载文件

	// 获取对象
	reader, err := storage.Get(ctx, "photos/2024/photo.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	// 保存到本地文件
	file, _ := os.Create("downloaded.jpg")
	io.Copy(file, reader)

## 删除文件

	err := storage.Delete(ctx, "photos/2024/photo.jpg")

## 检查文件是否存在

	exists, err := storage.Exists(ctx, "photos/2024/photo.jpg")
	if exists {
		fmt.Println("文件存在")
	}

## 获取文件信息

	info, err := storage.Head(ctx, "photos/2024/photo.jpg")
	fmt.Println(info.Size)         // 文件大小
	fmt.Println(info.ContentType)  // 内容类型
	fmt.Println(info.LastModified) // 最后修改时间

## 列出文件

	// 列出所有文件
	objects, err := storage.List(ctx)

	// 列出指定前缀的文件
	objects, err := storage.List(ctx,
		oss.WithPrefix("photos/2024/"),
		oss.WithMaxKeys(100),
	)

	for _, obj := range objects {
		fmt.Printf("%s (%d bytes)\n", obj.Key, obj.Size)
	}

## 生成预签名 URL

	// 生成下载 URL（有效期 1 小时）
	url, err := storage.PresignedURL(ctx, "photos/2024/photo.jpg",
		oss.WithExpires(time.Hour),
		oss.WithMethod("GET"),
	)

	// 生成上传 URL
	url, err := storage.PresignedURL(ctx, "uploads/new-file.jpg",
		oss.WithExpires(30*time.Minute),
		oss.WithMethod("PUT"),
	)

## Bucket 操作

	// 创建 Bucket
	err := storage.CreateBucket(ctx, "my-bucket")

	// 删除 Bucket
	err := storage.DeleteBucket(ctx, "my-bucket")

	// 列出所有 Bucket
	buckets, err := storage.ListBuckets(ctx)
	for _, bucket := range buckets {
		fmt.Println(bucket.Name)
	}

# 可用适配器

## 阿里云 OSS

	import "github.com/f2xme/gox/oss/adapter/aliossadapter"

	storage := aliossadapter.New(
		aliossadapter.WithEndpoint("oss-cn-hangzhou.aliyuncs.com"),
		aliossadapter.WithAccessKey("your-access-key"),
		aliossadapter.WithSecretKey("your-secret-key"),
		aliossadapter.WithBucket("my-bucket"),
	)

## 腾讯云 COS

	import "github.com/f2xme/gox/oss/adapter/cosadapter"

	storage := cosadapter.New(
		cosadapter.WithRegion("ap-guangzhou"),
		cosadapter.WithSecretID("your-secret-id"),
		cosadapter.WithSecretKey("your-secret-key"),
		cosadapter.WithBucket("my-bucket"),
	)

## AWS S3

	import "github.com/f2xme/gox/oss/adapter/s3adapter"

	storage := s3adapter.New(
		s3adapter.WithRegion("us-east-1"),
		s3adapter.WithAccessKey("your-access-key"),
		s3adapter.WithSecretKey("your-secret-key"),
		s3adapter.WithBucket("my-bucket"),
	)

## MinIO

	import "github.com/f2xme/gox/oss/adapter/minioadapter"

	storage := minioadapter.New(
		minioadapter.WithEndpoint("localhost:9000"),
		minioadapter.WithAccessKey("minioadmin"),
		minioadapter.WithSecretKey("minioadmin"),
		minioadapter.WithBucket("my-bucket"),
		minioadapter.WithSSL(false),
	)

# 最佳实践

## 1. 使用 context 控制超时

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := storage.Put(ctx, key, reader)

## 2. 设置正确的 Content-Type

	// 根据文件扩展名设置
	contentType := mime.TypeByExtension(filepath.Ext(filename))
	storage.Put(ctx, key, reader, oss.WithContentType(contentType))

## 3. 使用预签名 URL 实现直传

	// 后端生成上传 URL
	url, _ := storage.PresignedURL(ctx, key,
		oss.WithExpires(30*time.Minute),
		oss.WithMethod("PUT"),
	)

	// 前端直接上传到 OSS
	// PUT url (file data)

## 4. 合理组织文件路径

	// 推荐：按日期和类型组织
	key := fmt.Sprintf("uploads/%s/%s/%s",
		time.Now().Format("2006/01/02"),
		fileType,
		filename,
	)

	// 不推荐：所有文件放在根目录
	key := filename

## 5. 处理大文件上传

	// 使用分片上传（适配器内部实现）
	file, _ := os.Open("large-file.zip")
	defer file.Close()

	err := storage.Put(ctx, "files/large-file.zip", file,
		oss.WithContentType("application/zip"),
	)

## 6. 设置合理的 ACL

	// 公开读
	storage.Put(ctx, key, reader, oss.WithACL("public-read"))

	// 私有
	storage.Put(ctx, key, reader, oss.WithACL("private"))

## 7. 使用元数据

	storage.Put(ctx, key, reader,
		oss.WithMetadata(map[string]string{
			"user-id":    "123",
			"upload-by":  "admin",
			"created-at": time.Now().Format(time.RFC3339),
		}),
	)

# 错误处理

	err := storage.Get(ctx, key)
	if err != nil {
		if oss.IsNotFound(err) {
			// 文件不存在
		} else if oss.IsPermissionDenied(err) {
			// 权限不足
		} else {
			// 其他错误
		}
	}

# 性能考虑

  - 使用预签名 URL 实现客户端直传，减轻服务器压力
  - 大文件自动使用分片上传
  - 合理设置 context 超时时间
  - 使用 CDN 加速文件访问

# 线程安全

所有对象存储实现都应该是线程安全的，可以在多个 goroutine 中并发使用。
*/
package oss
