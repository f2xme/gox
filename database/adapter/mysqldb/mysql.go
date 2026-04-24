package mysqldb

import (
	"github.com/f2xme/gox/database"
	"github.com/f2xme/gox/database/adapter/internal/gormbase"
)

// MySQLDB 在 gormbase.GormDB 之上标识 MySQL 数据源，便于挂载 MySQL 专有 API
type MySQLDB struct {
	*gormbase.GormDB
}

var _ database.DB = (*MySQLDB)(nil)
