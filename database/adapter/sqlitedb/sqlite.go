package sqlitedb

import (
	"github.com/f2xme/gox/database"
	"github.com/f2xme/gox/database/adapter/internal/gormbase"
)

// SQLiteDB 在 gormbase.GormDB 之上标识 SQLite 数据源，便于挂载 SQLite 专有 API
type SQLiteDB struct {
	*gormbase.GormDB
}

var _ database.DB = (*SQLiteDB)(nil)

