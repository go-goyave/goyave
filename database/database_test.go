package database

import (
	"bytes"
	"testing"
	"time"

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

func TestNewDatabase(t *testing.T) {
	RegisterDialect("dummy", "host={host} port={port} user={username} dbname={name} password={password} {options}", openDummy)
	RegisterDialect("sqlite3_test", "file:{name}?{options}", sqlite.Open)
	t.Cleanup(func() {
		mu.Lock()
		delete(dialects, "dummy")
		delete(dialects, "sqlite3_test")
		mu.Unlock()
	})

	t.Run("RegisterDialect_already_exists", func(t *testing.T) {
		assert.Panics(t, func() {
			RegisterDialect("dummy", "", openDummy)
		})
	})

	t.Run("New", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("app.debug", true)
		cfg.Set("database.connection", "dummy")
		cfg.Set("database.host", "localhost")
		cfg.Set("database.port", 5432)
		cfg.Set("database.name", "dbname")
		cfg.Set("database.username", "user")
		cfg.Set("database.password", "secret")
		cfg.Set("database.options", "option=value")
		cfg.Set("database.maxOpenConnections", 123)
		cfg.Set("database.maxIdleConnections", 123)
		cfg.Set("database.maxLifetime", 123)
		cfg.Set("database.defaultReadQueryTimeout", 123)
		cfg.Set("database.defaultWriteQueryTimeout", 123)
		cfg.Set("database.config.skipDefaultTransaction", true)
		cfg.Set("database.config.dryRun", true)
		cfg.Set("database.config.prepareStmt", false)
		cfg.Set("database.config.disableNestedTransaction", true)
		cfg.Set("database.config.allowGlobalUpdate", true)
		cfg.Set("database.config.disableAutomaticPing", true)
		cfg.Set("database.config.disableForeignKeyConstraintWhenMigrating", true)

		slogger := slog.New(slog.NewHandler(true, &bytes.Buffer{}))
		db, err := New(cfg, func() *slog.Logger { return slogger })
		require.NoError(t, err)
		require.NotNil(t, db)

		if assert.NotNil(t, db.Config.Logger) {
			// Logging is enabled when app.debug is true
			l, ok := db.Config.Logger.(*Logger)
			if assert.True(t, ok) {
				assert.NotNil(t, l.slogger)
			}
		}

		dbConfig := db.Config
		// Can't check log level (gorm logger unexported)
		assert.True(t, dbConfig.SkipDefaultTransaction)
		assert.True(t, dbConfig.DryRun)
		assert.False(t, dbConfig.PrepareStmt)
		assert.True(t, dbConfig.DisableNestedTransaction)
		assert.True(t, dbConfig.AllowGlobalUpdate)
		assert.True(t, dbConfig.DisableAutomaticPing)
		assert.True(t, dbConfig.DisableAutomaticPing)

		// Cannot check the max open conns, idle conns and lifetime

		plugin, ok := db.Plugins[(&TimeoutPlugin{}).Name()]
		if assert.True(t, ok) {
			timeoutPlugin, ok := plugin.(*TimeoutPlugin)
			if assert.True(t, ok) {
				assert.Equal(t, 123*time.Millisecond, timeoutPlugin.ReadTimeout)
				assert.Equal(t, 123*time.Millisecond, timeoutPlugin.WriteTimeout)
			}
		}

	})

	t.Run("silent", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("app.debug", false)
		cfg.Set("database.connection", "dummy")
		cfg.Set("database.host", "localhost")
		cfg.Set("database.port", 5432)
		cfg.Set("database.name", "dbname")
		cfg.Set("database.username", "user")
		cfg.Set("database.password", "secret")
		cfg.Set("database.options", "option=value")
		cfg.Set("database.maxOpenConnections", 123)
		cfg.Set("database.maxIdleConnections", 123)
		cfg.Set("database.maxLifetime", 123)
		cfg.Set("database.defaultReadQueryTimeout", 123)
		cfg.Set("database.defaultWriteQueryTimeout", 123)
		cfg.Set("database.config.skipDefaultTransaction", true)
		cfg.Set("database.config.dryRun", true)
		cfg.Set("database.config.prepareStmt", false)
		cfg.Set("database.config.disableNestedTransaction", true)
		cfg.Set("database.config.allowGlobalUpdate", true)
		cfg.Set("database.config.disableAutomaticPing", true)
		cfg.Set("database.config.disableForeignKeyConstraintWhenMigrating", true)

		logger := slog.New(slog.NewHandler(false, &bytes.Buffer{}))
		db, err := New(cfg, func() *slog.Logger { return logger })
		require.NoError(t, err)
		require.NotNil(t, db)

		if assert.NotNil(t, db.Config.Logger) {
			// Logging is disable when app.debug is false
			l, ok := db.Config.Logger.(*Logger)
			if assert.True(t, ok) {
				assert.Nil(t, l.slogger)
			}
		}
	})

	t.Run("NewFromDialector", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("app.debug", true)
		cfg.Set("database.connection", "dummy")
		cfg.Set("database.host", "localhost")
		cfg.Set("database.port", 5432)
		cfg.Set("database.name", "dbname")
		cfg.Set("database.username", "user")
		cfg.Set("database.password", "secret")
		cfg.Set("database.options", "option=value")
		cfg.Set("database.maxOpenConnections", 123)
		cfg.Set("database.maxIdleConnections", 123)
		cfg.Set("database.maxLifetime", 123)
		cfg.Set("database.defaultReadQueryTimeout", 123)
		cfg.Set("database.defaultWriteQueryTimeout", 123)
		cfg.Set("database.config.skipDefaultTransaction", true)
		cfg.Set("database.config.dryRun", true)
		cfg.Set("database.config.prepareStmt", false)
		cfg.Set("database.config.disableNestedTransaction", true)
		cfg.Set("database.config.allowGlobalUpdate", true)
		cfg.Set("database.config.disableAutomaticPing", true)
		cfg.Set("database.config.disableForeignKeyConstraintWhenMigrating", true)

		dialector := &DummyDialector{}
		db, err := NewFromDialector(cfg, nil, dialector)
		require.NoError(t, err)
		require.NotNil(t, db)

		dbConfig := db.Config
		// Can't check log level (gorm logger unexported)
		assert.True(t, dbConfig.SkipDefaultTransaction)
		assert.True(t, dbConfig.DryRun)
		assert.False(t, dbConfig.PrepareStmt)
		assert.True(t, dbConfig.DisableNestedTransaction)
		assert.True(t, dbConfig.AllowGlobalUpdate)
		assert.True(t, dbConfig.DisableAutomaticPing)
		assert.True(t, dbConfig.DisableAutomaticPing)

		// Cannot check the max open conns, idle conns and lifetime

		plugin, ok := db.Plugins[(&TimeoutPlugin{}).Name()]
		if assert.True(t, ok) {
			timeoutPlugin, ok := plugin.(*TimeoutPlugin)
			if assert.True(t, ok) {
				assert.Equal(t, 123*time.Millisecond, timeoutPlugin.ReadTimeout)
				assert.Equal(t, 123*time.Millisecond, timeoutPlugin.WriteTimeout)
			}
		}

	})

	t.Run("New_connection_none", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("database.connection", "none")
		db, err := New(cfg, nil)
		assert.Nil(t, db)
		require.Error(t, err)
		assert.Equal(t, "Cannot create DB connection. Database is set to \"none\" in the config", err.Error())
	})

	t.Run("New_unknown_driver", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("database.connection", "notadriver")
		db, err := New(cfg, nil)
		assert.Nil(t, db)
		require.Error(t, err)
		assert.Equal(t, "DB Connection \"notadriver\" not supported, forgotten import?", err.Error())
	})

	t.Run("SQLite_query", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("app.debug", false)
		cfg.Set("database.connection", "sqlite3_test")
		cfg.Set("database.name", "database_test.db")
		cfg.Set("database.options", "mode=memory")

		db, err := New(cfg, nil)
		require.NoError(t, err)
		require.NotNil(t, db)

		dbNames := []string{}
		res := db.Table("pragma_database_list").Select("name").Find(&dbNames)
		require.NoError(t, res.Error)
		assert.Equal(t, []string{"main"}, dbNames)
	})
}
