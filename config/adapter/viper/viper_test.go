package viper_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/config/adapter/viper"
)

// TestNew_BasicGetters 测试基本的配置读取方法
func TestNew_BasicGetters(t *testing.T) {
	c, err := viper.New("testdata/test.yml")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if got := c.GetInt("server.port"); got != 8080 {
		t.Errorf("GetInt('server.port') = %d, want 8080", got)
	}
	if got := c.GetString("server.host"); got != "localhost" {
		t.Errorf("GetString('server.host') = %q, want 'localhost'", got)
	}
	if got := c.GetString("log.level"); got != "info" {
		t.Errorf("GetString('log.level') = %q, want 'info'", got)
	}
	if got := c.GetBool("database.enabled"); !got {
		t.Error("GetBool('database.enabled') = false, want true")
	}
	if got := c.GetInt("database.max_connections"); got != 100 {
		t.Errorf("GetInt('database.max_connections') = %d, want 100", got)
	}
	if got := c.GetDuration("database.timeout"); got != 30*time.Second {
		t.Errorf("GetDuration('database.timeout') = %v, want 30s", got)
	}
}

// TestNew_GetStringSlice 测试字符串切片读取
func TestNew_GetStringSlice(t *testing.T) {
	c, err := viper.New("testdata/test.yml")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	outputs := c.GetStringSlice("log.outputs")
	if len(outputs) != 2 {
		t.Fatalf("expected 2 outputs, got %d", len(outputs))
	}
	if outputs[0] != "stdout" || outputs[1] != "file" {
		t.Errorf("unexpected outputs: %v", outputs)
	}
}

// TestNew_GetStringMap 测试字符串映射读取
func TestNew_GetStringMap(t *testing.T) {
	c, err := viper.New("testdata/test.yml")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	m := c.GetStringMap("server")
	if m["host"] != "localhost" {
		t.Errorf("expected host 'localhost', got %v", m["host"])
	}
}

// TestNew_Get 测试通用 Get 方法
func TestNew_Get(t *testing.T) {
	c, err := viper.New("testdata/test.yml")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	v := c.Get("server.port")
	if v == nil {
		t.Error("Get('server.port') returned nil")
	}
}

// TestNew_MissingKey_ZeroValue 测试不存在的键返回零值
func TestNew_MissingKey_ZeroValue(t *testing.T) {
	c, err := viper.New("testdata/test.yml")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if got := c.GetString("nonexistent"); got != "" {
		t.Errorf("expected empty string for missing key, got %q", got)
	}
	if got := c.GetInt("nonexistent"); got != 0 {
		t.Errorf("expected 0 for missing key, got %d", got)
	}
	if got := c.GetBool("nonexistent"); got {
		t.Error("expected false for missing key")
	}
}

// TestNew_WithDefaults 测试默认值设置（不覆盖文件中的值）
func TestNew_WithDefaults(t *testing.T) {
	c, err := viper.New("testdata/test.yml",
		viper.WithDefaults(map[string]any{
			"app.name":    "test-app",
			"server.port": 9999,
		}),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if got := c.GetString("app.name"); got != "test-app" {
		t.Errorf("expected default 'test-app', got %q", got)
	}
	if got := c.GetInt("server.port"); got != 8080 {
		t.Errorf("expected file value 8080, got %d (default should not override)", got)
	}
}

// TestNew_WithEnvPrefix 测试环境变量覆盖
func TestNew_WithEnvPrefix(t *testing.T) {
	os.Setenv("TESTAPP_SERVER_PORT", "3000")
	defer os.Unsetenv("TESTAPP_SERVER_PORT")

	c, err := viper.New("testdata/test.yml",
		viper.WithEnvPrefix("TESTAPP"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if got := c.GetInt("server.port"); got != 3000 {
		t.Errorf("expected env override 3000, got %d", got)
	}
}

// TestNew_ImplementsWatcher 测试 Watcher 接口实现
func TestNew_ImplementsWatcher(t *testing.T) {
	c, err := viper.New("testdata/test.yml", viper.WithWatch())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	w, ok := c.(config.Watcher)
	if !ok {
		t.Fatal("expected config.Watcher interface")
	}
	if err := w.Watch(func() {}); err != nil {
		t.Errorf("Watch returned error: %v", err)
	}
}

// TestNew_WatchCallsCallback 测试配置文件变更时回调函数被调用
func TestNew_WatchCallsCallback(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "watch.yml")
	os.WriteFile(cfgFile, []byte("key: original\n"), 0644)

	c, err := viper.New(cfgFile, viper.WithWatch())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	called := make(chan bool, 1)
	w := c.(config.Watcher)
	w.Watch(func() {
		called <- true
	})

	time.Sleep(100 * time.Millisecond)
	os.WriteFile(cfgFile, []byte("key: modified\n"), 0644)

	select {
	case <-called:
		// success
	case <-time.After(2 * time.Second):
		t.Error("watch callback not called within 2s")
	}
}

// TestNew_ImplementsConfigInterface 测试 Config 接口实现
func TestNew_ImplementsConfigInterface(t *testing.T) {
	c, err := viper.New("testdata/test.yml")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	var _ config.Config = c
}

// TestNew_InvalidFile 测试无效文件路径
func TestNew_InvalidFile(t *testing.T) {
	_, err := viper.New("nonexistent.yml")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

// TestMustNew_Success 测试 MustNew 成功场景
func TestMustNew_Success(t *testing.T) {
	c := viper.MustNew("testdata/test.yml")
	if c == nil {
		t.Fatal("MustNew() returned nil")
	}
}

// TestMustNew_Panic 测试 MustNew 在无效文件时 panic
func TestMustNew_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNew should panic on invalid file")
		}
	}()
	viper.MustNew("nonexistent.yml")
}
