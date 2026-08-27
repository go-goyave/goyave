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
//	import _ "goyave.dev/goyave/v5/database/dialect/clickhouse"
//	import _ "goyave.dev/goyave/v5/database/dialect/bigquery"
func New(cfg *config.DatabaseConnection, logger func() *slog.Logger) (*gorm.DB, error) { // TODO logger from context?
	dialect, ok := dialects[cfg.Dialect]
	if !ok {
		return nil, errorutil.Errorf("DB dialect %q not supported, forgotten import?", cfg.Dialect)
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
func NewFromDialector(cfg *config.DatabaseConnection, logger func() *slog.Logger, dialector gorm.Dialector) (*gorm.DB, error) {
	db, err := gorm.Open(dialector, newConfig(cfg, logger))
	if err != nil {
		return nil, errorutil.New(err)
	}

	if err := initTimeoutPlugin(cfg, db); err != nil {
		return db, errorutil.New(err)
	}

	return db, initSQLDB(cfg, db)
}

func newConfig(cfg *config.DatabaseConnection, logger func() *slog.Logger) *gorm.Config {
	return &gorm.Config{
		Logger:                                   NewLogger(logger),
		SkipDefaultTransaction:                   cfg.GORM.SkipDefaultTransaction,
		DryRun:                                   cfg.GORM.DryRun,
		PrepareStmt:                              cfg.GORM.PrepareStmt,
		PrepareStmtMaxSize:                       cfg.GORM.PrepareStmtMaxSize,
		PrepareStmtTTL:                           time.Duration(cfg.GORM.PrepareStmtTTL) * time.Second,
		DisableNestedTransaction:                 cfg.GORM.DisableNestedTransaction,
		AllowGlobalUpdate:                        cfg.GORM.AllowGlobalUpdate,
		DisableAutomaticPing:                     cfg.GORM.DisableAutomaticPing,
		DisableForeignKeyConstraintWhenMigrating: cfg.GORM.DisableForeignKeyConstraintWhenMigrating,
		IgnoreRelationshipsWhenMigrating:         cfg.GORM.IgnoreRelationshipsWhenMigrating,
		// DefaultContextTimeout: 0,
		FullSaveAssociations: cfg.GORM.FullSaveAssociations,
		QueryFields:          cfg.GORM.QueryFields,
		CreateBatchSize:      cfg.GORM.CreateBatchSize,
		TranslateError:       cfg.GORM.TranslateError,
		PropagateUnscoped:    cfg.GORM.PropagateUnscoped,
	}
}

func initTimeoutPlugin(cfg *config.DatabaseConnection, db *gorm.DB) error {
	timeoutPlugin := &TimeoutPlugin{
		ReadTimeout:  time.Duration(cfg.DefaultReadQueryTimeoutMs) * time.Millisecond,
		WriteTimeout: time.Duration(cfg.DefaultWriteQueryTimeoutMs) * time.Millisecond,
	}
	return errorutil.New(db.Use(timeoutPlugin))
}

func initSQLDB(cfg *config.DatabaseConnection, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		if errors.Is(err, gorm.ErrInvalidDB) {
			return nil
		}
		return errorutil.New(err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.MaxIdleTime) * time.Second)
	return nil
}
