package logx

type nopLogger struct{}

func (nopLogger) Info(string, ...Meta)  {}
func (nopLogger) Warn(string, ...Meta)  {}
func (nopLogger) Error(error, ...Meta)  {}
func (nopLogger) Fatal(error, ...Meta)  {}

var globalLogger Logger = nopLogger{}

// Init 初始化全局日志记录器
//
// 设置后，包级别的日志函数（Info、Warn、Error、Fatal）将使用此 logger。
func Init(l Logger) {
	globalLogger = l
}

// Info 使用全局 logger 记录信息级别日志
func Info(msg string, metas ...Meta) {
	globalLogger.Info(msg, metas...)
}

// Warn 使用全局 logger 记录警告级别日志
func Warn(msg string, metas ...Meta) {
	globalLogger.Warn(msg, metas...)
}

// Error 使用全局 logger 记录错误级别日志
//
// 如果 err 为 nil，则不记录。
func Error(err error, metas ...Meta) {
	if err == nil {
		return
	}
	globalLogger.Error(err, metas...)
}

// Fatal 使用全局 logger 记录致命错误日志并退出程序
//
// 如果 err 为 nil，则不记录。
func Fatal(err error, metas ...Meta) {
	if err == nil {
		return
	}
	globalLogger.Fatal(err, metas...)
}

// Flush 刷新全局 logger 的缓冲区
//
// 如果全局 logger 实现了 Flusher 接口，则调用其 Flush 方法。
func Flush() error {
	if f, ok := globalLogger.(Flusher); ok {
		return f.Flush()
	}
	return nil
}

// Stop 停止全局 logger
//
// 如果全局 logger 实现了 Stopper 接口，则调用其 Stop 方法。
func Stop() error {
	if s, ok := globalLogger.(Stopper); ok {
		return s.Stop()
	}
	return nil
}

// Sync 同步全局 logger
//
// 如果全局 logger 实现了 Syncer 接口，则调用其 Sync 方法。
func Sync() error {
	if s, ok := globalLogger.(Syncer); ok {
		return s.Sync()
	}
	return nil
}
