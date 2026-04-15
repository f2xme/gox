package email

import (
	"testing"
)

// TestNew 测试 New 函数的各种输入场景
func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     int
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "empty host",
			host:     "",
			port:     587,
			username: "test@example.com",
			password: "password",
			wantErr:  true,
		},
		{
			name:     "invalid port",
			host:     "smtp.example.com",
			port:     0,
			username: "test@example.com",
			password: "password",
			wantErr:  true,
		},
		{
			name:     "empty username",
			host:     "smtp.example.com",
			port:     587,
			username: "",
			password: "password",
			wantErr:  true,
		},
		{
			name:     "empty password",
			host:     "smtp.example.com",
			port:     587,
			username: "test@example.com",
			password: "",
			wantErr:  true,
		},
		{
			name:     "valid config",
			host:     "smtp.example.com",
			port:     587,
			username: "test@example.com",
			password: "password",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.host, tt.port, tt.username, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestNewWithOptions 测试 NewWithOptions 函数的配置选项
func TestNewWithOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name: "valid options",
			opts: []Option{
				WithHost("smtp.example.com"),
				WithPort(587),
				WithUsername("test@example.com"),
				WithPassword("password"),
				WithName("Test Sender"),
				WithSSL(false),
			},
			wantErr: false,
		},
		{
			name: "missing host",
			opts: []Option{
				WithPort(587),
				WithUsername("test@example.com"),
				WithPassword("password"),
			},
			wantErr: true,
		},
		{
			name: "missing username",
			opts: []Option{
				WithHost("smtp.example.com"),
				WithPort(587),
				WithPassword("password"),
			},
			wantErr: true,
		},
		{
			name: "missing password",
			opts: []Option{
				WithHost("smtp.example.com"),
				WithPort(587),
				WithUsername("test@example.com"),
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			opts: []Option{
				WithHost("smtp.example.com"),
				WithPort(0),
				WithUsername("test@example.com"),
				WithPassword("password"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewWithOptions(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWithOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestNewBackwardCompatibility 测试旧版本 API 的向后兼容性
func TestNewBackwardCompatibility(t *testing.T) {
	// 测试旧版本 API 仍然工作
	client, err := New("smtp.example.com", 587, "test@example.com", "password",
		WithName("Test Sender"),
		WithSSL(true),
	)
	if err != nil {
		t.Errorf("New() with old API failed: %v", err)
	}
	if client == nil {
		t.Error("New() returned nil client")
	}
	if client.name != "Test Sender" {
		t.Errorf("Expected name 'Test Sender', got '%s'", client.name)
	}
	if !client.dialer.SSL {
		t.Error("Expected SSL to be true")
	}
}
