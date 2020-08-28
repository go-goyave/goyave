package database

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/System-Glitch/goyave/v3/config"
	"github.com/jinzhu/gorm"
)

// Initializer is a function meant to modify a connection settings
// at the global scope when it's created.
//
// Use `db.InstantSet()` and not `db.Set()`, since the latter clones
// the gorm.DB instance instead of modifying it.
type Initializer func(*gorm.DB)

var (
	dbConnection *gorm.DB
	mu           sync.Mutex
	models       []interface{}
	initializers []Initializer

	dialectOptions map[string]string = map[string]string{
		"mysql":    "{username}:{password}@({host}:{port})/{name}?{options}",
		"postgres": "host={host} port={port} user={username} dbname={name} password={password} {options}",
		"sqlite3":  "{name}",
		"mssql":    "sqlserver://{username}:{password}@{host}:{port}?database={name}&{options}",
	}

	optionPlaceholders map[string]string = map[string]string{
		"{username}": "database.username",
		"{password}": "database.password",
		"{host}":     "database.host",
		"{name}":     "database.name",
		"{options}":  "database.options",
	}
)

// GetConnection returns the global database connection pool.
// Creates a new connection pool if no connection is available.
//
// The connections will be closed automatically on server shutdown so you
// don't need to call "Close()" when you're done with the database.
func GetConnection() *gorm.DB {
	mu.Lock()
	defer mu.Unlock()
	if dbConnection == nil {
		dbConnection = newConnection()
	}
	return dbConnection
}

// Conn alias for GetConnection.
func Conn() *gorm.DB {
	return GetConnection()
}

// Close the database connections if they exist.
func Close() error {
	var err error = nil
	mu.Lock()
	defer mu.Unlock()
	if dbConnection != nil {
		err = dbConnection.Close()
		dbConnection = nil
	}

	return err
}

// AddInitializer adds a database connection initializer function.
// Initializer functions are meant to modify a connection settings
// at the global scope when it's created.
//
// Initializer functions are called in order, meaning that functions
// added last can override settings defined by previous ones.
//
// Use `db.InstantSet()` and not `db.Set()`, since the latter clones
// the gorm.DB instance instead of modifying it.
func AddInitializer(initializer Initializer) {
	initializers = append(initializers, initializer)
}

// ClearInitializers remove all database connection initializer functions.
func ClearInitializers() {
	initializers = []Initializer{}
}

// RegisterModel registers a model for auto-migration.
// When writing a model file, you should always register it in the init() function.
//  func init() {
//		database.RegisterModel(&MyModel{})
//  }
func RegisterModel(model interface{}) {
	models = append(models, model)
}

// GetRegisteredModels get the registered models.
// The returned slice is a copy of the original, so it
// cannot be modified.
func GetRegisteredModels() []interface{} {
	return append(make([]interface{}, 0, len(models)), models...)
}

// ClearRegisteredModels unregister all models.
func ClearRegisteredModels() {
	models = []interface{}{}
}

// Migrate migrates all registered models.
func Migrate() {
	db := GetConnection()
	for _, model := range models {
		if err := db.AutoMigrate(model).Error; err != nil {
			panic(err)
		}
	}
}

// RegisterDialect registers a connection string template for the given dialect.
//
// You cannot override a dialect that already exists.
//
// Template format accepts the following placeholders, which will be replaced with
// the corresponding configuration entries automatically:
//  - "{username}"
//  - "{password}"
//  - "{host}"
//  - "{port}"
//  - "{name}"
//  - "{options}"
// Example template for the "mysql" dialect:
//  {username}:{password}@({host}:{port})/{name}?{options}
func RegisterDialect(name, template string) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := dialectOptions[name]; ok {
		panic(fmt.Sprintf("Dialect %q already exists", name))
	}
	dialectOptions[name] = template
}

func newConnection() *gorm.DB {
	connection := config.GetString("database.connection")

	if connection == "none" {
		panic("Cannot create DB connection. Database is set to \"none\" in the config")
	}

	db, err := gorm.Open(connection, buildConnectionOptions(connection))
	if err != nil {
		panic(err)
	}

	db.LogMode(config.GetBool("app.debug"))
	db.DB().SetMaxOpenConns(config.GetInt("database.maxOpenConnections"))
	db.DB().SetMaxIdleConns(config.GetInt("database.maxIdleConnections"))
	db.DB().SetConnMaxLifetime(time.Duration(config.GetInt("database.maxLifetime")) * time.Second)
	for _, initializer := range initializers {
		initializer(db)
	}
	return db
}

func buildConnectionOptions(driver string) string {
	template, ok := dialectOptions[driver]
	if !ok {
		panic(fmt.Sprintf("DB Connection %q not supported", driver))
	}

	connStr := template
	for k, v := range optionPlaceholders {
		connStr = strings.Replace(connStr, k, config.GetString(v), 1)
	}
	connStr = strings.Replace(connStr, "{port}", strconv.Itoa(config.GetInt("database.port")), 1)

	return connStr
}
