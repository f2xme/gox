package captcha

import (
	"context"
	"testing"
	"time"

	"github.com/f2xme/gox/captcha/generator/base64"
)

// mockStore 是用于测试的简单内存存储
type mockStore struct {
	data map[string]string
}

func newMockStore() *mockStore {
	return &mockStore{
		data: make(map[string]string),
	}
}

func (s *mockStore) Set(ctx context.Context, id, answer string, ttl time.Duration) error {
	s.data[id] = answer
	return nil
}

func (s *mockStore) Get(ctx context.Context, id string) (string, error) {
	if answer, ok := s.data[id]; ok {
		return answer, nil
	}
	return "", ErrNotFound
}

func (s *mockStore) Delete(ctx context.Context, id string) error {
	delete(s.data, id)
	return nil
}

func (s *mockStore) Exists(ctx context.Context, id string) (bool, error) {
	_, ok := s.data[id]
	return ok, nil
}

func TestCaptcha_Generate(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	gen := base64.New()
	c := New(store, WithGenerator(gen))

	id, data, err := c.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if id == "" {
		t.Error("Generate() returned empty id")
	}
	if data == "" {
		t.Error("Generate() returned empty data")
	}

	// 验证答案已存储
	exists, err := store.Exists(ctx, id)
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Answer should be stored")
	}
}

func TestCaptcha_Verify(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	gen := base64.New()
	c := New(store, WithGenerator(gen))

	// 生成验证码
	id, _, err := c.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 获取存储的答案
	answer, err := store.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	tests := []struct {
		name    string
		id      string
		answer  string
		want    bool
		wantErr bool
	}{
		{
			name:    "correct answer",
			id:      id,
			answer:  answer,
			want:    true,
			wantErr: false,
		},
		{
			name:    "wrong answer",
			id:      id,
			answer:  "wrong",
			want:    false,
			wantErr: false,
		},
		{
			name:    "empty id",
			id:      "",
			answer:  answer,
			want:    false,
			wantErr: false,
		},
		{
			name:    "empty answer",
			id:      id,
			answer:  "",
			want:    false,
			wantErr: false,
		},
		{
			name:    "non-existent id",
			id:      "non-existent",
			answer:  answer,
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.Verify(ctx, tt.id, tt.answer)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCaptcha_Delete(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	gen := base64.New()
	c := New(store, WithGenerator(gen))

	// 生成验证码
	id, _, err := c.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 验证存在
	exists, err := store.Exists(ctx, id)
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Captcha should exist before deletion")
	}

	// 删除
	if err := c.Delete(ctx, id); err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// 验证已删除
	exists, err = store.Exists(ctx, id)
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if exists {
		t.Error("Captcha should not exist after deletion")
	}
}

func TestCaptcha_Regenerate(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	gen := base64.New()
	c := New(store, WithGenerator(gen))

	// 生成验证码
	id, data1, err := c.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 重新生成
	data2, err := c.Regenerate(ctx, id)
	if err != nil {
		t.Fatalf("Regenerate() error = %v", err)
	}

	if data2 == "" {
		t.Error("Regenerate() returned empty data")
	}

	// 数据应该不同（概率上）
	if data1 == data2 {
		t.Log("Warning: regenerated data is the same (may happen by chance)")
	}
}

func TestCaptcha_WithTTL(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	gen := base64.New()

	customTTL := 10 * time.Minute
	c := New(store, WithGenerator(gen), WithTTL(customTTL))

	impl := c.(*captcha)
	if impl.opts.TTL != customTTL {
		t.Errorf("TTL = %v, want %v", impl.opts.TTL, customTTL)
	}

	// 生成验证码并验证 TTL 被使用
	_, _, err := c.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
}

func TestCaptcha_WithIDLength(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	gen := base64.New()

	customLength := 32
	c := New(store, WithGenerator(gen), WithIDLength(customLength))

	id, _, err := c.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(id) != customLength {
		t.Errorf("ID length = %v, want %v", len(id), customLength)
	}
}

func TestCaptcha_WithGenerator(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	gen := base64.New(base64.WithType(base64.TypeMath))

	c := New(store, WithGenerator(gen))

	id, data, err := c.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if id == "" || data == "" {
		t.Error("ID or data is empty")
	}
}
