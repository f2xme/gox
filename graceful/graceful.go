package graceful

import (
	"context"
)

// Manager 管理资源的优雅关闭
type Manager interface {
	// Register 添加一个在关闭时需要关闭的资源
	// 资源按优先级顺序关闭（优先级高的先关闭）
	Register(name string, closer Closer, opts ...RegisterOption)

	// Wait 阻塞直到收到关闭信号，然后关闭所有资源
	Wait() error

	// Shutdown 立即启动关闭流程，不等待信号
	Shutdown(ctx context.Context) error
}

// Closer 表示可以被关闭的资源
type Closer interface {
	Close(ctx context.Context) error
}

// CloserFunc 是 Closer 接口的函数适配器
type CloserFunc func(ctx context.Context) error

func (f CloserFunc) Close(ctx context.Context) error {
	return f(ctx)
}
