package database

import "context"

// DB 定义数据库连接管理接口
type DB interface {
	// 基础 CRUD 操作
	Create(ctx context.Context, value any) error
	First(ctx context.Context, dest any, conds ...any) error
	Find(ctx context.Context, dest any, conds ...any) error
	Save(ctx context.Context, value any) error
	Update(ctx context.Context, model any, column string, value any) error
	Updates(ctx context.Context, model any, values any) error
	Delete(ctx context.Context, value any, conds ...any) error

	// 查询构建
	Where(query any, args ...any) DB
	Order(value any) DB
	Limit(limit int) DB
	Offset(offset int) DB
	Select(query any, args ...any) DB
	Omit(columns ...string) DB
	Joins(query string, args ...any) DB
	Group(name string) DB
	Having(query any, args ...any) DB
	Preload(query string, args ...any) DB
	Model(value any) DB
	Count(ctx context.Context, count *int64) error

	// 事务控制
	Begin(ctx context.Context) (DB, error)
	Commit() error
	Rollback() error
	Transaction(ctx context.Context, fn func(tx DB) error) error

	// 原生 SQL
	Exec(ctx context.Context, sql string, values ...any) error
	Raw(sql string, values ...any) DB
	Scan(ctx context.Context, dest any) error

	// 工具方法
	AutoMigrate(models ...any) error
	Close() error
	WithContext(ctx context.Context) DB

	// 底层访问（不推荐直接使用，但保留以支持高级用法）
	Unwrap() any
}
