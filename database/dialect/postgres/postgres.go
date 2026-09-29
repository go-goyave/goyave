package postgres

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"goyave.dev/goyave/v6/database"
)

func init() {
	database.Register("postgres", Dialect{})
}

type Dialect struct{}

func (Dialect) Open(conn gorm.ConnPool) gorm.Dialector {
	return postgres.New(postgres.Config{
		PreferSimpleProtocol: false,
		WithoutQuotingCheck:  false,
		WithoutReturning:     false,
		Conn:                 conn,
	})
}

func (Dialect) Driver() driver.Driver {
	return stdlib.GetDefaultDriver()
}

func (Dialect) DSN(cfg database.DSNConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s %s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.DatabaseName,
		cfg.Password,
		cfg.Options,
	)
}

func (Dialect) Attributes() []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.DBSystemNamePostgreSQL,
	}
}
