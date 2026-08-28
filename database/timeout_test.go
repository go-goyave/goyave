package database

import (
	"context"
	"database/sql/driver"
	"fmt"
	"math"
	"regexp"
	"testing"
	"testing/synctest"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/config"
)

func prepareTimeoutTest(t *testing.T, timeout int) (*gorm.DB, sqlmock.Sqlmock) {
	cfg := &config.DatabaseConnection{
		Dialect:                    "sqlmock",
		DatabaseName:               fmt.Sprintf("timeout_test_%s.db", t.Name()),
		DefaultReadQueryTimeoutMs:  timeout,
		DefaultWriteQueryTimeoutMs: timeout,
		MaxIdleConnections:         1,             // TODO document this is important for tests overwise the mock connection gets closed
		GORM:                       config.GORM{}, // Disabling PrepareStmt is important to avoid errors caused by mock
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

	return db, mock
}

func TestTimeoutPlugin(t *testing.T) {
	t.Run("Callbacks", func(t *testing.T) {
		db, _ := prepareTimeoutTest(t, 5)

		callbacks := db.Callback()

		assert.NotNil(t, callbacks.Create().Get(timeoutCallbackBeforeName))
		assert.NotNil(t, callbacks.Create().Get(timeoutCallbackAfterName))

		assert.NotNil(t, callbacks.Query().Get(timeoutCallbackBeforeName))
		assert.NotNil(t, callbacks.Query().Get(timeoutCallbackAfterName))

		assert.NotNil(t, callbacks.Delete().Get(timeoutCallbackBeforeName))
		assert.NotNil(t, callbacks.Delete().Get(timeoutCallbackAfterName))

		assert.NotNil(t, callbacks.Update().Get(timeoutCallbackBeforeName))
		assert.NotNil(t, callbacks.Update().Get(timeoutCallbackAfterName))

		// assert.NotNil(t, callbacks.Row().Get(timeoutCallbackBeforeName))
		// assert.NotNil(t, callbacks.Row().Get(timeoutCallbackAfterName))

		assert.NotNil(t, callbacks.Raw().Get(timeoutCallbackBeforeName))
		assert.NotNil(t, callbacks.Raw().Get(timeoutCallbackAfterName))
	})

	t.Run("timeout", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			db, mock := prepareTimeoutTest(t, 5)

			mock.ExpectQuery(regexp.QuoteMeta("SELECT `id` FROM `test_users` WHERE `email` = ?")).
				WithArgs("johndoe@example.org").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)).
				WillDelayFor(time.Millisecond * 6)

			users := []*TestUser{}
			res := db.Select("id").Where("email", "johndoe@example.org").Find(&users)
			require.Error(t, res.Error)
			assert.ErrorIs(t, res.Error, sqlmock.ErrCancelled) // not context.DeadlineExceeded because sqlMock checks `<-ctx.Done()`
		})
	})

	t.Run("timeout_Exec", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			db, mock := prepareTimeoutTest(t, 5)

			mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET updated_at = NOW()")).
				WillDelayFor(time.Millisecond * 6).WillReturnResult(driver.ResultNoRows)

			res := db.Exec("UPDATE users SET updated_at = NOW()")
			require.Error(t, res.Error)
			assert.ErrorIs(t, res.Error, sqlmock.ErrCancelled) // not context.DeadlineExceeded because sqlMock checks `<-ctx.Done()`
		})
	})

	t.Run("timeout_Scan", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			db, mock := prepareTimeoutTest(t, 5)

			mock.ExpectQuery(regexp.QuoteMeta("SELECT `id` FROM `test_users` WHERE `email` = ?")).
				WithArgs("johndoe@example.org").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)).
				WillDelayFor(time.Millisecond * 6)

			users := []*TestUser{}
			res := db.Raw("SELECT `id` FROM `test_users` WHERE `email` = ?", "johndoe@example.org").Scan(&users)
			require.Error(t, res.Error)
			assert.ErrorIs(t, res.Error, sqlmock.ErrCancelled) // not context.DeadlineExceeded because sqlMock checks `<-ctx.Done()`
		})
	})

	// This test fails currently because: https://github.com/go-gorm/gorm/pull/7809#issuecomment-5452508771
	t.Run("no_timeout_Scan", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			db, mock := prepareTimeoutTest(t, 5)

			mock.ExpectQuery(regexp.QuoteMeta("SELECT `id` FROM `test_users` WHERE `email` = ?")).
				WithArgs("johndoe@example.org").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)).
				WillDelayFor(time.Millisecond * 6)

			users := []*TestUser{}
			res := db.Raw("SELECT `id` FROM `test_users` WHERE `email` = ?", "johndoe@example.org").Scan(&users)
			require.NoError(t, res.Error)
			want := []*TestUser{{ID: 1}}
			assert.Equal(t, want, users)
		})
	})

	t.Run("re-use_statement", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			// The statement is re-used for consecutive queries.
			// Each individual query in the transaction is supposed to have its own timeout
			// so if the two queries together exceed the configured timeout, there should be no error.
			db, mock := prepareTimeoutTest(t, 5)

			for range 2 {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT `id` FROM `test_users` WHERE `email` = ?")).
					WithArgs("johndoe@example.org").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)).
					WillDelayFor(time.Millisecond * 4)
			}

			users := []*TestUser{}
			db = db.Select("id").Where("email", "johndoe@example.org").Find(&users)
			require.NoError(t, db.Error)
			db = db.Find(&users)
			require.NoError(t, db.Error)
		})
	})

	t.Run("dont_override_predefined_context", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			db, mock := prepareTimeoutTest(t, 5)

			users := []*TestUser{}
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
			defer cancel()

			mock.ExpectQuery(regexp.QuoteMeta("SELECT `id` FROM `test_users` WHERE `email` = ?")).
				WithArgs("johndoe@example.org").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)).
				WillDelayFor(time.Millisecond * 9)

			// The context is replaced with a longer timeout so the query can be completed.
			res := db.WithContext(ctx).Select("id").Where("email", "johndoe@example.org").Find(&users)
			require.NoError(t, res.Error)
		})
	})

	t.Run("disabled", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			db, mock := prepareTimeoutTest(t, 0)

			mock.ExpectQuery(regexp.QuoteMeta("SELECT `id` FROM `test_users` WHERE `email` = ?")).
				WithArgs("johndoe@example.org").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)).
				WillDelayFor(math.MaxInt64)

			users := []*TestUser{}
			res := db.Select("id").Where("email", "johndoe@example.org").Find(&users)
			require.NoError(t, res.Error)
		})
	})

	t.Run("transaction_many_queries", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			db, mock := prepareTimeoutTest(t, 5)

			mock.ExpectBegin()
			for i := range 3 {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT `id` FROM `test_users` WHERE `email` = ?")).
					WithArgs(fmt.Sprintf("johndoe%d@example.org", i)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(i + 1)).
					WillDelayFor(time.Millisecond * 4)
			}
			mock.ExpectCommit()

			// Each individual query in the transaction is supposed to have its own timeout.
			// If we execute a lot of queries that take a cumulated time
			// superior to the configured timeout, we should have no error.
			err := db.Transaction(func(tx *gorm.DB) error {
				for i := range 3 {
					users := []*TestUser{}
					res := tx.Select("id").Where("email", fmt.Sprintf("johndoe%d@example.org", i)).Find(&users)
					if res.Error != nil {
						return res.Error
					}
				}
				return nil
			})
			require.NoError(t, err)
		})
	})

	t.Run("transaction_one_timeout", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			db, mock := prepareTimeoutTest(t, 5)

			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta("SELECT `id` FROM `test_users` WHERE `email` = ?")).
				WithArgs("johndoe0@example.org").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)).
				WillDelayFor(time.Millisecond * 3)
			mock.ExpectQuery(regexp.QuoteMeta("SELECT `id` FROM `test_users` WHERE `email` = ?")).
				WithArgs("johndoe1@example.org").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2)).
				WillDelayFor(time.Millisecond * 6)
			// Don't expect a third query since second will timeout.
			mock.ExpectRollback()

			err := db.Transaction(func(tx *gorm.DB) error {
				for i := range 3 {
					users := []*TestUser{}
					res := tx.Select("id").Where("email", fmt.Sprintf("johndoe%d@example.org", i)).Find(&users)
					if res.Error != nil {
						return res.Error
					}
				}
				return nil
			})
			require.ErrorIs(t, err, sqlmock.ErrCancelled) // not context.DeadlineExceeded because sqlMock checks `<-ctx.Done()`
		})
	})

	// TODO test Exec
	// TODO test Raw/Scan
}
