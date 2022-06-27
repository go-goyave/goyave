package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"goyave.dev/goyave/v4/config"
)

// New create a new connection pool using the settings defined in the given configuration.
//
// In order to use a specific driver / dialect ("mysql", "sqlite3", ...), you must not
// forget to blank-import it in your main file.
//
//   import _ "goyave.dev/goyave/v5/database/dialect/mysql"
//   import _ "goyave.dev/goyave/v5/database/dialect/postgres"
//   import _ "goyave.dev/goyave/v5/database/dialect/sqlite"
//   import _ "goyave.dev/goyave/v5/database/dialect/mssql"
func New(cfg *config.Config) (*gorm.DB, error) {
	driver := cfg.GetString("database.connection")

	if driver == "none" {
		return nil, fmt.Errorf("Cannot create DB connection. Database is set to \"none\" in the config")
	}

	dialect, ok := dialects[driver]
	if !ok {
		return nil, fmt.Errorf("DB Connection %q not supported, forgotten import?", driver)
	}

	dsn := dialect.buildDSNV5(cfg)
	db, err := gorm.Open(dialect.initializer(dsn), newConfigV5(cfg))
	if err != nil {
		return nil, err
	}

	initSQLDBV5(cfg, db)
	return db, err
}

// NewFromDialector create a new connection pool from a gorm dialector and using the settings
// defined in the given configuration.
//
// This can be used in tests to create a mock connection pool.
func NewFromDialector(cfg *config.Config, dialector gorm.Dialector) (*gorm.DB, error) {
	db, err := gorm.Open(dialector, newConfigV5(cfg))
	if err != nil {
		return nil, err
	}

	initSQLDBV5(cfg, db)
	return db, nil
}

func newConfigV5(cfg *config.Config) *gorm.Config {
	logLevel := logger.Silent
	if cfg.GetBool("app.debug") {
		logLevel = logger.Info
	}
	return &gorm.Config{
		Logger:                                   logger.Default.LogMode(logLevel),
		SkipDefaultTransaction:                   cfg.GetBool("database.config.skipDefaultTransaction"),
		DryRun:                                   cfg.GetBool("database.config.dryRun"),
		PrepareStmt:                              cfg.GetBool("database.config.prepareStmt"),
		DisableNestedTransaction:                 cfg.GetBool("database.config.disableNestedTransaction"),
		AllowGlobalUpdate:                        cfg.GetBool("database.config.allowGlobalUpdate"),
		DisableAutomaticPing:                     cfg.GetBool("database.config.disableAutomaticPing"),
		DisableForeignKeyConstraintWhenMigrating: cfg.GetBool("database.config.disableForeignKeyConstraintWhenMigrating"),
	}
}

func initSQLDBV5(cfg *config.Config, db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxOpenConns(cfg.GetInt("database.maxOpenConnections"))
	sqlDB.SetMaxIdleConns(cfg.GetInt("database.maxIdleConnections"))
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.GetInt("database.maxLifetime")) * time.Second)

	for _, initializer := range initializers {
		initializer(db)
	}
}

// Migrate migrates all registered models.
func MigrateV5(db *gorm.DB) error {
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return err
		}
	}
	return nil
}
