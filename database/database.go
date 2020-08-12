package database

import (
	"fmt"
	"sync"
	"time"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/jinzhu/gorm"
)

var dbConnection *gorm.DB
var mu sync.Mutex

var models []interface{}

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

// Close the database connections if they exist.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if dbConnection != nil {
		dbConnection.Close()
		dbConnection = nil
	}
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
	return db
}

func buildConnectionOptions(connection string) string { // TODO add a way to register a new dialect
	switch connection {
	case "mysql":
		return fmt.Sprintf(
			"%s:%s@(%s:%d)/%s?%s",
			config.GetString("database.username"),
			config.GetString("database.password"),
			config.GetString("database.host"),
			config.GetInt("database.port"),
			config.GetString("database.name"),
			config.GetString("database.options"),
		)
	case "postgres":
		return fmt.Sprintf(
			"host=%s port=%d user=%s dbname=%s password=%s %s",
			config.GetString("database.host"),
			config.GetInt("database.port"),
			config.GetString("database.username"),
			config.GetString("database.name"),
			config.GetString("database.password"),
			config.GetString("database.options"),
		)
	case "sqlite3":
		return config.GetString("database.name")
	case "mssql":
		return fmt.Sprintf(
			"sqlserver://%s:%s@%s:%d?database=%s&%s",
			config.GetString("database.username"),
			config.GetString("database.password"),
			config.GetString("database.host"),
			config.GetInt("database.port"),
			config.GetString("database.name"),
			config.GetString("database.options"),
		)
	}

	panic(fmt.Sprintf("DB Connection %s not supported", connection))
}
