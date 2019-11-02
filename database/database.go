package database

import (
	"fmt"
	"log"

	"github.com/System-Glitch/goyave/config"
	"github.com/jinzhu/gorm"
)

var dbConnection *gorm.DB

var models []interface{}

// GetConnection returns the global database connection pool.
// Creates a new connection pool if no connection is available.
//
// The connections will be closed automatically on server shutdown so you
// don't need to call "Close()" when you're done with the database.
func GetConnection() *gorm.DB {
	if dbConnection == nil {
		dbConnection = newConnection()
	}
	return dbConnection
}

// Close the database connections if they exist.
func Close() {
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

// Migrate migrates all registered models.
func Migrate() {
	db := GetConnection()
	for _, model := range models {
		db.AutoMigrate(model)
	}
}

func newConnection() *gorm.DB {
	connection := config.GetString("dbConnection")

	if connection == "none" {
		log.Panicf("Cannot create DB connection. Database is set to \"none\" in the config")
	}

	db, err := gorm.Open(connection, buildConnectionOptions(connection))
	if err != nil {
		panic(err)
	}

	db.LogMode(config.GetBool("debug"))
	db.DB().SetMaxOpenConns(int(config.Get("dbMaxOpenConnections").(float64)))
	db.DB().SetMaxIdleConns(int(config.Get("dbMaxIdleConnections").(float64)))
	return db
}

func buildConnectionOptions(connection string) string {
	switch connection {
	case "mysql":
		return fmt.Sprintf(
			"%s:%s@(%s:%d)/%s?%s",
			config.GetString("dbUsername"),
			config.GetString("dbPassword"),
			config.GetString("dbHost"),
			int64(config.Get("dbPort").(float64)),
			config.GetString("dbName"),
			config.GetString("dbOptions"),
		)
	case "postgres":
		return fmt.Sprintf(
			"host=%s port=%d user=%s dbname=%s password=%s options='%s'",
			config.GetString("dbHost"),
			int64(config.Get("dbPort").(float64)),
			config.GetString("dbUsername"),
			config.GetString("dbName"),
			config.GetString("dbPassword"),
			config.GetString("dbOptions"),
		)
	case "sqlite3":
		return config.GetString("dbName")
	case "mssql":
		return fmt.Sprintf(
			"sqlserver://%s:%s@%s:%d?database=%s&%s",
			config.GetString("dbUsername"),
			config.GetString("dbPassword"),
			config.GetString("dbHost"),
			int64(config.Get("dbPort").(float64)),
			config.GetString("dbName"),
			config.GetString("dbOptions"),
		)
	}

	log.Fatalf("DB Connection %s not supported", connection)
	return ""
}
