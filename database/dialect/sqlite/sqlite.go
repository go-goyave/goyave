package sqlite

import (
	"database/sql/driver"
	"fmt"

	"github.com/mattn/go-sqlite3"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"goyave.dev/goyave/v6/database"
)

func init() {
	database.Register("sqlite3", Dialect{})
}

type Dialect struct{}

func (Dialect) Open(conn gorm.ConnPool) gorm.Dialector {
	return sqlite.New(sqlite.Config{
		Conn: conn,
	})
}

func (Dialect) Driver() driver.Driver {
	return &sqlite3.SQLiteDriver{}
}

func (Dialect) DSN(cfg database.DSNConfig) string {
	return fmt.Sprintf(
		"file:%s?%s",
		cfg.DatabaseName,
		cfg.Options,
	)
}

func (Dialect) Attributes() []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.DBSystemNameSQLite,
	}
}
