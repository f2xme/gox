module github.com/f2xme/gox/database/adapter/sqlitedb

go 1.25.7

replace github.com/f2xme/gox => ../../../

require (
	github.com/f2xme/gox v0.14.2
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.42 // indirect
	golang.org/x/text v0.36.0 // indirect
)
