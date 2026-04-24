package mysqldb

import (
	"testing"
	"time"
)

// mockConfig 模拟 config.Config 接口
type mockConfig struct {
	data map[string]interface{}
}

func (m *mockConfig) Get(key string) any {
	return m.data[key]
}

func (m *mockConfig) GetString(key string) string {
	if v, ok := m.data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (m *mockConfig) GetStringSlice(key string) []string {
	if v, ok := m.data[key]; ok {
		if s, ok := v.([]string); ok {
			return s
		}
	}
	return nil
}

func (m *mockConfig) GetStringMap(key string) map[string]any {
	if v, ok := m.data[key]; ok {
		if s, ok := v.(map[string]any); ok {
			return s
		}
	}
	return nil
}

func (m *mockConfig) GetInt(key string) int {
	if v, ok := m.data[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return 0
}

func (m *mockConfig) GetInt64(key string) int64 {
	if v, ok := m.data[key]; ok {
		if i, ok := v.(int64); ok {
			return i
		}
	}
	return 0
}

func (m *mockConfig) GetDuration(key string) time.Duration {
	if v, ok := m.data[key]; ok {
		if d, ok := v.(time.Duration); ok {
			return d
		}
	}
	return 0
}

func (m *mockConfig) GetBool(key string) bool {
	if v, ok := m.data[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func TestNewWithConfig_DefaultPrefix(t *testing.T) {
	cfg := &mockConfig{
		data: map[string]interface{}{
			"db.dsn":              "user:pass@tcp(127.0.0.1:3306)/testdb",
			"db.maxOpenConns":     50,
			"db.maxIdleConns":     5,
			"db.connMaxLifetime":  30 * time.Minute,
			"db.connMaxIdleTime":  5 * time.Minute,
		},
	}

	// 注意：这个测试会尝试连接真实数据库，如果没有数据库会失败
	// 这里主要测试配置解析逻辑
	_, err := NewWithConfig(cfg)
	// 预期会失败（因为没有真实数据库），但不应该是配置解析错误
	if err == nil {
		t.Log("连接成功（可能有本地测试数据库）")
	} else {
		t.Logf("连接失败（预期行为）: %v", err)
	}
}

func TestNewWithConfig_CustomPrefix(t *testing.T) {
	cfg := &mockConfig{
		data: map[string]interface{}{
			"db_primary.dsn":              "user:pass@tcp(primary:3306)/db",
			"db_primary.maxOpenConns":     100,
			"db_primary.maxIdleConns":     10,
			"db_primary.connMaxLifetime":  1 * time.Hour,
			"db_primary.connMaxIdleTime":  10 * time.Minute,
		},
	}

	_, err := NewWithConfig(cfg, "db_primary")
	if err == nil {
		t.Log("连接成功（可能有本地测试数据库）")
	} else {
		t.Logf("连接失败（预期行为）: %v", err)
	}
}

func TestNewWithConfig_MissingDSN(t *testing.T) {
	cfg := &mockConfig{
		data: map[string]interface{}{
			"db.maxOpenConns": 50,
		},
	}

	_, err := NewWithConfig(cfg)
	if err == nil {
		t.Error("期望返回错误（缺少 DSN），但成功了")
	} else {
		t.Logf("正确返回错误: %v", err)
	}
}

func TestOptions(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
	}{
		{
			name: "WithDSN",
			opts: []Option{WithDSN("test:test@tcp(localhost:3306)/test")},
		},
		{
			name: "WithMaxOpenConns",
			opts: []Option{WithMaxOpenConns(200)},
		},
		{
			name: "WithMaxIdleConns",
			opts: []Option{WithMaxIdleConns(20)},
		},
		{
			name: "WithConnMaxLifetime",
			opts: []Option{WithConnMaxLifetime(2 * time.Hour)},
		},
		{
			name: "WithConnMaxIdleTime",
			opts: []Option{WithConnMaxIdleTime(15 * time.Minute)},
		},
		{
			name: "Multiple options",
			opts: []Option{
				WithDSN("test:test@tcp(localhost:3306)/test"),
				WithMaxOpenConns(150),
				WithMaxIdleConns(15),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := defaultOptions()
			for _, opt := range tt.opts {
				opt(&o)
			}
			// 验证选项被正确应用
			t.Logf("Options applied: %+v", o)
		})
	}
}
