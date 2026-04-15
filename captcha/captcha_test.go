package captcha

import (
	"testing"
	"time"

	"github.com/mojocn/base64Captcha"
)

// TestGenerate 测试验证码生成功能
// 验证不同类型验证码的生成是否正常
func TestGenerate(t *testing.T) {
	tests := []struct {
		name        string
		captchaType CaptchaType
		wantErr     bool
	}{
		{
			name:        "generate digit captcha",
			captchaType: TypeDigit,
			wantErr:     false,
		},
		{
			name:        "generate string captcha",
			captchaType: TypeString,
			wantErr:     false,
		},
		{
			name:        "generate math captcha",
			captchaType: TypeMath,
			wantErr:     false,
		},
		{
			name:        "generate audio captcha",
			captchaType: TypeAudio,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := base64Captcha.DefaultMemStore
			c := New(store, WithType(tt.captchaType))

			id, b64s, err := c.Generate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if id == "" {
					t.Error("Generate() returned empty id")
				}
				if b64s == "" {
					t.Error("Generate() returned empty base64 string")
				}
			}
		})
	}
}

// TestVerify 测试验证码验证功能
// 验证各种边界情况下的验证行为
func TestVerify(t *testing.T) {
	store := base64Captcha.DefaultMemStore
	c := New(store)

	// 首先生成一个验证码
	id, _, err := c.Generate()
	if err != nil {
		t.Fatalf("Failed to generate captcha: %v", err)
	}

	tests := []struct {
		name   string
		id     string
		answer string
		want   bool
	}{
		{
			name:   "empty id",
			id:     "",
			answer: "1234",
			want:   false,
		},
		{
			name:   "empty answer",
			id:     id,
			answer: "",
			want:   false,
		},
		{
			name:   "wrong answer",
			id:     id,
			answer: "wrong",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.Verify(tt.id, tt.answer); got != tt.want {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestOptions 测试验证码配置选项
// 验证各种配置选项是否正确应用
func TestOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantW   int
		wantH   int
		wantLen int
	}{
		{
			name:    "default options",
			opts:    nil,
			wantW:   240,
			wantH:   80,
			wantLen: 4,
		},
		{
			name:    "custom width",
			opts:    []Option{WithWidth(300)},
			wantW:   300,
			wantH:   80,
			wantLen: 4,
		},
		{
			name:    "custom height",
			opts:    []Option{WithHeight(100)},
			wantW:   240,
			wantH:   100,
			wantLen: 4,
		},
		{
			name:    "custom length",
			opts:    []Option{WithLength(6)},
			wantW:   240,
			wantH:   80,
			wantLen: 6,
		},
		{
			name:    "multiple options",
			opts:    []Option{WithWidth(320), WithHeight(120), WithLength(5)},
			wantW:   320,
			wantH:   120,
			wantLen: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := base64Captcha.DefaultMemStore
			c := New(store, tt.opts...)

			impl, ok := c.(*captchaImpl)
			if !ok {
				t.Fatal("Failed to cast to captchaImpl")
			}

			if impl.opts.Width != tt.wantW {
				t.Errorf("Width = %v, want %v", impl.opts.Width, tt.wantW)
			}
			if impl.opts.Height != tt.wantH {
				t.Errorf("Height = %v, want %v", impl.opts.Height, tt.wantH)
			}
			if impl.opts.Length != tt.wantLen {
				t.Errorf("Length = %v, want %v", impl.opts.Length, tt.wantLen)
			}
		})
	}
}

// TestMemoryStore 测试内存存储功能
// 验证验证码的存储、获取和自动删除机制
func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore(100, 5*time.Minute)
	c := New(store)

	// 立即生成并验证
	id, _, err := c.Generate()
	if err != nil {
		t.Fatalf("Failed to generate captcha: %v", err)
	}

	// 验证前先存储答案
	answer := store.Get(id, false)
	if answer == "" {
		t.Fatal("Failed to get answer from store")
	}

	// 使用正确答案验证应该成功
	if !c.Verify(id, answer) {
		t.Error("Verify() failed for correct answer")
	}

	// 验证成功后，验证码应被删除（verify 使用 clear=true）
	// 所以再次验证应该失败
	if c.Verify(id, answer) {
		t.Error("Verify() should fail after first successful verification")
	}
}
