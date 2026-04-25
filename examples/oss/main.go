package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/f2xme/gox/oss"
	"github.com/f2xme/gox/oss/adapter/aliyun"
)

func main() {
	fmt.Println("=== OSS 包使用示例 ===")
	fmt.Println()

	// 注意：需要配置真实的阿里云 OSS 凭证才能运行
	// 这里仅展示 API 用法
	exampleBasicUsage()
	exampleWithOptions()
	exampleListObjects()
	examplePresignedURL()
}

// exampleBasicUsage 展示基本用法
func exampleBasicUsage() {
	fmt.Println("## 基本用法")

	// 创建存储实例
	storage, err := aliyun.New(
		aliyun.WithEndpoint("oss-cn-hangzhou.aliyuncs.com"),
		aliyun.WithCredentials("your-access-key-id", "your-access-key-secret"),
		aliyun.WithBucket("your-bucket-name"),
	)
	if err != nil {
		fmt.Println("创建存储实例失败:", err)
		return
	}

	ctx := context.Background()

	// 上传对象
	reader := strings.NewReader("Hello, OSS!")
	err = storage.Put(ctx, "hello.txt", reader)
	if err != nil {
		fmt.Println("上传失败:", err)
	} else {
		fmt.Println("✓ 上传成功")
	}

	// 检查对象是否存在
	exists, err := storage.Exists(ctx, "hello.txt")
	if err != nil {
		fmt.Println("检查失败:", err)
	} else {
		fmt.Println("✓ 对象存在:", exists)
	}

	// 下载对象
	body, err := storage.Get(ctx, "hello.txt")
	if err != nil {
		fmt.Println("下载失败:", err)
	} else {
		body.Close()
		fmt.Println("✓ 下载成功")
	}

	// 删除对象
	err = storage.Delete(ctx, "hello.txt")
	if err != nil {
		fmt.Println("删除失败:", err)
	} else {
		fmt.Println("✓ 删除成功")
	}

	fmt.Println()
}

// exampleWithOptions 展示使用选项
func exampleWithOptions() {
	fmt.Println("## 使用选项")

	storage, _ := aliyun.New(
		aliyun.WithEndpoint("oss-cn-hangzhou.aliyuncs.com"),
		aliyun.WithCredentials("your-access-key-id", "your-access-key-secret"),
		aliyun.WithBucket("your-bucket-name"),
		aliyun.WithEnableCRC(true),
		aliyun.WithTimeout(60),
	)

	ctx := context.Background()

	// 上传时设置 Content-Type 和元数据
	reader := strings.NewReader("photo data")
	err := storage.Put(ctx, "photos/photo.jpg", reader,
		oss.WithContentType("image/jpeg"),
		oss.WithMetadata(map[string]string{
			"author":  "alice",
			"version": "1.0",
		}),
	)
	if err != nil {
		fmt.Println("上传失败:", err)
	} else {
		fmt.Println("✓ 上传成功（带选项）")
	}

	// 获取对象元信息
	info, err := storage.Stat(ctx, "photos/photo.jpg")
	if err != nil {
		fmt.Println("获取元信息失败:", err)
	} else {
		fmt.Println("✓ 文件大小:", info.Size, "bytes")
		fmt.Println("✓ Content-Type:", info.ContentType)
	}

	fmt.Println()
}

// exampleListObjects 展示列出对象
func exampleListObjects() {
	fmt.Println("## 列出对象")

	storage, _ := aliyun.New(
		aliyun.WithEndpoint("oss-cn-hangzhou.aliyuncs.com"),
		aliyun.WithCredentials("your-access-key-id", "your-access-key-secret"),
		aliyun.WithBucket("your-bucket-name"),
	)

	ctx := context.Background()

	// 列出指定前缀的对象
	result, err := storage.List(ctx,
		oss.WithPrefix("photos/"),
		oss.WithLimit(10),
	)
	if err != nil {
		fmt.Println("列出对象失败:", err)
	} else {
		fmt.Println("✓ 找到对象数量:", len(result.Objects))
		for _, obj := range result.Objects {
			fmt.Println("  -", obj.Key, "(", obj.Size, "bytes)")
		}
	}

	fmt.Println()
}

// examplePresignedURL 展示预签名 URL
func examplePresignedURL() {
	fmt.Println("## 预签名 URL")

	storage, _ := aliyun.New(
		aliyun.WithEndpoint("oss-cn-hangzhou.aliyuncs.com"),
		aliyun.WithCredentials("your-access-key-id", "your-access-key-secret"),
		aliyun.WithBucket("your-bucket-name"),
	)

	ctx := context.Background()

	// 生成下载 URL（有效期 15 分钟）
	url, err := storage.SignURL(ctx, "photos/photo.jpg",
		oss.WithMethod(oss.MethodGet),
		oss.WithExpires(15*time.Minute),
	)
	if err != nil {
		fmt.Println("生成 URL 失败:", err)
	} else {
		fmt.Println("✓ 下载 URL:", url)
	}

	// 生成上传 URL（有效期 30 分钟）
	url, err = storage.SignURL(ctx, "uploads/new-file.jpg",
		oss.WithMethod(oss.MethodPut),
		oss.WithExpires(30*time.Minute),
	)
	if err != nil {
		fmt.Println("生成 URL 失败:", err)
	} else {
		fmt.Println("✓ 上传 URL:", url)
	}

	fmt.Println()
}
