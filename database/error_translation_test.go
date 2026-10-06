package database

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"goyave.dev/goyave/v6"
)

func prepareErrorTranslationTest(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	cfg := &Config{
		Dialect:                    "sqlmock",
		DatabaseName:               fmt.Sprintf("errtranslate_test_%s.db", t.Name()),
		DefaultReadQueryTimeoutMs:  0, // Disable timeout for this test
		DefaultWriteQueryTimeoutMs: 0,
		MaxIdleConnections:         1,
		Debug:                      false,
		GORM:                       GORMConfig{}, // Disabling PrepareStmt is important to avoid errors caused by mock
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

	db, err := NewFromDialector(cfg, dialector)
	if err != nil {
		require.NoError(t, err)
	}

	return db, mock
}

func TestErrorTransaltionPlugin(t *testing.T) {
	t.Run("Callbacks", func(t *testing.T) {
		db, _ := prepareErrorTranslationTest(t)

		callbacks := db.Callback()

		assert.NotNil(t, callbacks.Create().Get(errorTranslationCallbackAfterName))
		assert.NotNil(t, callbacks.Query().Get(errorTranslationCallbackAfterName))
		assert.NotNil(t, callbacks.Delete().Get(errorTranslationCallbackAfterName))
		assert.NotNil(t, callbacks.Update().Get(errorTranslationCallbackAfterName))
		assert.NotNil(t, callbacks.Row().Get(errorTranslationCallbackAfterName))
		assert.NotNil(t, callbacks.Raw().Get(errorTranslationCallbackAfterName))
	})

	t.Run("not_found_error_translated", func(t *testing.T) {
		db, mock := prepareErrorTranslationTest(t)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT `id` FROM `test_users` WHERE `email` = ? ORDER BY `test_users`.`id` LIMIT 1")).
			WithArgs("johndoe@example.org").
			WillReturnError(gorm.ErrRecordNotFound)

		user := &TestUser{}
		err := db.Select("id").Where("email", "johndoe@example.org").First(&user).Error
		assert.ErrorIs(t, err, goyave.ErrNotFound)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}
