module github.com/f2xme/gox/database/adapter/mysqldb

go 1.25.7

replace github.com/f2xme/gox => ../../../

require (
	github.com/f2xme/gox v0.0.0-00010101000000-000000000000
	gorm.io/driver/mysql v1.6.0
	gorm.io/gorm v1.31.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/text v0.36.0 // indirect
)
