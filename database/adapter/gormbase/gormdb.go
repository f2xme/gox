package gormbase

import (
	"context"
	"database/sql"

	"github.com/f2xme/gox/database"
	"gorm.io/gorm"
)

type GormDB struct {
	db    *gorm.DB
	sqlDB *sql.DB
}

var _ database.DB = (*GormDB)(nil)

// Unwrap 返回底层的 *gorm.DB 实例
func (g *GormDB) Unwrap() any { return g.db }

// Create 插入新记录
func (g *GormDB) Create(ctx context.Context, value any) error {
	return g.db.WithContext(ctx).Create(value).Error
}

// First 查询第一条匹配的记录
func (g *GormDB) First(ctx context.Context, dest any, conds ...any) error {
	return g.db.WithContext(ctx).First(dest, conds...).Error
}

// Find 查询所有匹配的记录
func (g *GormDB) Find(ctx context.Context, dest any, conds ...any) error {
	return g.db.WithContext(ctx).Find(dest, conds...).Error
}

// Save 保存记录（插入或更新）
func (g *GormDB) Save(ctx context.Context, value any) error {
	return g.db.WithContext(ctx).Save(value).Error
}

// Update 更新单个字段
func (g *GormDB) Update(ctx context.Context, model any, column string, value any) error {
	return g.db.WithContext(ctx).Model(model).Update(column, value).Error
}

// Updates 更新多个字段
func (g *GormDB) Updates(ctx context.Context, model any, values any) error {
	return g.db.WithContext(ctx).Model(model).Updates(values).Error
}

// Delete 删除记录
func (g *GormDB) Delete(ctx context.Context, value any, conds ...any) error {
	return g.db.WithContext(ctx).Delete(value, conds...).Error
}

// Where 添加查询条件
func (g *GormDB) Where(query any, args ...any) database.DB {
	return &GormDB{db: g.db.Where(query, args...), sqlDB: g.sqlDB}
}

// Order 指定排序
func (g *GormDB) Order(value any) database.DB {
	return &GormDB{db: g.db.Order(value), sqlDB: g.sqlDB}
}

// Limit 限制返回记录数
func (g *GormDB) Limit(limit int) database.DB {
	return &GormDB{db: g.db.Limit(limit), sqlDB: g.sqlDB}
}

// Offset 设置偏移量
func (g *GormDB) Offset(offset int) database.DB {
	return &GormDB{db: g.db.Offset(offset), sqlDB: g.sqlDB}
}

// Select 指定要查询的字段
func (g *GormDB) Select(query any, args ...any) database.DB {
	return &GormDB{db: g.db.Select(query, args...), sqlDB: g.sqlDB}
}

// Omit 指定要忽略的字段
func (g *GormDB) Omit(columns ...string) database.DB {
	return &GormDB{db: g.db.Omit(columns...), sqlDB: g.sqlDB}
}

// Joins 指定 JOIN 条件
func (g *GormDB) Joins(query string, args ...any) database.DB {
	return &GormDB{db: g.db.Joins(query, args...), sqlDB: g.sqlDB}
}

// Group 指定 GROUP BY
func (g *GormDB) Group(name string) database.DB {
	return &GormDB{db: g.db.Group(name), sqlDB: g.sqlDB}
}

// Having 指定 HAVING 条件
func (g *GormDB) Having(query any, args ...any) database.DB {
	return &GormDB{db: g.db.Having(query, args...), sqlDB: g.sqlDB}
}

// Preload 预加载关联
func (g *GormDB) Preload(query string, args ...any) database.DB {
	return &GormDB{db: g.db.Preload(query, args...), sqlDB: g.sqlDB}
}

// Model 指定模型
func (g *GormDB) Model(value any) database.DB {
	return &GormDB{db: g.db.Model(value), sqlDB: g.sqlDB}
}

// Count 统计记录数
func (g *GormDB) Count(ctx context.Context, count *int64) error {
	return g.db.WithContext(ctx).Count(count).Error
}

// Begin 开始事务
func (g *GormDB) Begin(ctx context.Context) (database.DB, error) {
	tx := g.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &GormDB{db: tx, sqlDB: g.sqlDB}, nil
}

// Commit 提交事务
func (g *GormDB) Commit() error {
	return g.db.Commit().Error
}

// Rollback 回滚事务
func (g *GormDB) Rollback() error {
	return g.db.Rollback().Error
}

// Transaction 执行事务
func (g *GormDB) Transaction(ctx context.Context, fn func(tx database.DB) error) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&GormDB{db: tx, sqlDB: g.sqlDB})
	})
}

// Exec 执行原生 SQL
func (g *GormDB) Exec(ctx context.Context, sql string, values ...any) error {
	return g.db.WithContext(ctx).Exec(sql, values...).Error
}

// Raw 创建原生 SQL 查询
func (g *GormDB) Raw(sql string, values ...any) database.DB {
	return &GormDB{db: g.db.Raw(sql, values...), sqlDB: g.sqlDB}
}

// Scan 扫描查询结果
func (g *GormDB) Scan(ctx context.Context, dest any) error {
	return g.db.WithContext(ctx).Scan(dest).Error
}

// AutoMigrate 自动迁移表结构
func (g *GormDB) AutoMigrate(models ...any) error {
	return g.db.AutoMigrate(models...)
}

// Close 关闭数据库连接
func (g *GormDB) Close() error {
	return g.sqlDB.Close()
}

// WithContext 设置 context
func (g *GormDB) WithContext(ctx context.Context) database.DB {
	return &GormDB{db: g.db.WithContext(ctx), sqlDB: g.sqlDB}
}
