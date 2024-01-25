package database

import (
	"strconv"
	"strings"
	"sync"

	"gorm.io/gorm"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/util/errors"
)

var (
	mu sync.Mutex

	dialects = map[string]dialect{}

	optionPlaceholders = map[string]string{
		"{username}": "database.username",
		"{password}": "database.password",
		"{host}":     "database.host",
		"{name}":     "database.name",
		"{options}":  "database.options",
	}
)

// DialectorInitializer function initializing a GORM Dialector using the given
// data source name (DSN).
type DialectorInitializer func(dsn string) gorm.Dialector

type dialect struct {
	initializer DialectorInitializer
	template    string
}

func (d dialect) buildDSN(cfg *config.Config) string {
	connStr := d.template
	for k, v := range optionPlaceholders {
		connStr = strings.Replace(connStr, k, cfg.GetString(v), 1)
	}
	connStr = strings.Replace(connStr, "{port}", strconv.Itoa(cfg.GetInt("database.port")), 1)

	return connStr
}

// RegisterDialect registers a connection string template for the given dialect.
//
// You cannot override a dialect that already exists.
//
// Template format accepts the following placeholders, which will be replaced with
// the corresponding configuration entries automatically:
//   - "{username}"
//   - "{password}"
//   - "{host}"
//   - "{port}"
//   - "{name}"
//   - "{options}"
//
// Example template for the "mysql" dialect:
//
//	{username}:{password}@({host}:{port})/{name}?{options}
func RegisterDialect(name, template string, initializer DialectorInitializer) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := dialects[name]; ok {
		panic(errors.Errorf("dialect %q already exists", name))
	}
	dialects[name] = dialect{initializer, template}
}
