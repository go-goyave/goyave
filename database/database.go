package database

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"goyave.dev/goyave/v4/config"
)

// Initializer is a function meant to modify a connection settings
// at the global scope when it's created.
//
// Use `db.InstantSet()` and not `db.Set()`, since the latter clones
// the gorm.DB instance instead of modifying it.
type Initializer func(*gorm.DB)

// DialectorInitializer function initializing a GORM Dialector using the given
// data source name (DSN).
type DialectorInitializer func(dsn string) gorm.Dialector

type dialect struct {
	initializer DialectorInitializer
	template    string
}

var (
	dbConnection *gorm.DB
	mu           sync.Mutex
	models       []interface{}
	initializers []Initializer

	dialects = map[string]dialect{}

	optionPlaceholders = map[string]string{
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

// SetConnection manually replace the automatic DB connection.
// If a connection already exists, closes it before discarding it.
// This can be used to create a mock DB in tests. Using this function
// is not recommended outside of tests. Prefer using a custom dialect.
func SetConnection(dialector gorm.Dialector) (*gorm.DB, error) {
	mu.Lock()
	defer mu.Unlock()
	if dbConnection != nil {
		if err := closeDB(); err != nil {
			return nil, err
		}
	}
	db, err := gorm.Open(dialector, newConfig())
	if err != nil {
		return nil, err
	}

	initSQLDB(db)
	dbConnection = db
	return dbConnection, nil
}

func closeDB() error {
	var err error
	if dbConnection != nil {
		db, _ := dbConnection.DB()
		err = db.Close()
		dbConnection = nil
	}
	return err
}

// Close the database connections if they exist.
func Close() error {
	mu.Lock()
	defer mu.Unlock()
	return closeDB()
}

// AddInitializer adds a database connection initializer function.
// Initializer functions are meant to modify a connection settings
// at the global scope when it's created.
//
// Initializer functions are called in order, meaning that functions
// added last can override settings defined by previous ones.
func AddInitializer(initializer Initializer) {
	initializers = append(initializers, initializer)
}

// ClearInitializers remove all database connection initializer functions.
func ClearInitializers() {
	initializers = []Initializer{}
}

// RegisterModel registers a model for auto-migration.
// When writing a model file, you should always register it in the init() function.
//
//	 func init() {
//			database.RegisterModel(&MyModel{})
//	 }
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
		if err := db.AutoMigrate(model); err != nil {
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
		panic(fmt.Sprintf("Dialect %q already exists", name))
	}
	dialects[name] = dialect{initializer, template}
}

func newConnection() *gorm.DB {
	driver := config.GetString("database.connection")

	if driver == "none" {
		panic("Cannot create DB connection. Database is set to \"none\" in the config")
	}

	dialect, ok := dialects[driver]
	if !ok {
		panic(fmt.Sprintf("DB Connection %q not supported, forgotten import?", driver))
	}

	dsn := dialect.buildDSN()
	db, err := gorm.Open(dialect.initializer(dsn), newConfig())
	if err != nil {
		panic(err)
	}

	initSQLDB(db)
	return db
}

func newConfig() *gorm.Config {
	logLevel := logger.Silent
	if config.GetBool("app.debug") {
		logLevel = logger.Info
	}
	return &gorm.Config{
		Logger:                                   logger.Default.LogMode(logLevel),
		SkipDefaultTransaction:                   config.GetBool("database.config.skipDefaultTransaction"),
		DryRun:                                   config.GetBool("database.config.dryRun"),
		PrepareStmt:                              config.GetBool("database.config.prepareStmt"),
		DisableNestedTransaction:                 config.GetBool("database.config.disableNestedTransaction"),
		AllowGlobalUpdate:                        config.GetBool("database.config.allowGlobalUpdate"),
		DisableAutomaticPing:                     config.GetBool("database.config.disableAutomaticPing"),
		DisableForeignKeyConstraintWhenMigrating: config.GetBool("database.config.disableForeignKeyConstraintWhenMigrating"),
	}
}

func initSQLDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxOpenConns(config.GetInt("database.maxOpenConnections"))
	sqlDB.SetMaxIdleConns(config.GetInt("database.maxIdleConnections"))
	sqlDB.SetConnMaxLifetime(time.Duration(config.GetInt("database.maxLifetime")) * time.Second)

	for _, initializer := range initializers {
		initializer(db)
	}
}

func (d dialect) buildDSN() string {
	connStr := d.template
	for k, v := range optionPlaceholders {
		connStr = strings.Replace(connStr, k, config.GetString(v), 1)
	}
	connStr = strings.Replace(connStr, "{port}", strconv.Itoa(config.GetInt("database.port")), 1)

	return connStr
}
