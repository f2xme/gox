package httpadapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/f2xme/gox/config"
)

type httpConfig struct {
	url          string
	client       *http.Client
	headers      map[string]string
	format       Format
	defaults     map[string]any
	lastBody     string
	watch        bool
	interval     time.Duration
	maxBodyBytes int64
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex
	values       map[string]any
	watchFns     []func()
	watchOnce    sync.Once
}

var (
	_ config.Config  = (*httpConfig)(nil)
	_ config.Watcher = (*httpConfig)(nil)
)

// New 创建一个由 HTTP 远端配置支持的 config.Config 实例。
func New(url string, opts ...Option) (config.Config, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	if strings.TrimSpace(url) == "" {
		return nil, fmt.Errorf("httpadapter: url is required")
	}

	client := options.client
	if client == nil {
		client = &http.Client{Timeout: options.timeout}
	} else if client.Timeout == 0 && options.timeout > 0 {
		copied := *client
		copied.Timeout = options.timeout
		client = &copied
	}

	ctx, cancel := context.WithCancel(context.Background())
	cfg := &httpConfig{
		url:          url,
		client:       client,
		headers:      cloneStringMap(options.headers),
		format:       options.format,
		defaults:     cloneMap(options.defaults),
		watch:        options.watch,
		interval:     options.watchInterval,
		maxBodyBytes: options.maxBodyBytes,
		ctx:          ctx,
		cancel:       cancel,
		values:       cloneMap(options.defaults),
	}

	if err := cfg.reload(); err != nil {
		if options.failOnLoadError {
			cancel()
			return nil, err
		}
	}

	return cfg, nil
}

// MustNew 创建一个由 HTTP 远端配置支持的 config.Config 实例，失败时 panic。
func MustNew(url string, opts ...Option) config.Config {
	cfg, err := New(url, opts...)
	if err != nil {
		panic(fmt.Errorf("httpadapter: failed to load %s: %w", url, err))
	}
	return cfg
}

// Get 返回指定配置键的原始值。
func (c *httpConfig) Get(key string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return cloneValue(getValue(c.values, key))
}

// GetString 返回指定配置键的字符串值。
func (c *httpConfig) GetString(key string) string {
	return toString(c.Get(key))
}

// GetStringSlice 返回指定配置键的字符串切片值。
func (c *httpConfig) GetStringSlice(key string) []string {
	switch val := c.Get(key).(type) {
	case nil:
		return nil
	case []string:
		return append([]string(nil), val...)
	case []any:
		items := make([]string, 0, len(val))
		for _, item := range val {
			items = append(items, toString(item))
		}
		return items
	case string:
		if val == "" {
			return nil
		}
		parts := strings.Split(val, ",")
		items := make([]string, 0, len(parts))
		for _, part := range parts {
			item := strings.TrimSpace(part)
			if item != "" {
				items = append(items, item)
			}
		}
		return items
	default:
		return []string{toString(val)}
	}
}

// GetStringMap 返回指定配置键的映射值。
func (c *httpConfig) GetStringMap(key string) map[string]any {
	val := c.Get(key)
	if m, ok := val.(map[string]any); ok {
		return cloneMap(m)
	}
	return map[string]any{}
}

// GetInt 返回指定配置键的 int 值。
func (c *httpConfig) GetInt(key string) int {
	return int(toInt64(c.Get(key)))
}

// GetInt64 返回指定配置键的 int64 值。
func (c *httpConfig) GetInt64(key string) int64 {
	return toInt64(c.Get(key))
}

// GetDuration 返回指定配置键的 time.Duration 值。
func (c *httpConfig) GetDuration(key string) time.Duration {
	switch val := c.Get(key).(type) {
	case time.Duration:
		return val
	case string:
		if val == "" {
			return 0
		}
		duration, err := time.ParseDuration(val)
		if err == nil {
			return duration
		}
		return time.Duration(toInt64(val))
	default:
		return time.Duration(toInt64(val))
	}
}

// GetBool 返回指定配置键的 bool 值。
func (c *httpConfig) GetBool(key string) bool {
	switch val := c.Get(key).(type) {
	case bool:
		return val
	case string:
		parsed, err := strconv.ParseBool(val)
		return err == nil && parsed
	default:
		return toString(val) == "true"
	}
}

// Watch 注册远端配置变更回调函数。
//
// 只有通过 WithWatch 启用轮询后，远端配置变化才会触发回调。
func (c *httpConfig) Watch(fn func()) error {
	if fn == nil {
		return nil
	}

	c.mu.Lock()
	c.watchFns = append(c.watchFns, fn)
	c.mu.Unlock()

	if c.watch {
		c.watchOnce.Do(c.watchLoop)
	}
	return nil
}

// Close 停止后台配置轮询。
func (c *httpConfig) Close() error {
	c.cancel()
	return nil
}

func (c *httpConfig) reload() error {
	body, header, err := c.fetch()
	if err != nil {
		return err
	}

	values, err := parseConfig(body, c.format, c.url, header)
	if err != nil {
		return err
	}

	merged := mergeMaps(c.defaults, values)

	c.mu.Lock()
	c.values = merged
	c.lastBody = string(body)
	c.mu.Unlock()

	return nil
}

func (c *httpConfig) fetch() ([]byte, http.Header, error) {
	req, err := http.NewRequestWithContext(c.ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("httpadapter: create request: %w", err)
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("httpadapter: get %s: %w", c.url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, fmt.Errorf("httpadapter: get %s: unexpected status %d", c.url, resp.StatusCode)
	}

	readLimit := c.maxBodyBytes
	if readLimit < math.MaxInt64 {
		readLimit++
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, readLimit))
	if err != nil {
		return nil, nil, fmt.Errorf("httpadapter: read response body: %w", err)
	}
	if int64(len(body)) > c.maxBodyBytes {
		return nil, nil, fmt.Errorf("httpadapter: response body exceeds %d bytes", c.maxBodyBytes)
	}
	return body, resp.Header, nil
}

func (c *httpConfig) watchLoop() {
	interval := c.interval
	if interval <= 0 {
		interval = defaultTimeout
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-c.ctx.Done():
				return
			case <-ticker.C:
				changed, err := c.reloadIfChanged()
				if err != nil || !changed {
					continue
				}
				c.notifyWatchers()
			}
		}
	}()
}

func (c *httpConfig) reloadIfChanged() (bool, error) {
	body, header, err := c.fetch()
	if err != nil {
		return false, err
	}

	nextBody := string(body)
	c.mu.RLock()
	same := c.lastBody == nextBody
	c.mu.RUnlock()
	if same {
		return false, nil
	}

	values, err := parseConfig(body, c.format, c.url, header)
	if err != nil {
		return false, err
	}
	merged := mergeMaps(c.defaults, values)

	c.mu.Lock()
	c.values = merged
	c.lastBody = nextBody
	c.mu.Unlock()
	return true, nil
}

func (c *httpConfig) notifyWatchers() {
	c.mu.RLock()
	fns := append([]func(){}, c.watchFns...)
	c.mu.RUnlock()

	for _, fn := range fns {
		select {
		case <-c.ctx.Done():
			return
		default:
			fn()
		}
	}
}

func getValue(values map[string]any, key string) any {
	parts := strings.Split(key, ".")
	var current any = values
	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			if val, ok := values[key]; ok {
				return val
			}
			return nil
		}
		current, ok = m[part]
		if !ok {
			if val, ok := values[key]; ok {
				return val
			}
			return nil
		}
	}
	return current
}

func mergeMaps(defaults, values map[string]any) map[string]any {
	result := cloneMap(defaults)
	for k, v := range values {
		if defaultMap, ok := result[k].(map[string]any); ok {
			if valueMap, ok := v.(map[string]any); ok {
				result[k] = mergeMaps(defaultMap, valueMap)
				continue
			}
		}
		result[k] = cloneValue(v)
	}
	return result
}

func cloneMap(values map[string]any) map[string]any {
	result := make(map[string]any, len(values))
	for k, v := range values {
		result[k] = cloneValue(v)
	}
	return result
}

func cloneStringMap(values map[string]string) map[string]string {
	result := make(map[string]string, len(values))
	for k, v := range values {
		result[k] = v
	}
	return result
}

func cloneValue(value any) any {
	if value == nil {
		return nil
	}
	return cloneReflectValue(reflect.ValueOf(value)).Interface()
}

func cloneReflectValue(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		result := reflect.New(value.Type()).Elem()
		result.Set(cloneReflectValue(value.Elem()))
		return result
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		result := reflect.MakeMapWithSize(value.Type(), value.Len())
		iter := value.MapRange()
		for iter.Next() {
			result.SetMapIndex(iter.Key(), cloneReflectValue(iter.Value()))
		}
		return result
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for i := 0; i < value.Len(); i++ {
			result.Index(i).Set(cloneReflectValue(value.Index(i)))
		}
		return result
	case reflect.Array:
		result := reflect.New(value.Type()).Elem()
		for i := 0; i < value.Len(); i++ {
			result.Index(i).Set(cloneReflectValue(value.Index(i)))
		}
		return result
	default:
		return value
	}
}

func toString(val any) string {
	if val == nil {
		return ""
	}
	if str, ok := val.(string); ok {
		return str
	}
	return fmt.Sprint(val)
}

func toInt64(val any) int64 {
	switch v := val.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		if uint64(v) > math.MaxInt64 {
			return 0
		}
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		if v > math.MaxInt64 {
			return 0
		}
		return int64(v)
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	case json.Number:
		parsed, err := v.Int64()
		if err != nil {
			return 0
		}
		return parsed
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}
