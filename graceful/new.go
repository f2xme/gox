package graceful

import (
	"os"
	"syscall"
	"time"
)

// New 创建新的优雅关闭管理器
//
// 参数：
//   - opts: 配置选项函数
//
// 返回值：
//   - Manager: 管理器实例
//
// 示例：
//
//	mgr := graceful.New(
//		graceful.WithTimeout(10 * time.Second),
//		graceful.WithSignals(syscall.SIGTERM, syscall.SIGINT),
//	)
func New(opts ...Option) Manager {
	m := &manager{
		resources: make([]*resource, 0),
		signals:   []os.Signal{syscall.SIGTERM, syscall.SIGINT},
		timeout:   30 * time.Second,
		logger:    &defaultLogger{},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}
