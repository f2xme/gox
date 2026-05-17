package env_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/f2xme/gox/config"
	"github.com/f2xme/gox/config/adapter/env"
)

func TestNew_GettersFromEnvironment(t *testing.T) {
	t.Setenv("GOX_ENV_TEST_GETTERS_SERVER_PORT", "3000")
	t.Setenv("GOX_ENV_TEST_GETTERS_SERVER_HOST", "localhost")
	t.Setenv("GOX_ENV_TEST_GETTERS_LOG_OUTPUTS", "stdout, file")
	t.Setenv("GOX_ENV_TEST_GETTERS_DATABASE_ENABLED", "true")
	t.Setenv("GOX_ENV_TEST_GETTERS_DATABASE_TIMEOUT", "30s")

	cfg := env.New(env.WithPrefix("GOX_ENV_TEST_GETTERS"))

	if got := cfg.GetInt("server.port"); got != 3000 {
		t.Errorf("GetInt('server.port') = %d, want 3000", got)
	}
	if got := cfg.GetString("server.host"); got != "localhost" {
		t.Errorf("GetString('server.host') = %q, want localhost", got)
	}
	if got := cfg.GetStringSlice("log.outputs"); !reflect.DeepEqual(got, []string{"stdout", "file"}) {
		t.Errorf("GetStringSlice('log.outputs') = %#v, want stdout/file", got)
	}
	if got := cfg.GetBool("database.enabled"); !got {
		t.Error("GetBool('database.enabled') = false, want true")
	}
	if got := cfg.GetDuration("database.timeout"); got != 30*time.Second {
		t.Errorf("GetDuration('database.timeout') = %v, want 30s", got)
	}
}

func TestNew_WithDefaults(t *testing.T) {
	t.Setenv("GOX_ENV_TEST_DEFAULTS_SERVER_PORT", "3000")

	cfg := env.New(
		env.WithPrefix("GOX_ENV_TEST_DEFAULTS"),
		env.WithDefaults(map[string]any{
			"app.name":        "test-app",
			"server.port":     8080,
			"server.timeout":  5 * time.Second,
			"server.features": []string{"http", "grpc"},
		}),
	)

	if got := cfg.GetString("app.name"); got != "test-app" {
		t.Errorf("GetString('app.name') = %q, want test-app", got)
	}
	if got := cfg.GetInt("server.port"); got != 3000 {
		t.Errorf("GetInt('server.port') = %d, want env override 3000", got)
	}
	if got := cfg.GetDuration("server.timeout"); got != 5*time.Second {
		t.Errorf("GetDuration('server.timeout') = %v, want 5s", got)
	}
	if got := cfg.GetStringSlice("server.features"); !reflect.DeepEqual(got, []string{"http", "grpc"}) {
		t.Errorf("GetStringSlice('server.features') = %#v, want defaults", got)
	}
}

func TestNew_KeyNormalization(t *testing.T) {
	t.Setenv("GOX_ENV_TEST_NORMALIZE_DATABASE_MAX_CONNECTIONS", "100")

	cfg := env.New(env.WithPrefix("gox_env_test_normalize_"))

	if got := cfg.GetInt("database.max-connections"); got != 100 {
		t.Errorf("GetInt('database.max-connections') = %d, want 100", got)
	}
}

func TestNew_GetStringMap(t *testing.T) {
	t.Setenv("GOX_ENV_TEST_MAP_SERVER_HOST", "localhost")
	t.Setenv("GOX_ENV_TEST_MAP_SERVER_PORT", "8080")
	t.Setenv("GOX_ENV_TEST_MAP_SERVER_MAX_CONNECTIONS", "100")

	cfg := env.New(
		env.WithPrefix("GOX_ENV_TEST_MAP"),
		env.WithDefaults(map[string]any{
			"server": map[string]any{
				"scheme": "http",
			},
		}),
	)

	got := cfg.GetStringMap("server")
	want := map[string]any{
		"host":            "localhost",
		"max_connections": "100",
		"port":            "8080",
		"scheme":          "http",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetStringMap('server') = %#v, want %#v", got, want)
	}
}

func TestNew_MissingKey_ZeroValue(t *testing.T) {
	cfg := env.New(env.WithPrefix("GOX_ENV_TEST_MISSING"))

	if got := cfg.GetString("missing"); got != "" {
		t.Errorf("GetString('missing') = %q, want empty", got)
	}
	if got := cfg.GetInt("missing"); got != 0 {
		t.Errorf("GetInt('missing') = %d, want 0", got)
	}
	if got := cfg.GetBool("missing"); got {
		t.Error("GetBool('missing') = true, want false")
	}
	if got := cfg.GetDuration("missing"); got != 0 {
		t.Errorf("GetDuration('missing') = %v, want 0", got)
	}
	if got := cfg.GetStringMap("missing"); len(got) != 0 {
		t.Errorf("GetStringMap('missing') = %#v, want empty", got)
	}
}

func TestNew_ImplementsConfigInterface(t *testing.T) {
	cfg := env.New()
	var _ config.Config = cfg
}

func TestNew_ImplementsWatcher(t *testing.T) {
	cfg := env.New()
	w, ok := cfg.(config.Watcher)
	if !ok {
		t.Fatal("expected config.Watcher interface")
	}
	if err := w.Watch(func() {}); err != nil {
		t.Errorf("Watch returned error: %v", err)
	}
}

func TestNew_WatchCallsCallback(t *testing.T) {
	t.Setenv("GOX_ENV_TEST_WATCH_SERVER_PORT", "8080")

	cfg := env.New(
		env.WithPrefix("GOX_ENV_TEST_WATCH"),
		env.WithWatch(10*time.Millisecond),
	)

	called := make(chan struct{}, 1)
	w := cfg.(config.Watcher)
	if err := w.Watch(func() {
		select {
		case called <- struct{}{}:
		default:
		}
	}); err != nil {
		t.Fatalf("Watch returned error: %v", err)
	}

	time.Sleep(30 * time.Millisecond)
	t.Setenv("GOX_ENV_TEST_WATCH_SERVER_PORT", "9090")

	select {
	case <-called:
	case <-time.After(500 * time.Millisecond):
		t.Error("watch callback not called within 500ms")
	}
}
