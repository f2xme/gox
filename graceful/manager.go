package graceful

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"sync"
	"time"
)

type resource struct {
	name     string
	closer   Closer
	priority int
	timeout  time.Duration
}

type manager struct {
	mu        sync.Mutex
	resources []*resource
	signals   []os.Signal
	timeout   time.Duration
	logger    Logger
	hooks     hooks
	sorted    bool
}

type hooks struct {
	beforeShutdown func()
	afterShutdown  func()
	onTimeout      func(string)
}

// Logger 简单的日志接口
type Logger interface {
	Printf(format string, v ...interface{})
}

const logPrefix = "[graceful]"

type defaultLogger struct{}

func (l *defaultLogger) Printf(format string, v ...interface{}) {
	fmt.Printf(logPrefix+" "+format+"\n", v...)
}

// Register 添加一个在关闭时需要关闭的资源
func (m *manager) Register(name string, closer Closer, opts ...RegisterOption) {
	cfg := &registerConfig{
		priority: 0,
		timeout:  0,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	timeout := cfg.timeout
	if timeout == 0 {
		timeout = m.timeout
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.resources = append(m.resources, &resource{
		name:     name,
		closer:   closer,
		priority: cfg.priority,
		timeout:  timeout,
	})
	m.sorted = false
}

// Wait 阻塞直到收到关闭信号，然后关闭所有资源
func (m *manager) Wait() error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, m.signals...)

	sig := <-sigChan
	m.logger.Printf("received signal: %v", sig)

	ctx := context.Background()
	return m.Shutdown(ctx)
}

// Shutdown 立即启动关闭流程，不等待信号
func (m *manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	if !m.sorted {
		sort.Slice(m.resources, func(i, j int) bool {
			return m.resources[i].priority > m.resources[j].priority
		})
		m.sorted = true
	}
	resources := make([]*resource, len(m.resources))
	copy(resources, m.resources)
	m.mu.Unlock()

	if m.hooks.beforeShutdown != nil {
		m.hooks.beforeShutdown()
	}

	var firstErr error

	for _, res := range resources {
		m.logger.Printf("closing %s (priority: %d, timeout: %v)", res.name, res.priority, res.timeout)

		closeCtx, cancel := context.WithTimeout(ctx, res.timeout)
		err := m.closeResource(closeCtx, res)
		cancel()

		if err != nil {
			m.logger.Printf("error closing %s: %v", res.name, err)
			if firstErr == nil {
				firstErr = err
			}
		} else {
			m.logger.Printf("closed %s successfully", res.name)
		}
	}

	if m.hooks.afterShutdown != nil {
		m.hooks.afterShutdown()
	}

	return firstErr
}

func (m *manager) closeResource(ctx context.Context, res *resource) error {
	done := make(chan error, 1)

	go func() {
		done <- res.closer.Close(ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		if m.hooks.onTimeout != nil {
			m.hooks.onTimeout(res.name)
		}
		return fmt.Errorf("timeout closing %s", res.name)
	}
}
