package mysql

import (
	"gorm.io/driver/mysql"
	"goyave.dev/goyave/v4/database"
)

func init() {
	database.RegisterDialect("mysql", "{username}:{password}@({host}:{port})/{name}?{options}", mysql.Open)
}
