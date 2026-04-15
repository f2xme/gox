package alioss

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/f2xme/gox/oss"
)

// 注意：这些是集成测试，需要真实的阿里云 OSS 凭证
// 运行时使用 go test -short 跳过集成测试

func TestNew(t *testing.T) {
	storage, err := New(
		"oss-cn-hangzhou.aliyuncs.com",
		"your-access-key-id",
		"your-access-key-secret",
		"test-bucket",
	)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	if storage == nil {
		t.Fatal("storage should not be nil")
	}
}

func TestStorage_Put(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	storage, err := New(
		"oss-cn-hangzhou.aliyuncs.com",
		"your-access-key-id",
		"your-access-key-secret",
		"test-bucket",
	)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	ctx := context.Background()
	key := "test/hello.txt"
	content := "Hello, OSS!"

	// 上传对象
	err = storage.Put(ctx, key, strings.NewReader(content))
	if err != nil {
		t.Fatalf("failed to put object: %v", err)
	}

	// 清理
	defer storage.Delete(ctx, key)
}

func TestStorage_PutWithOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	storage, err := New(
		"oss-cn-hangzhou.aliyuncs.com",
		"your-access-key-id",
		"your-access-key-secret",
		"test-bucket",
	)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	ctx := context.Background()
	key := "test/hello.txt"
	content := "Hello, OSS!"

	// 上传对象，带选项
	err = storage.Put(ctx, key, strings.NewReader(content),
		oss.WithContentType("text/plain"),
		oss.WithMetadata(map[string]string{
			"author":  "test",
			"version": "1.0",
		}),
	)
	if err != nil {
		t.Fatalf("failed to put object: %v", err)
	}

	// 清理
	defer storage.Delete(ctx, key)
}

func TestStorage_Get(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	storage, err := New(
		"oss-cn-hangzhou.aliyuncs.com",
		"your-access-key-id",
		"your-access-key-secret",
		"test-bucket",
	)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	ctx := context.Background()
	key := "test/hello.txt"
	content := "Hello, OSS!"

	// 上传对象
	err = storage.Put(ctx, key, strings.NewReader(content))
	if err != nil {
		t.Fatalf("failed to put object: %v", err)
	}
	defer storage.Delete(ctx, key)

	// 下载对象
	reader, err := storage.Get(ctx, key)
	if err != nil {
		t.Fatalf("failed to get object: %v", err)
	}
	defer reader.Close()

	// 读取内容
	buf := new(strings.Builder)
	_, err = io.Copy(buf, reader)
	if err != nil {
		t.Fatalf("failed to read content: %v", err)
	}

	if buf.String() != content {
		t.Errorf("content = %q, want %q", buf.String(), content)
	}
}

func TestStorage_Exists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	storage, err := New(
		"oss-cn-hangzhou.aliyuncs.com",
		"your-access-key-id",
		"your-access-key-secret",
		"test-bucket",
	)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	ctx := context.Background()
	key := "test/hello.txt"
	content := "Hello, OSS!"

	// 上传对象
	err = storage.Put(ctx, key, strings.NewReader(content))
	if err != nil {
		t.Fatalf("failed to put object: %v", err)
	}
	defer storage.Delete(ctx, key)

	// 检查对象存在
	exists, err := storage.Exists(ctx, key)
	if err != nil {
		t.Fatalf("failed to check existence: %v", err)
	}
	if !exists {
		t.Error("object should exist")
	}

	// 检查不存在的对象
	exists, err = storage.Exists(ctx, "test/nonexistent.txt")
	if err != nil {
		t.Fatalf("failed to check existence: %v", err)
	}
	if exists {
		t.Error("object should not exist")
	}
}

func TestStorage_Head(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	storage, err := New(
		"oss-cn-hangzhou.aliyuncs.com",
		"your-access-key-id",
		"your-access-key-secret",
		"test-bucket",
	)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	ctx := context.Background()
	key := "test/hello.txt"
	content := "Hello, OSS!"

	// 上传对象
	err = storage.Put(ctx, key, strings.NewReader(content),
		oss.WithMetadata(map[string]string{
			"author": "test",
		}),
	)
	if err != nil {
		t.Fatalf("failed to put object: %v", err)
	}
	defer storage.Delete(ctx, key)

	// 获取对象元信息
	info, err := storage.Head(ctx, key)
	if err != nil {
		t.Fatalf("failed to head object: %v", err)
	}

	if info.Key != key {
		t.Errorf("Key = %q, want %q", info.Key, key)
	}
	if info.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", info.Size, len(content))
	}
	if info.Metadata["author"] != "test" {
		t.Errorf("Metadata[author] = %q, want %q", info.Metadata["author"], "test")
	}
}

func TestStorage_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	storage, err := New(
		"oss-cn-hangzhou.aliyuncs.com",
		"your-access-key-id",
		"your-access-key-secret",
		"test-bucket",
	)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	ctx := context.Background()
	key := "test/hello.txt"
	content := "Hello, OSS!"

	// 上传对象
	err = storage.Put(ctx, key, strings.NewReader(content))
	if err != nil {
		t.Fatalf("failed to put object: %v", err)
	}

	// 删除对象
	err = storage.Delete(ctx, key)
	if err != nil {
		t.Fatalf("failed to delete object: %v", err)
	}

	// 验证对象不存在
	exists, err := storage.Exists(ctx, key)
	if err != nil {
		t.Fatalf("failed to check existence: %v", err)
	}
	if exists {
		t.Error("object should not exist after deletion")
	}
}

func TestStorage_List(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	storage, err := New(
		"oss-cn-hangzhou.aliyuncs.com",
		"your-access-key-id",
		"your-access-key-secret",
		"test-bucket",
	)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	ctx := context.Background()
	prefix := "test/list/"

	// 上传多个对象
	keys := []string{
		prefix + "file1.txt",
		prefix + "file2.txt",
		prefix + "file3.txt",
	}
	for _, key := range keys {
		err = storage.Put(ctx, key, strings.NewReader("test"))
		if err != nil {
			t.Fatalf("failed to put object: %v", err)
		}
		defer storage.Delete(ctx, key)
	}

	// 列出对象
	objects, err := storage.List(ctx,
		oss.WithPrefix(prefix),
		oss.WithMaxKeys(10),
	)
	if err != nil {
		t.Fatalf("failed to list objects: %v", err)
	}

	if len(objects) < len(keys) {
		t.Errorf("got %d objects, want at least %d", len(objects), len(keys))
	}
}

func TestStorage_PresignedURL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	storage, err := New(
		"oss-cn-hangzhou.aliyuncs.com",
		"your-access-key-id",
		"your-access-key-secret",
		"test-bucket",
	)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	ctx := context.Background()
	key := "test/hello.txt"

	// 生成 GET 预签名 URL
	url, err := storage.PresignedURL(ctx, key,
		oss.WithMethod("GET"),
		oss.WithExpires(15*time.Minute),
	)
	if err != nil {
		t.Fatalf("failed to generate presigned URL: %v", err)
	}
	if url == "" {
		t.Error("presigned URL should not be empty")
	}

	// 生成 PUT 预签名 URL
	url, err = storage.PresignedURL(ctx, key,
		oss.WithMethod("PUT"),
		oss.WithExpires(15*time.Minute),
	)
	if err != nil {
		t.Fatalf("failed to generate presigned URL: %v", err)
	}
	if url == "" {
		t.Error("presigned URL should not be empty")
	}
}
