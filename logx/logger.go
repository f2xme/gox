package logx

// Logger 定义日志记录器接口
//
// 所有日志实现都必须实现此接口，提供统一的日志记录方法。
type Logger interface {
	Info(msg string, metas ...Meta)
	Warn(msg string, metas ...Meta)
	Error(err error, metas ...Meta)
	// Fatal 记录致命错误日志并退出程序
	Fatal(err error, metas ...Meta)
}

// Flusher 定义刷新缓冲区接口
//
// 实现此接口的 Logger 可以手动刷新缓冲区。
type Flusher interface{ Flush() error }

// Syncer 定义同步接口
//
// 实现此接口的 Logger 可以同步日志到持久化存储。
type Syncer interface{ Sync() error }

// Stopper 定义停止接口
//
// 实现此接口的 Logger 可以优雅地停止日志记录。
type Stopper interface{ Stop() error }

// Meta 定义结构化日志字段接口
//
// 用于添加键值对形式的结构化字段到日志中。
type Meta interface {
	Key() string
	Value() any
}
