package env

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/f2xme/gox/config"
)

var keyReplacer = strings.NewReplacer(".", "_", "-", "_")

type envConfig struct {
	prefix        string
	defaults      map[string]any
	watch         bool
	watchInterval time.Duration
	mu            sync.RWMutex
	watchFns      []func()
	watchOnce     sync.Once
}

var (
	_ config.Config  = (*envConfig)(nil)
	_ config.Watcher = (*envConfig)(nil)
)

// New 创建一个由环境变量支持的 config.Config 实例。
func New(opts ...Option) config.Config {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	defaults := make(map[string]any, len(options.defaults))
	for k, v := range options.defaults {
		defaults[k] = v
	}

	return &envConfig{
		prefix:        normalizePrefix(options.prefix),
		defaults:      defaults,
		watch:         options.watch,
		watchInterval: options.watchInterval,
	}
}

// Get 返回指定配置键的原始值。
func (c *envConfig) Get(key string) any {
	if val, ok := os.LookupEnv(c.envKey(key)); ok {
		return val
	}
	return c.defaults[key]
}

// GetString 返回指定配置键的字符串值。
func (c *envConfig) GetString(key string) string {
	return toString(c.Get(key))
}

// GetStringSlice 返回指定配置键的字符串切片值。
func (c *envConfig) GetStringSlice(key string) []string {
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
func (c *envConfig) GetStringMap(key string) map[string]any {
	result := make(map[string]any)

	if defaults, ok := c.defaults[key].(map[string]any); ok {
		for k, v := range defaults {
			result[k] = v
		}
	}

	prefix := c.envKey(key)
	if prefix != "" {
		prefix += "_"
	}
	for _, item := range os.Environ() {
		name, val, ok := strings.Cut(item, "=")
		if !ok || !strings.HasPrefix(name, prefix) || name == prefix {
			continue
		}
		childKey := strings.ToLower(strings.TrimPrefix(name, prefix))
		result[childKey] = val
	}

	return result
}

// GetInt 返回指定配置键的 int 值。
func (c *envConfig) GetInt(key string) int {
	return int(toInt64(c.Get(key)))
}

// GetInt64 返回指定配置键的 int64 值。
func (c *envConfig) GetInt64(key string) int64 {
	return toInt64(c.Get(key))
}

// GetDuration 返回指定配置键的 time.Duration 值。
func (c *envConfig) GetDuration(key string) time.Duration {
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
func (c *envConfig) GetBool(key string) bool {
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

// Watch 注册环境变量变更回调函数。
//
// 只有通过 WithWatch 启用监听后，环境变量变化才会触发回调。
func (c *envConfig) Watch(fn func()) error {
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

func (c *envConfig) watchLoop() {
	interval := c.watchInterval
	if interval <= 0 {
		interval = defaultWatchInterval
	}

	go func() {
		previous := c.snapshot()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			current := c.snapshot()
			if equalSnapshot(previous, current) {
				continue
			}
			previous = current
			c.notifyWatchers()
		}
	}()
}

func (c *envConfig) snapshot() map[string]string {
	result := make(map[string]string)
	for _, item := range os.Environ() {
		name, val, ok := strings.Cut(item, "=")
		if !ok || !c.matchEnvName(name) {
			continue
		}
		result[name] = val
	}
	return result
}

func (c *envConfig) matchEnvName(name string) bool {
	if c.prefix == "" {
		return true
	}
	return name == c.prefix || strings.HasPrefix(name, c.prefix+"_")
}

func (c *envConfig) notifyWatchers() {
	c.mu.RLock()
	fns := append([]func(){}, c.watchFns...)
	c.mu.RUnlock()

	for _, fn := range fns {
		go fn()
	}
}

func (c *envConfig) envKey(key string) string {
	name := strings.ToUpper(keyReplacer.Replace(key))
	if c.prefix == "" {
		return name
	}
	if name == "" {
		return c.prefix
	}
	return c.prefix + "_" + name
}

func normalizePrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	prefix = strings.Trim(prefix, "_")
	return strings.ToUpper(prefix)
}

func equalSnapshot(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		other, ok := b[k]
		if !ok || other != v {
			return false
		}
	}
	return true
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
