package pgsqldb

import (
	"github.com/f2xme/gox/database"
	"github.com/f2xme/gox/database/adapter/internal/gormbase"
)

// PostgresDB 在 gormbase.GormDB 之上标识 PostgreSQL 数据源，便于挂载 PostgreSQL 专有 API
type PostgresDB struct {
	*gormbase.GormDB
}

var _ database.DB = (*PostgresDB)(nil)

