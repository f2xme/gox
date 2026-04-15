package logx

import (
	"context"
	"runtime"
)

// ContextExtractor 定义从 context 中提取日志字段的函数类型
//
// 用于自动从 context 中提取通用字段（如 request_id、user_id 等）。
type ContextExtractor func(ctx context.Context) []Meta

var contextExtractor ContextExtractor

// SetContextExtractor 设置全局的 context 字段提取器
//
// 示例：
//
//	logx.SetContextExtractor(func(ctx context.Context) []logx.Meta {
//		if reqID := ctx.Value("request_id"); reqID != nil {
//			return []logx.Meta{logx.NewKV("request_id", reqID)}
//		}
//		return nil
//	})
func SetContextExtractor(fn ContextExtractor) {
	contextExtractor = fn
}

// ContextLogger 提供带 context 的日志记录器
//
// 自动添加调用位置和从 context 提取的字段。
type ContextLogger struct {
	ctx        context.Context
	metas      []Meta
	callerMeta Meta
}

func ctxWithSkip(ctx context.Context, skip int) *ContextLogger {
	if ctx == nil {
		ctx = context.Background()
	}
	_, file, line, _ := runtime.Caller(skip)
	callerMeta := NewKV("caller", formatCaller(file, line))

	cl := &ContextLogger{ctx: ctx, callerMeta: callerMeta}
	if contextExtractor != nil {
		cl.metas = contextExtractor(ctx)
	}
	return cl
}

// Ctx 创建一个带 context 的日志记录器
//
// 自动添加调用位置和从 context 提取的字段。
//
// 示例：
//
//	logx.Ctx(ctx).Info("processing request", logx.String("user_id", "123"))
func Ctx(ctx context.Context) *ContextLogger {
	return ctxWithSkip(ctx, 2)
}

// With 添加额外的日志字段
//
// 返回一个新的 ContextLogger，不修改原有实例。
//
// 示例：
//
//	logger := logx.Ctx(ctx).With(logx.String("user_id", "123"))
//	logger.Info("user action")
func (cl *ContextLogger) With(metas ...Meta) *ContextLogger {
	newMetas := make([]Meta, len(cl.metas)+len(metas))
	copy(newMetas, cl.metas)
	copy(newMetas[len(cl.metas):], metas)
	return &ContextLogger{ctx: cl.ctx, metas: newMetas, callerMeta: cl.callerMeta}
}

// WithCaller 设置自定义的调用位置
//
// 返回一个新的 ContextLogger，不修改原有实例。
func (cl *ContextLogger) WithCaller(caller string) *ContextLogger {
	return &ContextLogger{ctx: cl.ctx, metas: cl.metas, callerMeta: NewKV("caller", caller)}
}

func (cl *ContextLogger) buildMetas(extra []Meta) []Meta {
	total := 1 + len(cl.metas) + len(extra)
	all := make([]Meta, 0, total)
	all = append(all, cl.callerMeta)
	all = append(all, cl.metas...)
	all = append(all, extra...)
	return all
}

// Info 记录信息级别日志
func (cl *ContextLogger) Info(msg string, metas ...Meta) {
	globalLogger.Info(msg, cl.buildMetas(metas)...)
}

// Warn 记录警告级别日志
func (cl *ContextLogger) Warn(msg string, metas ...Meta) {
	globalLogger.Warn(msg, cl.buildMetas(metas)...)
}

// Error 记录错误级别日志
func (cl *ContextLogger) Error(err error, metas ...Meta) {
	if err == nil {
		return
	}
	globalLogger.Error(err, cl.buildMetas(metas)...)
}

// Fatal 记录致命错误日志并退出程序
func (cl *ContextLogger) Fatal(err error, metas ...Meta) {
	if err == nil {
		return
	}
	globalLogger.Fatal(err, cl.buildMetas(metas)...)
}

// InfoCtx 使用 context 记录信息级别日志
//
// 自动添加调用位置和从 context 提取的字段。
func InfoCtx(ctx context.Context, msg string, metas ...Meta) {
	ctxWithSkip(ctx, 2).Info(msg, metas...)
}

// WarnCtx 使用 context 记录警告级别日志
//
// 自动添加调用位置和从 context 提取的字段。
func WarnCtx(ctx context.Context, msg string, metas ...Meta) {
	ctxWithSkip(ctx, 2).Warn(msg, metas...)
}

// ErrorCtx 使用 context 记录错误级别日志
//
// 自动添加调用位置和从 context 提取的字段。
func ErrorCtx(ctx context.Context, err error, metas ...Meta) {
	ctxWithSkip(ctx, 2).Error(err, metas...)
}

// FatalCtx 使用 context 记录致命错误日志并退出程序
//
// 自动添加调用位置和从 context 提取的字段。
func FatalCtx(ctx context.Context, err error, metas ...Meta) {
	ctxWithSkip(ctx, 2).Fatal(err, metas...)
}
