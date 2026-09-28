package clickhouse

import (
	"database/sql/driver"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	gormclickhouse "gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/database"
)

func init() {
	database.Register("clickhouse", Dialect{})
}

type Dialect struct{}

func (Dialect) Open(conn gorm.ConnPool) gorm.Dialector {
	return gormclickhouse.New(gormclickhouse.Config{
		Conn: conn,
	})
}

func (Dialect) Driver() driver.Driver {
	return clickhouse.Connector(nil).Driver()
}

func (Dialect) DSN(cfg database.DSNConfig) string {
	return fmt.Sprintf(
		"clickhouse://%s:%s@%s:%d/%s?%s",
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
		semconv.DBSystemNameClickHouse,
	}
}
