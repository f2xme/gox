package database

import "context"

// DB 定义数据库连接管理接口
type DB interface {
	Engine() any
	Transaction(ctx context.Context, fn func(tx DB) error) error
	AutoMigrate(models ...any) error
	Close() error
}
