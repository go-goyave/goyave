package bigquery

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"gorm.io/driver/bigquery"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/database"
	"goyave.dev/goyave/v5/util/errwrap"
)

var dbSystemNameBigquery = semconv.DBSystemNameKey.String("gcp.bigquery")

func init() {
	database.Register("bigquery", Dialect{})
}

type Dialect struct{}

func (Dialect) Open(conn gorm.ConnPool) gorm.Dialector {
	return &bigquery.Dialector{
		Config: &bigquery.Config{
			PreferSimpleProtocol: false,
			Conn:                 conn.(*sql.DB),
		},
	}
}

func (Dialect) Driver() driver.Driver {
	// Driver is not exported, we have no choice but to open
	// an empty connection and retrieve the driver...
	db, err := sql.Open("bigquery", "scanner")
	if err != nil {
		panic(errwrap.New(err))
	}

	driver := db.Driver()

	// Close to avoid any leak
	if err = db.Close(); err != nil {
		panic(errwrap.New(err))
	}
	return driver
}

func (Dialect) DSN(cfg database.DSNConfig) string {
	// location is optional for BigQuery
	// possible name values = ["{projectID}/{location}/{dataSet}", "{projectID}/{dataSet}"]
	return fmt.Sprintf(
		"bigquery://%s?%s",
		cfg.DatabaseName,
		cfg.Options,
	)
}

func (Dialect) Attributes() []attribute.KeyValue {
	return []attribute.KeyValue{
		dbSystemNameBigquery,
	}
}
