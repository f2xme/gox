package aliyun

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/f2xme/gox/oss"
)

func TestNew(t *testing.T) {
	storage, err := New(
		WithEndpoint("oss-cn-hangzhou.aliyuncs.com"),
		WithCredentials("your-access-key-id", "your-access-key-secret"),
		WithBucket("test-bucket"),
	)
	if err != nil {
		t.Fatalf("创建存储实例失败: %v", err)
	}
	if storage == nil {
		t.Fatal("存储实例不能为空")
	}
}

func TestNewValidation(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
		code string
	}{
		{
			name: "missing endpoint",
			opts: []Option{WithCredentials("id", "secret"), WithBucket("bucket")},
			code: oss.ErrCodeInvalidArgument,
		},
		{
			name: "missing access key id",
			opts: []Option{WithEndpoint("endpoint"), WithCredentials("", "secret"), WithBucket("bucket")},
			code: oss.ErrCodeInvalidArgument,
		},
		{
			name: "missing access key secret",
			opts: []Option{WithEndpoint("endpoint"), WithCredentials("id", ""), WithBucket("bucket")},
			code: oss.ErrCodeInvalidArgument,
		},
		{
			name: "missing bucket",
			opts: []Option{WithEndpoint("endpoint"), WithCredentials("id", "secret")},
			code: oss.ErrCodeInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.opts...)
			if !oss.IsCode(err, tt.code) {
				t.Fatalf("错误码 = %q, want %q, err=%v", oss.ErrorCode(err), tt.code, err)
			}
		})
	}
}

func TestNewWithOptionsNil(t *testing.T) {
	_, err := NewWithOptions(nil)
	if !oss.IsCode(err, oss.ErrCodeInvalidArgument) {
		t.Fatalf("错误码 = %q, want %q, err=%v", oss.ErrorCode(err), oss.ErrCodeInvalidArgument, err)
	}
}

func TestNewWithConfigValidation(t *testing.T) {
	cfg := mapConfig{}
	_, err := NewWithConfig(cfg)
	if !oss.IsCode(err, oss.ErrCodeInvalidArgument) {
		t.Fatalf("错误码 = %q, want %q, err=%v", oss.ErrorCode(err), oss.ErrCodeInvalidArgument, err)
	}
}

func TestNewWithConfigPrefix(t *testing.T) {
	cfg := mapConfig{
		"storage.aliyun.endpoint":        "oss-cn-hangzhou.aliyuncs.com",
		"storage.aliyun.accessKeyID":     "access-key-id",
		"storage.aliyun.accessKeySecret": "access-key-secret",
		"storage.aliyun.bucket":          "test-bucket",
		"storage.aliyun.securityToken":   "security-token",
		"storage.aliyun.enableCRC":       true,
		"storage.aliyun.timeout":         int64(30),
	}

	storage, err := NewWithConfig(cfg, "storage.aliyun")
	if err != nil {
		t.Fatalf("使用配置创建存储实例失败: %v", err)
	}
	if storage == nil {
		t.Fatal("存储实例不能为空")
	}
}

func TestStorage_SignURL(t *testing.T) {
	storage, err := New(
		WithEndpoint("oss-cn-hangzhou.aliyuncs.com"),
		WithCredentials("test-access-key-id", "test-access-key-secret"),
		WithBucket("test-bucket"),
	)
	if err != nil {
		t.Fatalf("创建存储实例失败: %v", err)
	}

	url, err := storage.SignURL(context.Background(), "test/hello.txt",
		oss.WithMethod(oss.MethodPut),
		oss.WithExpires(15*time.Minute),
		oss.WithSignContentType("text/plain"),
	)
	if err != nil {
		t.Fatalf("生成预签名 URL 失败: %v", err)
	}
	if url == "" {
		t.Fatal("预签名 URL 不能为空")
	}
}

type mapConfig map[string]any

func (m mapConfig) Get(key string) any {
	return m[key]
}

func (m mapConfig) GetString(key string) string {
	value, _ := m[key].(string)
	return value
}

func (m mapConfig) GetStringSlice(key string) []string {
	value, _ := m[key].([]string)
	return value
}

func (m mapConfig) GetStringMap(key string) map[string]any {
	value, _ := m[key].(map[string]any)
	return value
}

func (m mapConfig) GetInt(key string) int {
	value, _ := m[key].(int)
	return value
}

func (m mapConfig) GetInt64(key string) int64 {
	value, _ := m[key].(int64)
	return value
}

func (m mapConfig) GetDuration(key string) time.Duration {
	value, _ := m[key].(time.Duration)
	return value
}

func (m mapConfig) GetBool(key string) bool {
	value, _ := m[key].(bool)
	return value
}

func TestStorage_Integration(t *testing.T) {
	storage := newIntegrationStorage(t)

	ctx := context.Background()
	key := "gox-tests/hello.txt"
	content := "Hello, OSS!"

	if err := storage.Put(ctx, key, strings.NewReader(content),
		oss.WithContentType("text/plain"),
		oss.WithMetadata(map[string]string{"author": "gox"}),
	); err != nil {
		t.Fatalf("上传对象失败: %v", err)
	}
	defer func() {
		if err := storage.Delete(ctx, key); err != nil {
			t.Logf("清理对象失败: %v", err)
		}
	}()

	reader, err := storage.Get(ctx, key)
	if err != nil {
		t.Fatalf("下载对象失败: %v", err)
	}
	defer reader.Close()

	var buf strings.Builder
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatalf("读取对象失败: %v", err)
	}
	if buf.String() != content {
		t.Fatalf("对象内容 = %q, want %q", buf.String(), content)
	}

	info, err := storage.Stat(ctx, key)
	if err != nil {
		t.Fatalf("获取对象元信息失败: %v", err)
	}
	if info.Size != int64(len(content)) {
		t.Fatalf("对象大小 = %d, want %d", info.Size, len(content))
	}
	if info.Metadata["author"] != "gox" {
		t.Fatalf("对象元数据 author = %q, want %q", info.Metadata["author"], "gox")
	}

	result, err := storage.List(ctx, oss.WithPrefix("gox-tests/"), oss.WithLimit(10))
	if err != nil {
		t.Fatalf("列出对象失败: %v", err)
	}
	if len(result.Objects) == 0 {
		t.Fatal("对象列表不能为空")
	}

	exists, err := storage.Exists(ctx, key)
	if err != nil {
		t.Fatalf("检查对象存在性失败: %v", err)
	}
	if !exists {
		t.Fatal("对象应该存在")
	}
}

func newIntegrationStorage(t *testing.T) *Storage {
	t.Helper()

	endpoint := os.Getenv("GOX_ALIYUN_ENDPOINT")
	accessKeyID := os.Getenv("GOX_ALIYUN_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("GOX_ALIYUN_ACCESS_KEY_SECRET")
	bucket := os.Getenv("GOX_ALIYUN_BUCKET")
	if endpoint == "" || accessKeyID == "" || accessKeySecret == "" || bucket == "" {
		t.Skip("跳过阿里云 OSS 集成测试：需要设置 GOX_ALIYUN_ENDPOINT、GOX_ALIYUN_ACCESS_KEY_ID、GOX_ALIYUN_ACCESS_KEY_SECRET、GOX_ALIYUN_BUCKET")
	}

	storage, err := New(
		WithEndpoint(endpoint),
		WithCredentials(accessKeyID, accessKeySecret),
		WithBucket(bucket),
		WithSecurityToken(os.Getenv("GOX_ALIYUN_SECURITY_TOKEN")),
	)
	if err != nil {
		t.Fatalf("创建集成测试存储实例失败: %v", err)
	}
	return storage
}
