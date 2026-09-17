package database

import (
	"strconv"
	"strings"
	"sync"

	"gorm.io/gorm"
	"goyave.dev/goyave/v5/util/errors"
)

var (
	mu sync.Mutex

	dialects = map[string]dialect{}

	optionPlaceholders = map[string]func(*Config) string{
		"{username}": func(dc *Config) string { return dc.Username },
		"{password}": func(dc *Config) string { return dc.Password },
		"{host}":     func(dc *Config) string { return dc.Host },
		"{port}":     func(dc *Config) string { return strconv.Itoa(dc.Port) },
		"{name}":     func(dc *Config) string { return dc.DatabaseName },
		"{options}":  func(dc *Config) string { return dc.Options },
	}
)

// DialectorInitializer function initializing a GORM Dialector using the given
// data source name (DSN).
type DialectorInitializer func(dsn string) gorm.Dialector

type dialect struct {
	initializer DialectorInitializer
	template    string
}

func (d dialect) buildDSN(cfg *Config) string {
	connStr := d.template
	for k, v := range optionPlaceholders {
		connStr = strings.Replace(connStr, k, v(cfg), 1)
	}

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
