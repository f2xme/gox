package httpadapter_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/f2xme/gox/config"
	httpadapter "github.com/f2xme/gox/config/adapter/http"
)

func TestNew_YAMLGetters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write([]byte(`
server:
  port: 8080
  host: localhost
  timeout: 30s
log:
  outputs:
    - stdout
    - file
database:
  enabled: true
`))
	}))
	defer server.Close()

	cfg, err := httpadapter.New(server.URL, httpadapter.WithFormat(httpadapter.YAML))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if got := cfg.GetInt("server.port"); got != 8080 {
		t.Errorf("GetInt('server.port') = %d, want 8080", got)
	}
	if got := cfg.GetString("server.host"); got != "localhost" {
		t.Errorf("GetString('server.host') = %q, want localhost", got)
	}
	if got := cfg.GetDuration("server.timeout"); got != 30*time.Second {
		t.Errorf("GetDuration('server.timeout') = %v, want 30s", got)
	}
	if got := cfg.GetStringSlice("log.outputs"); !reflect.DeepEqual(got, []string{"stdout", "file"}) {
		t.Errorf("GetStringSlice('log.outputs') = %#v, want stdout/file", got)
	}
	if got := cfg.GetBool("database.enabled"); !got {
		t.Error("GetBool('database.enabled') = false, want true")
	}
}

func TestNew_JSONGetters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"server":{"port":8080},"debug":true}`))
	}))
	defer server.Close()

	cfg, err := httpadapter.New(server.URL)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if got := cfg.GetInt("server.port"); got != 8080 {
		t.Errorf("GetInt('server.port') = %d, want 8080", got)
	}
	if got := cfg.GetBool("debug"); !got {
		t.Error("GetBool('debug') = false, want true")
	}
}

func TestGetReturnsIndependentAggregateValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"server":{"hosts":["one","two"]}}`))
	}))
	defer server.Close()

	cfg, err := httpadapter.New(server.URL)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	serverValue := cfg.Get("server").(map[string]any)
	hosts := serverValue["hosts"].([]any)
	serverValue["name"] = "mutated"
	hosts[0] = "mutated"

	got := cfg.Get("server").(map[string]any)
	if _, ok := got["name"]; ok {
		t.Fatalf("Get() returned mutable internal map: %#v", got)
	}
	if gotHosts := got["hosts"].([]any); gotHosts[0] != "one" {
		t.Fatalf("Get() returned mutable internal slice: %#v", gotHosts)
	}
}

func TestWithDefaultsCopiesTypedAggregates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	labels := map[string]string{"env": "prod"}
	ports := []int{8080, 8081}
	nested := map[string][]int{"ports": ports}
	cfg, err := httpadapter.New(server.URL, httpadapter.WithDefaults(map[string]any{
		"labels": labels,
		"ports":  ports,
		"nested": nested,
	}))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	labels["env"] = "mutated"
	ports[0] = 9000
	nested["ports"] = []int{9001}
	if got := cfg.Get("labels").(map[string]string)["env"]; got != "prod" {
		t.Fatalf("labels after input mutation = %q, want prod", got)
	}
	if got := cfg.Get("ports").([]int)[0]; got != 8080 {
		t.Fatalf("ports after input mutation = %d, want 8080", got)
	}
	if got := cfg.Get("nested").(map[string][]int)["ports"][0]; got != 8080 {
		t.Fatalf("nested ports after input mutation = %d, want 8080", got)
	}

	returnedLabels := cfg.Get("labels").(map[string]string)
	returnedPorts := cfg.Get("ports").([]int)
	returnedNested := cfg.Get("nested").(map[string][]int)
	returnedLabels["env"] = "changed"
	returnedPorts[0] = 9002
	returnedNested["ports"][0] = 9003
	if got := cfg.Get("labels").(map[string]string)["env"]; got != "prod" {
		t.Fatalf("labels after Get mutation = %q, want prod", got)
	}
	if got := cfg.Get("ports").([]int)[0]; got != 8080 {
		t.Fatalf("ports after Get mutation = %d, want 8080", got)
	}
	if got := cfg.Get("nested").(map[string][]int)["ports"][0]; got != 8080 {
		t.Fatalf("nested ports after Get mutation = %d, want 8080", got)
	}
}

func TestNew_WithDefaults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`server:
  port: 3000
`))
	}))
	defer server.Close()

	cfg, err := httpadapter.New(
		server.URL,
		httpadapter.WithDefaults(map[string]any{
			"app.name":       "test-app",
			"server.port":    8080,
			"server.timeout": 5 * time.Second,
		}),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if got := cfg.GetString("app.name"); got != "test-app" {
		t.Errorf("GetString('app.name') = %q, want test-app", got)
	}
	if got := cfg.GetInt("server.port"); got != 3000 {
		t.Errorf("GetInt('server.port') = %d, want remote override 3000", got)
	}
	if got := cfg.GetDuration("server.timeout"); got != 5*time.Second {
		t.Errorf("GetDuration('server.timeout') = %v, want 5s", got)
	}
}

func TestNew_WithHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Token"); got != "secret" {
			t.Errorf("X-Token = %q, want secret", got)
		}
		_, _ = w.Write([]byte(`key: value`))
	}))
	defer server.Close()

	cfg, err := httpadapter.New(server.URL, httpadapter.WithHeader("X-Token", "secret"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if got := cfg.GetString("key"); got != "value" {
		t.Errorf("GetString('key') = %q, want value", got)
	}
}

func TestNew_FailOnLoadErrorFalseUsesDefaults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	cfg, err := httpadapter.New(
		server.URL,
		httpadapter.WithFailOnLoadError(false),
		httpadapter.WithDefaults(map[string]any{"server.port": 8080}),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if got := cfg.GetInt("server.port"); got != 8080 {
		t.Errorf("GetInt('server.port') = %d, want default 8080", got)
	}
}

func TestNew_LoadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	if _, err := httpadapter.New(server.URL); err == nil {
		t.Fatal("New() error = nil, want error")
	}
}

func TestNew_RejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"value":"0123456789"}`))
	}))
	defer server.Close()

	if _, err := httpadapter.New(server.URL, httpadapter.WithFormat(httpadapter.JSON), httpadapter.WithMaxBodyBytes(16)); err == nil {
		t.Fatal("New() error = nil, want oversized response error")
	}
}

func TestNew_RejectsTrailingJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"first":1}{"second":2}`))
	}))
	defer server.Close()

	if _, err := httpadapter.New(server.URL); err == nil {
		t.Fatal("New() error = nil, want trailing JSON error")
	}
}

func TestNew_ImplementsInterfaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`key: value`))
	}))
	defer server.Close()

	cfg, err := httpadapter.New(server.URL)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	var _ config.Config = cfg
	if _, ok := cfg.(config.Watcher); !ok {
		t.Fatal("expected config.Watcher interface")
	}
}

func TestNew_WatchCallsCallback(t *testing.T) {
	var version atomic.Int64
	version.Store(1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("server:\n  port: " + strconvFormatInt(version.Load()) + "\n"))
	}))
	defer server.Close()

	cfg, err := httpadapter.New(server.URL, httpadapter.WithWatch(10*time.Millisecond))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer cfg.(interface{ Close() error }).Close()

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

	version.Store(2)

	select {
	case <-called:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("watch callback not called within 500ms")
	}

	if got := cfg.GetInt("server.port"); got != 2 {
		t.Errorf("GetInt('server.port') = %d, want 2", got)
	}
}

func TestNew_WatchSerializesSlowCallback(t *testing.T) {
	var version atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := version.Add(1)
		_, _ = w.Write([]byte("version: " + strconvFormatInt(current) + "\n"))
	}))
	defer server.Close()

	cfg, err := httpadapter.New(server.URL, httpadapter.WithWatch(5*time.Millisecond))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer cfg.(interface{ Close() error }).Close()

	var active, maxActive, calls atomic.Int64
	called := make(chan struct{}, 1)
	err = cfg.(config.Watcher).Watch(func() {
		current := active.Add(1)
		for {
			previous := maxActive.Load()
			if current <= previous || maxActive.CompareAndSwap(previous, current) {
				break
			}
		}
		time.Sleep(25 * time.Millisecond)
		active.Add(-1)
		if calls.Add(1) >= 2 {
			select {
			case called <- struct{}{}:
			default:
			}
		}
	})
	if err != nil {
		t.Fatalf("Watch() error = %v", err)
	}

	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("slow callback was not called twice")
	}
	if got := maxActive.Load(); got != 1 {
		t.Fatalf("maximum concurrent callbacks = %d, want 1", got)
	}
}

func TestMustNew_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNew should panic on invalid URL")
		}
	}()
	httpadapter.MustNew("")
}

func strconvFormatInt(v int64) string {
	return strconv.FormatInt(v, 10)
}
