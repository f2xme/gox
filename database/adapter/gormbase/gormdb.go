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

func (g *GormDB) Engine() any { return g.db }

func (g *GormDB) Transaction(ctx context.Context, fn func(tx database.DB) error) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&GormDB{db: tx, sqlDB: g.sqlDB})
	})
}

func (g *GormDB) AutoMigrate(models ...any) error {
	return g.db.AutoMigrate(models...)
}

func (g *GormDB) Close() error {
	return g.sqlDB.Close()
}
