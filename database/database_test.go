package database

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/XSAM/otelsql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type dummyDialector struct {
	tests.DummyDialector
	conn gorm.ConnPool
}

func (d *dummyDialector) Initialize(db *gorm.DB) error {
	_ = d.DummyDialector.Initialize(db)
	db.ConnPool = d.conn
	return nil
}

type dummyDialect struct {
	name   string
	driver driver.Driver
}

func (d dummyDialect) Attributes() []attribute.KeyValue {
	return []attribute.KeyValue{semconv.DBSystemNameOtherSQL}
}

func (d dummyDialect) DSN(_ DSNConfig) string {
	return d.name
}

func (d dummyDialect) Driver() driver.Driver {
	return d.driver
}

func (d dummyDialect) Open(c gorm.ConnPool) gorm.Dialector {
	return &dummyDialector{conn: c}
}

var testConnectionConfig = &Config{
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
	Debug:                      true,
	GORM: GORMConfig{
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

func TestDatabase(t *testing.T) {
	t.Run("RegisterDialect_already_exists", func(t *testing.T) {
		Register("dummy", &dummyDialect{})
		t.Cleanup(func() {
			mu.Lock()
			delete(dialects, "dummy")
			mu.Unlock()
		})
		assert.Panics(t, func() {
			Register("dummy", &dummyDialect{})
		})
	})

	t.Run("New", func(t *testing.T) {
		Register("dummy", &dummyDialect{})
		t.Cleanup(func() {
			mu.Lock()
			delete(dialects, "dummy")
			mu.Unlock()
		})

		db, err := New(testConnectionConfig)
		require.NoError(t, err)
		require.NotNil(t, db)

		if assert.NotNil(t, db.Logger) {
			assert.IsType(t, &Logger{}, db.Logger)
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

		assert.NoError(t, Close(db))
	})

	t.Run("silent", func(t *testing.T) {
		Register("dummy", &dummyDialect{})
		t.Cleanup(func() {
			mu.Lock()
			delete(dialects, "dummy")
			mu.Unlock()
		})

		cfg := *testConnectionConfig
		cfg.Debug = false
		db, err := New(&cfg)
		require.NoError(t, err)
		require.NotNil(t, db)

		require.NotNil(t, db.Logger)
		assert.IsType(t, &DiscardLogger{}, db.Logger)
	})

	t.Run("NewFromDialector", func(t *testing.T) {
		db, err := NewFromDialector(testConnectionConfig, tests.DummyDialector{})
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
		cfg := &Config{
			Dialect: "notadriver",
			Debug:   false,
		}
		db, err := New(cfg)
		assert.Nil(t, db)
		require.Error(t, err)
		assert.Equal(t, "DB dialect \"notadriver\" not supported, forgotten import?", err.Error())
	})

	t.Run("SQLite_query", func(t *testing.T) {
		cfg := &Config{
			Dialect:            "sqlmock",
			DatabaseName:       "database_test.db",
			MaxIdleConnections: 1,
			Debug:              false,
			GORM:               GORMConfig{}, // Disabling PrepareStmt is important to avoid errors caused by mock
		}

		mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		require.NoError(t, err)

		dialector := &sqlite.Dialector{
			DriverName: "sqlite3_timeout_test",
			DSN:        fmt.Sprintf("file:%s?%s", cfg.DatabaseName, cfg.Options),
			Conn:       mockDB,
		}

		// The SQLite dialector selects the sqlite version first to know which callback clauses it can use.
		mock.ExpectQuery(regexp.QuoteMeta(`select sqlite_version()`)).WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("3.53.4"))

		db, err := NewFromDialector(cfg, dialector)
		if err != nil {
			require.NoError(t, err)
		}
		t.Cleanup(func() {
			mock.ExpectClose()
			assert.NoError(t, Close(db))
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		mock.ExpectQuery(regexp.QuoteMeta("SELECT name FROM `pragma_database_list`")).WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("main"))

		dbNames := []string{}
		res := db.Table("pragma_database_list").Select("name").Find(&dbNames)
		require.NoError(t, res.Error)
		assert.Equal(t, []string{"main"}, dbNames)
	})

	t.Run("OpenTelemetry", func(t *testing.T) {
		meterReader := sdkmetric.NewManualReader()
		meterProvider := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(meterReader),
		)

		spanRecorder := tracetest.NewSpanRecorder()
		traceProvider := sdktrace.NewTracerProvider(
			sdktrace.WithSpanProcessor(spanRecorder),
		)

		t.Cleanup(func() {
			_ = meterProvider.Shutdown(t.Context())
			_ = traceProvider.Shutdown(t.Context())
		})

		// We use t.Name() as DSN / identifier for the mock
		// When New opens the database using dummyDialect, it will use this DSN, connecting the dots
		// between the mock and the newly created *sql.DB.
		// The *sql.DB returned when opening with New should be mockDB.
		mockDB, mock, err := sqlmock.NewWithDSN(t.Name(), sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		require.NoError(t, err)

		Register("dummy", &dummyDialect{name: t.Name(), driver: mockDB.Driver()})
		t.Cleanup(func() {
			mu.Lock()
			delete(dialects, "dummy")
			mu.Unlock()
		})
		t.Cleanup(func() {
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		mock.ExpectQuery(regexp.QuoteMeta("SELECT 1")).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

		cfg := &Config{
			Dialect:            "dummy",
			DatabaseName:       "database_test.db",
			MaxIdleConnections: 1,
			Debug:              false,
			GORM:               GORMConfig{},
		}

		db, err := New(
			cfg,
			WithTraceProvider(traceProvider),
			WithMeterProvider(meterProvider),
			WithOpenTelemetryOptions(otelsql.WithAttributes(attribute.Bool("test", true))),
		)
		require.NoError(t, err)
		require.NotNil(t, db)

		dst := int64(0)
		err = db.Raw("SELECT 1").Scan(&dst).Error
		require.NoError(t, err)

		spans := spanRecorder.Ended()
		if assert.Len(t, spans, 2) { // sql.conn.query + sql.rows
			querySpan := spans[0]

			assert.Equal(t, string(otelsql.MethodConnQuery), querySpan.Name())

			attrs := querySpan.Attributes()
			assert.Contains(t, attrs, semconv.DBSystemNameOtherSQL)
			assert.Contains(t, attrs, attribute.Bool("test", true))

			rowsSpan := spans[1]
			assert.Equal(t, string(otelsql.MethodRows), rowsSpan.Name())

			attrs = rowsSpan.Attributes()
			assert.Contains(t, attrs, semconv.DBSystemNameOtherSQL)
			assert.Contains(t, attrs, attribute.Bool("test", true))

			// No need for more assertions on the span, it would be like re-testing the entire otelsql instrumentation.
		}

		var metrics metricdata.ResourceMetrics
		require.NoError(t, meterReader.Collect(t.Context(), &metrics))
		assert.NotEmpty(t, metrics.ScopeMetrics)
		// No need for more assertions on the metrics, it would be like re-testing the entire otelsql instrumentation.

		mock.ExpectClose()
		assert.NoError(t, Close(db)) // Should unregister the meter
	})

	t.Run("Close_invalid_db", func(t *testing.T) {
		db := &gorm.DB{}
		assert.NoError(t, Close(db))

		db.Statement = &gorm.Statement{}
		db.Config = &gorm.Config{}
		assert.NoError(t, Close(db))
	})
}
