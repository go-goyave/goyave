package database

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/slog"
)

type DummyDialector struct {
	tests.DummyDialector
	DSN string
}

func openDummy(dsn string) gorm.Dialector {
	return &DummyDialector{
		DSN: dsn,
	}
}

var testConnectionConfig = &config.DatabaseConnection{
	Dialect:                    "dummy",
	Host:                       "localhost",
	Port:                       5432,
	DatabaseName:               "dbname",
	Username:                   "user",
	Password:                   "secret",
	Options:                    "option=value",
	MaxOpenConnections:         123,
	MaxIdleConnections:         123,
	MaxLifetime:                123,
	MaxIdleTime:                123,
	DefaultReadQueryTimeoutMs:  123,
	DefaultWriteQueryTimeoutMs: 123,
	GORM: config.GORM{
		SkipDefaultTransaction:                   true,
		PrepareStmtMaxSize:                       123,
		PrepareStmtTTL:                           123,
		CreateBatchSize:                          123,
		PrepareStmt:                              false,
		DryRun:                                   true,
		DisableNestedTransaction:                 true,
		AllowGlobalUpdate:                        true,
		DisableAutomaticPing:                     true,
		FullSaveAssociations:                     true,
		QueryFields:                              true,
		PropagateUnscoped:                        true,
		TranslateError:                           true,
		DisableForeignKeyConstraintWhenMigrating: true,
		IgnoreRelationshipsWhenMigrating:         true,
	},
}

func TestNewDatabase(t *testing.T) {
	RegisterDialect("dummy", "host={host} port={port} user={username} dbname={name} password={password} {options}", openDummy)
	t.Cleanup(func() {
		mu.Lock()
		delete(dialects, "dummy")
		mu.Unlock()
	})

	t.Run("RegisterDialect_already_exists", func(t *testing.T) {
		assert.Panics(t, func() {
			RegisterDialect("dummy", "", openDummy)
		})
	})

	t.Run("New", func(t *testing.T) {
		slogger := slog.New(slog.NewHandler(true, &bytes.Buffer{}))
		db, err := New(testConnectionConfig, func() *slog.Logger { return slogger })
		require.NoError(t, err)
		require.NotNil(t, db)

		if assert.NotNil(t, db.Logger) {
			// Logging is enabled when app.debug is true
			l, ok := db.Logger.(*Logger)
			if assert.True(t, ok) {
				assert.NotNil(t, l.slogger)
			}
		}

		// Can't check log level (gorm logger unexported)
		assert.Equal(t, testConnectionConfig.GORM.SkipDefaultTransaction, db.SkipDefaultTransaction)
		assert.Equal(t, testConnectionConfig.GORM.AllowGlobalUpdate, db.AllowGlobalUpdate)
		assert.Equal(t, testConnectionConfig.GORM.CreateBatchSize, db.CreateBatchSize)
		assert.Equal(t, testConnectionConfig.GORM.DisableAutomaticPing, db.DisableAutomaticPing)
		assert.Equal(t, testConnectionConfig.GORM.DisableForeignKeyConstraintWhenMigrating, db.DisableForeignKeyConstraintWhenMigrating)
		assert.Equal(t, testConnectionConfig.GORM.DisableNestedTransaction, db.DisableNestedTransaction)
		assert.Equal(t, testConnectionConfig.GORM.DryRun, db.DryRun)
		assert.Equal(t, testConnectionConfig.GORM.FullSaveAssociations, db.FullSaveAssociations)
		assert.Equal(t, testConnectionConfig.GORM.IgnoreRelationshipsWhenMigrating, db.IgnoreRelationshipsWhenMigrating)
		assert.Equal(t, testConnectionConfig.GORM.PrepareStmt, db.PrepareStmt)
		assert.Equal(t, testConnectionConfig.GORM.PrepareStmtMaxSize, db.PrepareStmtMaxSize)
		assert.Equal(t, time.Duration(testConnectionConfig.GORM.PrepareStmtTTL)*time.Second, db.PrepareStmtTTL)
		assert.Equal(t, testConnectionConfig.GORM.PropagateUnscoped, db.PropagateUnscoped)
		assert.Equal(t, testConnectionConfig.GORM.QueryFields, db.QueryFields)
		assert.Equal(t, testConnectionConfig.GORM.TranslateError, db.TranslateError)

		// Cannot check the max open conns, idle conns and lifetime

		plugin, ok := db.Plugins[(&TimeoutPlugin{}).Name()]
		if assert.True(t, ok) {
			timeoutPlugin, ok := plugin.(*TimeoutPlugin)
			if assert.True(t, ok) {
				assert.Equal(t, 123*time.Millisecond, timeoutPlugin.ReadTimeout)
				assert.Equal(t, 123*time.Millisecond, timeoutPlugin.WriteTimeout)
			}
		}

		assert.Equal(t, "dummy", db.Name())
	})

	t.Run("silent", func(t *testing.T) {
		db, err := New(testConnectionConfig, nil)
		require.NoError(t, err)
		require.NotNil(t, db)

		require.NotNil(t, db.Logger)
		l, ok := db.Logger.(*Logger)
		if assert.True(t, ok) {
			assert.Nil(t, l.slogger)
		}
	})

	t.Run("NewFromDialector", func(t *testing.T) {
		dialector := &DummyDialector{}
		db, err := NewFromDialector(testConnectionConfig, nil, dialector)
		require.NoError(t, err)
		require.NotNil(t, db)

		// Can't check log level (gorm logger unexported)
		assert.Equal(t, testConnectionConfig.GORM.SkipDefaultTransaction, db.SkipDefaultTransaction)
		assert.Equal(t, testConnectionConfig.GORM.AllowGlobalUpdate, db.AllowGlobalUpdate)
		assert.Equal(t, testConnectionConfig.GORM.CreateBatchSize, db.CreateBatchSize)
		assert.Equal(t, testConnectionConfig.GORM.DisableAutomaticPing, db.DisableAutomaticPing)
		assert.Equal(t, testConnectionConfig.GORM.DisableForeignKeyConstraintWhenMigrating, db.DisableForeignKeyConstraintWhenMigrating)
		assert.Equal(t, testConnectionConfig.GORM.DisableNestedTransaction, db.DisableNestedTransaction)
		assert.Equal(t, testConnectionConfig.GORM.DryRun, db.DryRun)
		assert.Equal(t, testConnectionConfig.GORM.FullSaveAssociations, db.FullSaveAssociations)
		assert.Equal(t, testConnectionConfig.GORM.IgnoreRelationshipsWhenMigrating, db.IgnoreRelationshipsWhenMigrating)
		assert.Equal(t, testConnectionConfig.GORM.PrepareStmt, db.PrepareStmt)
		assert.Equal(t, testConnectionConfig.GORM.PrepareStmtMaxSize, db.PrepareStmtMaxSize)
		assert.Equal(t, time.Duration(testConnectionConfig.GORM.PrepareStmtTTL)*time.Second, db.PrepareStmtTTL)
		assert.Equal(t, testConnectionConfig.GORM.PropagateUnscoped, db.PropagateUnscoped)
		assert.Equal(t, testConnectionConfig.GORM.QueryFields, db.QueryFields)
		assert.Equal(t, testConnectionConfig.GORM.TranslateError, db.TranslateError)

		// Cannot check the max open conns, idle conns and lifetime

		plugin, ok := db.Plugins[(&TimeoutPlugin{}).Name()]
		if assert.True(t, ok) {
			timeoutPlugin, ok := plugin.(*TimeoutPlugin)
			if assert.True(t, ok) {
				assert.Equal(t, 123*time.Millisecond, timeoutPlugin.ReadTimeout)
				assert.Equal(t, 123*time.Millisecond, timeoutPlugin.WriteTimeout)
			}
		}

		assert.Equal(t, "dummy", db.Name())
	})

	t.Run("New_unknown_driver", func(t *testing.T) {
		cfg := &config.DatabaseConnection{
			Dialect: "notadriver",
		}
		db, err := New(cfg, nil)
		assert.Nil(t, db)
		require.Error(t, err)
		assert.Equal(t, "DB dialect \"notadriver\" not supported, forgotten import?", err.Error())
	})

	t.Run("SQLite_query", func(t *testing.T) {
		cfg := &config.DatabaseConnection{
			Dialect:            "sqlmock",
			DatabaseName:       "paginator_test.db",
			MaxIdleConnections: 1,
			GORM:               config.GORM{}, // Disabling PrepareStmt is important to avoid errors caused by mock
		}

		mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		require.NoError(t, err)

		dialector := &sqlite.Dialector{
			DriverName: "sqlite3_timeout_test",
			DSN:        fmt.Sprintf("file:%s?%s", cfg.DatabaseName, cfg.Options),
			Conn:       mockDB,
		}

		t.Cleanup(func() {
			mock.ExpectClose()
			assert.NoError(t, mockDB.Close())
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		// The SQLite dialector selects the sqlite version first to know which callback clauses it can use.
		mock.ExpectQuery(regexp.QuoteMeta(`select sqlite_version()`)).WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("3.53.4"))

		db, err := NewFromDialector(cfg, nil, dialector)
		if err != nil {
			require.NoError(t, err)
		}

		mock.ExpectQuery(regexp.QuoteMeta("SELECT name FROM `pragma_database_list`")).WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("main"))

		dbNames := []string{}
		res := db.Table("pragma_database_list").Select("name").Find(&dbNames)
		require.NoError(t, res.Error)
		assert.Equal(t, []string{"main"}, dbNames)
	})
}
