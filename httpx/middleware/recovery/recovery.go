package recovery

import (
	"fmt"

	"github.com/f2xme/gox/httpx"
)

// New 创建一个 recovery 中间件，用于捕获 panic 并将其转换为错误。
func New(opts ...Option) httpx.Middleware {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	return func(next httpx.Handler) httpx.Handler {
		return func(ctx httpx.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					if e, ok := r.(error); ok {
						err = e
					} else {
						err = fmt.Errorf("%v", r)
					}
					if o.Handler != nil {
						o.Handler(ctx, err)
					}
				}
			}()
			return next(ctx)
		}
	}
}
