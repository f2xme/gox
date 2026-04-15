package graceful

import (
	"context"
	"database/sql"
	"io"
	"net/http"
)

// HTTPServer 为 HTTP 服务器创建 Closer
func HTTPServer(server *http.Server) Closer {
	return CloserFunc(func(ctx context.Context) error {
		return server.Shutdown(ctx)
	})
}

// DBCloser 为数据库连接创建 Closer
func DBCloser(db *sql.DB) Closer {
	return CloserFunc(func(ctx context.Context) error {
		return db.Close()
	})
}

// IOCloser 为 io.Closer 创建 Closer
func IOCloser(closer io.Closer) Closer {
	return CloserFunc(func(ctx context.Context) error {
		return closer.Close()
	})
}

// GenericCloser 从不接受 context 的函数创建 Closer
func GenericCloser(fn func() error) Closer {
	return CloserFunc(func(ctx context.Context) error {
		return fn()
	})
}
