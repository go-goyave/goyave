package database

import (
	"database/sql"
	"errors"
	"time"

	"go.opentelemetry.io/otel/metric"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"goyave.dev/goyave/v5/util/errwrap"

	"github.com/XSAM/otelsql"
)

const otelRegistrationKey = "goyave.otel.meter_registration"

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
func New(cfg *Config, opts ...Option) (*gorm.DB, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	dialect, ok := dialects[cfg.Dialect]
	if !ok {
		return nil, errwrap.Errorf("DB dialect %q not supported, forgotten import?", cfg.Dialect)
	}

	connector := connector{
		driver: dialect.Driver(),
		dsn:    dialect.DSN(cfg.DSNConfig),
	}

	if o.otelTraceProvier != nil {
		connector.driver = otelsql.WrapDriver(connector.driver, o.getOTelOptions(dialect)...)
	}

	sqlDB := sql.OpenDB(connector)

	db, err := gorm.Open(dialect.Open(sqlDB), newConfig(cfg))
	if err != nil {
		return nil, errwrap.New(err)
	}

	if err := initTimeoutPlugin(cfg, db); err != nil {
		return db, errwrap.New(err)
	}

	return db, initSQLDB(cfg, db, dialect, o)
}

// NewFromDialector create a new connection pool from a [gorm.Dialector] and using the settings
// defined in the given configuration.
//
// This can be used in tests to create a mock connection pool.
//
// Note that connections opened using this function cannot use OpenTelemetry unless you open the
// underlying [gorm.ConnPool] manually with [otelsql.Open] or [otelsql.WrapDriver].
func NewFromDialector(cfg *Config, dialector gorm.Dialector) (*gorm.DB, error) {
	db, err := gorm.Open(dialector, newConfig(cfg))
	if err != nil {
		return nil, errwrap.New(err)
	}

	if err := initTimeoutPlugin(cfg, db); err != nil {
		return db, errwrap.New(err)
	}

	return db, initSQLDB(cfg, db, nil, nil)
}

func newConfig(cfg *Config) *gorm.Config {
	var logger logger.Interface
	if cfg.Debug {
		logger = NewLogger()
	} else {
		logger = NewDiscardLogger()
	}
	return &gorm.Config{
		Logger:                                   logger,
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
		// DefaultContextTimeout: 0, // Handled by the timeout plugin
		// DefaultTransactionTimeout: 0, // Newly added option but also suffering from context leak
		FullSaveAssociations: cfg.GORM.FullSaveAssociations,
		QueryFields:          cfg.GORM.QueryFields,
		CreateBatchSize:      cfg.GORM.CreateBatchSize,
		TranslateError:       cfg.GORM.TranslateError,
		PropagateUnscoped:    cfg.GORM.PropagateUnscoped,
	}
}

func initTimeoutPlugin(cfg *Config, db *gorm.DB) error {
	timeoutPlugin := &TimeoutPlugin{
		ReadTimeout:  time.Duration(cfg.DefaultReadQueryTimeoutMs) * time.Millisecond,
		WriteTimeout: time.Duration(cfg.DefaultWriteQueryTimeoutMs) * time.Millisecond,
	}
	return errwrap.New(db.Use(timeoutPlugin))
}

func initSQLDB(cfg *Config, db *gorm.DB, dialect Dialect, o *options) error {
	sqlDB, err := db.DB()
	if err != nil {
		if errors.Is(err, gorm.ErrInvalidDB) {
			return nil
		}
		return errwrap.New(err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.MaxIdleTime) * time.Second)

	if o != nil && o.otelMeterProvider != nil {
		reg, err := otelsql.RegisterDBStatsMetrics(sqlDB, o.getOTelOptions(dialect)...)
		if err != nil {
			return errwrap.New(err)
		}
		// Store the registration in the GORM store so we can unregister later.
		db.Statement.Settings.Store(otelRegistrationKey, reg)
	}

	return nil
}

// Close the [*sql.DB] used by the GORM instance and unregister
// the OpenTelemetry metric callback, if any.
func Close(db *gorm.DB) error {
	if db == nil || db.Statement == nil || db.Config == nil {
		return nil
	}
	reg, ok := db.Get(otelRegistrationKey)
	if ok {
		_ = reg.(metric.Registration).Unregister()
	}

	sqlDB, err := db.DB()
	if err != nil {
		if errors.Is(err, gorm.ErrInvalidDB) {
			return nil
		}
		return errwrap.New(err)
	}
	return errwrap.New(sqlDB.Close())
}
