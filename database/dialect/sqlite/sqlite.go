package sqlite

import (
	"gorm.io/driver/sqlite"
	"goyave.dev/goyave/v5/database"
)

func init() {
	database.RegisterDialect("sqlite3", "file:{name}?{options}", sqlite.Open)
}
