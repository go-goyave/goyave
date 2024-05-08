package database

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/slog"

	errorutil "goyave.dev/goyave/v5/util/errors"
)

// New create a new connection pool using the settings defined in the given configuration.
//
// In order to use a specific driver / dialect ("mysql", "sqlite3", ...), you must not
// forget to blank-import it in your main file.
//
//	import _ "goyave.dev/goyave/v5/database/dialect/mysql"
//	import _ "goyave.dev/goyave/v5/database/dialect/postgres"
//	import _ "goyave.dev/goyave/v5/database/dialect/sqlite"
//	import _ "goyave.dev/goyave/v5/database/dialect/mssql"
func New(cfg *config.Config, logger func() *slog.Logger) (*gorm.DB, error) {
	driver := cfg.GetString("database.connection")

	if driver == "none" {
		return nil, errorutil.Errorf("Cannot create DB connection. Database is set to \"none\" in the config")
	}

	dialect, ok := dialects[driver]
	if !ok {
		return nil, errorutil.Errorf("DB Connection %q not supported, forgotten import?", driver)
	}

	dsn := dialect.buildDSN(cfg)
	db, err := gorm.Open(dialect.initializer(dsn), newConfig(cfg, logger))
	if err != nil {
		return nil, errorutil.New(err)
	}

	if err := initTimeoutPlugin(cfg, db); err != nil {
		return db, errorutil.New(err)
	}

	return db, initSQLDB(cfg, db)
}

// NewFromDialector create a new connection pool from a gorm dialector and using the settings
// defined in the given configuration.
//
// This can be used in tests to create a mock connection pool.
func NewFromDialector(cfg *config.Config, logger func() *slog.Logger, dialector gorm.Dialector) (*gorm.DB, error) {
	db, err := gorm.Open(dialector, newConfig(cfg, logger))
	if err != nil {
		return nil, errorutil.New(err)
	}

	if err := initTimeoutPlugin(cfg, db); err != nil {
		return db, errorutil.New(err)
	}

	return db, initSQLDB(cfg, db)
}

func newConfig(cfg *config.Config, logger func() *slog.Logger) *gorm.Config {
	if !cfg.GetBool("app.debug") {
		// Stay silent about DB operations when not in debug mode
		logger = nil
	}
	return &gorm.Config{
		Logger:                                   NewLogger(logger),
		SkipDefaultTransaction:                   cfg.GetBool("database.config.skipDefaultTransaction"),
		DryRun:                                   cfg.GetBool("database.config.dryRun"),
		PrepareStmt:                              cfg.GetBool("database.config.prepareStmt"),
		DisableNestedTransaction:                 cfg.GetBool("database.config.disableNestedTransaction"),
		AllowGlobalUpdate:                        cfg.GetBool("database.config.allowGlobalUpdate"),
		DisableAutomaticPing:                     cfg.GetBool("database.config.disableAutomaticPing"),
		DisableForeignKeyConstraintWhenMigrating: cfg.GetBool("database.config.disableForeignKeyConstraintWhenMigrating"),
	}
}

func initTimeoutPlugin(cfg *config.Config, db *gorm.DB) error {
	timeoutPlugin := &TimeoutPlugin{
		ReadTimeout:  time.Duration(cfg.GetInt("database.defaultReadQueryTimeout")) * time.Millisecond,
		WriteTimeout: time.Duration(cfg.GetInt("database.defaultWriteQueryTimeout")) * time.Millisecond,
	}
	return errorutil.New(db.Use(timeoutPlugin))
}

func initSQLDB(cfg *config.Config, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		if errors.Is(err, gorm.ErrInvalidDB) {
			return nil
		}
		return errorutil.New(err)
	}
	sqlDB.SetMaxOpenConns(cfg.GetInt("database.maxOpenConnections"))
	sqlDB.SetMaxIdleConns(cfg.GetInt("database.maxIdleConnections"))
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.GetInt("database.maxLifetime")) * time.Second)
	return nil
}
