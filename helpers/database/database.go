package database

import (
	"fmt"
	"log"

	"github.com/System-Glitch/goyave/config"
	"github.com/jinzhu/gorm"
)

// New create a new DB connection
func New() *gorm.DB {
	connection := config.GetString("dbConnection")
	db, err := gorm.Open(connection, buildConnectionOptions(connection))
	if err != nil {
		panic(err)
	}

	db.LogMode(config.GetBool("debug"))
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
