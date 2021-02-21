package sqlite

import (
	"github.com/System-Glitch/goyave/v3/database"
	"gorm.io/driver/sqlite"
)

func init() {
	database.RegisterDialect("sqlite3", "file:{name}?{options}", sqlite.Open)
}
