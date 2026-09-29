package mysql

import (
	"database/sql/driver"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"goyave.dev/goyave/v6/database"
)

func init() {
	database.Register("mysql", Dialect{})
}

type Dialect struct{}

func (Dialect) Open(conn gorm.ConnPool) gorm.Dialector {
	return gormmysql.New(gormmysql.Config{
		Conn: conn,
	})
}

func (Dialect) Driver() driver.Driver {
	return &mysql.MySQLDriver{}
}

func (Dialect) DSN(cfg database.DSNConfig) string {
	return fmt.Sprintf(
		"%s:%s@(%s:%d)/%s?%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DatabaseName,
		cfg.Options,
	)
}

func (Dialect) Attributes() []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.DBSystemNameMySQL,
	}
}
