package trace

import "context"

// EndFunc Span 结束函数类型
type EndFunc func(errPtr *error)

// Service 创建 Service 层 Span，返回结束函数
//
// 使用示例：
//
//	func (s *Service) GetUser(ctx context.Context, id int64) (user *User, err error) {
//	    defer trace.Service(ctx, "GetUser", "id", id)(&err)
//	    // ...
//	}
func Service(ctx context.Context, name string, attrs ...any) EndFunc {
	return start(ctx, SpanKindService, name, attrs...)
}

// DAO 创建 DAO 层 Span，返回结束函数
//
// 使用示例：
//
//	func (d *Dao) GetUserByID(ctx context.Context, id int64) (user *User, err error) {
//	    defer trace.DAO(ctx, "GetUserByID", "id", id)(&err)
//	    // ...
//	}
func DAO(ctx context.Context, name string, attrs ...any) EndFunc {
	return start(ctx, SpanKindDAO, name, attrs...)
}

// Cache 创建 Cache 层 Span，返回结束函数
func Cache(ctx context.Context, name string, attrs ...any) EndFunc {
	return start(ctx, SpanKindCache, name, attrs...)
}

// RPC 创建 RPC 调用 Span，返回结束函数
func RPC(ctx context.Context, name string, attrs ...any) EndFunc {
	return start(ctx, SpanKindRPC, name, attrs...)
}

// MQ 创建消息队列 Span，返回结束函数
func MQ(ctx context.Context, name string, attrs ...any) EndFunc {
	return start(ctx, SpanKindMQ, name, attrs...)
}

// start 内部方法，创建 Span 并返回结束函数
func start(ctx context.Context, kind SpanKind, name string, attrs ...any) EndFunc {
	span := StartSpan(ctx, kind, name)

	// 解析属性（key-value 对）
	for i := 0; i+1 < len(attrs); i += 2 {
		if key, ok := attrs[i].(string); ok {
			span.Set(key, attrs[i+1])
		}
	}

	return func(errPtr *error) {
		var err error
		if errPtr != nil {
			err = *errPtr
		}
		result := span.End(err)

		if callback != nil {
			callback(result)
		}
	}
}

// Callback Span 结束回调函数类型
type Callback func(*SpanResult)

var callback Callback // 注意：应在程序启动时设置，避免运行时并发修改

// SetCallback 设置全局 Span 结束回调
// 用于记录日志、上报指标等
// 警告：非并发安全，应仅在程序启动时调用一次
//
// 使用示例：
//
//	trace.SetCallback(func(r *trace.SpanResult) {
//	    fields := []log.Meta{
//	        log.NewKV("span_kind", r.Kind()),
//	        log.NewKV("span_name", r.Name()),
//	        log.NewKV("duration_ms", r.DurationMs()),
//	    }
//	    for k, v := range r.Attrs() {
//	        fields = append(fields, log.NewKV(k, v))
//	    }
//	    if r.Error != nil {
//	        log.Ctx(r.Context()).Error(r.Error, fields...)
//	    } else {
//	        log.Ctx(r.Context()).Info("span", fields...)
//	    }
//	})
func SetCallback(cb Callback) {
	callback = cb
}
