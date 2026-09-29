package mssql

import (
	"database/sql/driver"
	"fmt"

	mssql "github.com/microsoft/go-mssqldb"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"goyave.dev/goyave/v6/database"
)

func init() {
	database.Register("mssql", Dialect{})
}

type Dialect struct{}

func (Dialect) Open(conn gorm.ConnPool) gorm.Dialector {
	return sqlserver.New(sqlserver.Config{
		Conn: conn,
	})
}

func (Dialect) Driver() driver.Driver {
	return &mssql.Driver{}
}

func (Dialect) DSN(cfg database.DSNConfig) string {
	return fmt.Sprintf(
		"sqlserver://%s:%s@%s:%d?database=%s&%s",
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
		semconv.DBSystemNameMicrosoftSQLServer,
	}
}
