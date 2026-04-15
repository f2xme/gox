package viper

import (
	"sync"
	"time"

	"github.com/f2xme/gox/config"
	"github.com/spf13/viper"
)

type viperConfig struct {
	v        *viper.Viper
	mu       sync.RWMutex
	watchFns []func()
}

var (
	_ config.Config  = (*viperConfig)(nil)
	_ config.Watcher = (*viperConfig)(nil)
)

func (c *viperConfig) Get(key string) any                     { return c.v.Get(key) }
func (c *viperConfig) GetString(key string) string            { return c.v.GetString(key) }
func (c *viperConfig) GetStringSlice(key string) []string     { return c.v.GetStringSlice(key) }
func (c *viperConfig) GetStringMap(key string) map[string]any { return c.v.GetStringMap(key) }
func (c *viperConfig) GetInt(key string) int                  { return c.v.GetInt(key) }
func (c *viperConfig) GetInt64(key string) int64              { return c.v.GetInt64(key) }
func (c *viperConfig) GetDuration(key string) time.Duration   { return c.v.GetDuration(key) }
func (c *viperConfig) GetBool(key string) bool                { return c.v.GetBool(key) }

// Set 设置配置键的默认值
//
// 注意：该方法设置的是默认值，不会覆盖配置文件中已存在的值。
// 如果需要在运行时修改配置，应该直接修改配置文件并依赖热重载机制。
func (c *viperConfig) Set(key string, value any) {
	c.v.SetDefault(key, value)
}

// Watch 注册配置文件变更的回调函数
//
// 当配置文件发生变更时，所有已注册的回调函数会被依次调用。
// 该方法是线程安全的，可以在多个 goroutine 中调用。
//
// 注意：
//   - 必须在创建配置时使用 WithWatch() 选项才能启用热重载
//   - 回调函数应避免阻塞操作，建议使用 goroutine 处理耗时任务
//
// 示例：
//
//	cfg.(config.Watcher).Watch(func() {
//		log.Println("配置已更新")
//		// 重新读取配置
//	})
func (c *viperConfig) Watch(fn func()) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.watchFns = append(c.watchFns, fn)
	return nil
}
