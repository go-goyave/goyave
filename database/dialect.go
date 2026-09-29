package database

import (
	"context"
	"database/sql/driver"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
	"goyave.dev/goyave/v6/util/errwrap"
)

var (
	mu sync.Mutex

	dialects = map[string]Dialect{}
)

// Dialect provides information and implementations needed to open a database connection pool
// and interface it with a matching GORM dialector.
//
// Thanks to this interface, [New] can automatically setup OpenTelemetry for the SQL driver if
// trace/meter providers are given as [Option].
type Dialect interface {
	Open(conn gorm.ConnPool) gorm.Dialector
	Driver() driver.Driver
	// DSN generate a connection string from the provided configuration.
	DSN(cfg DSNConfig) string
	// Attributes return the OpenTelemtry attributes the tracer/meters will be initialized with.
	// It MUST include the "db.system.name" attribute.
	Attributes() []attribute.KeyValue
}

func Register(name string, dialect Dialect) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := dialects[name]; ok {
		panic(errwrap.Errorf("dialect %q already exists", name))
	}
	dialects[name] = dialect
}

type connector struct {
	driver driver.Driver
	dsn    string
}

func (c connector) Connect(_ context.Context) (driver.Conn, error) {
	conn, err := c.driver.Open(c.dsn)
	if err != nil {
		return nil, errwrap.New(err)
	}
	return conn, nil
}

func (c connector) Driver() driver.Driver {
	return c.driver
}
