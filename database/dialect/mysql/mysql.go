package mysql

import (
	"github.com/System-Glitch/goyave/v3/database"
	"gorm.io/driver/mysql"
)

func init() {
	database.RegisterDialect("mysql", "{username}:{password}@({host}:{port})/{name}?{options}", mysql.Open)
}
