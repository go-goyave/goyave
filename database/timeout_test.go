package database

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/config"
)

func prepareTimeoutTest(dbName string) *gorm.DB {
	cfg := config.LoadDefault()
	cfg.Set("app.debug", false)
	cfg.Set("database.connection", "sqlite3_timeout_test")
	cfg.Set("database.name", fmt.Sprintf("timeout_test_%s.db", dbName))
	cfg.Set("database.options", "mode=memory")
	cfg.Set("database.defaultReadQueryTimeout", 5)
	cfg.Set("database.defaultWriteQueryTimeout", 5)
	db, err := New(cfg, nil)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	if err := db.Session(&gorm.Session{NewDB: true, Context: ctx}).AutoMigrate(&TestUser{}); err != nil {
		cancel()
		panic(err)
	}
	cancel()

	author := userGenerator()
	if err := db.Create(author).Error; err != nil {
		panic(err)
	}
	return db
}

func TestTimeoutPlugin(t *testing.T) {
	RegisterDialect("sqlite3_timeout_test", "file:{name}?{options}", sqlite.Open)
	t.Cleanup(func() {
		mu.Lock()
		delete(dialects, "sqlite3_timeout_test")
		mu.Unlock()
	})

	t.Run("Callbacks", func(t *testing.T) {
		db := prepareTimeoutTest(t.Name())

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
		db := prepareTimeoutTest(t.Name())

		// Generate a huge WHERE condition to artificially make the query very long
		args := lo.RepeatBy(20000, func(index int) string {
			return fmt.Sprintf("foobar_%d@example.org", index)
		})

		users := []*TestUser{}
		res := db.Select("*").Where("email IN (?)", args).Find(&users)
		require.Error(t, res.Error)
		assert.Equal(t, "context deadline exceeded", res.Error.Error())
	})

	t.Run("re-use_statement", func(t *testing.T) {
		db := prepareTimeoutTest(t.Name())

		users := []*TestUser{}
		db = db.Select("*").Where("email", "johndoe@example.org").Find(&users)
		require.NoError(t, db.Error)
		db = db.Select("*").Where("email", "johndoe@example.org").Find(&users)
		require.NoError(t, db.Error)
	})

	t.Run("dont_override_predefined_context", func(t *testing.T) {
		db := prepareTimeoutTest(t.Name())

		// Generate a huge WHERE condition to artificially make the query very long
		args := lo.RepeatBy(20000, func(index int) string {
			return fmt.Sprintf("foobar_%d@example.org", index)
		})

		users := []*TestUser{}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// The context is replaced with a longer timeout so the query can be completed.
		res := db.WithContext(ctx).Select("*").Where("email IN (?)", args).Find(&users)
		require.NoError(t, res.Error)
	})

	t.Run("disabled", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("app.debug", false)
		cfg.Set("database.connection", "sqlite3_timeout_test")
		cfg.Set("database.name", "timeout_test_disabled.db")
		cfg.Set("database.options", "mode=memory")
		cfg.Set("database.defaultReadQueryTimeout", 0)
		cfg.Set("database.defaultWriteQueryTimeout", 0)
		db, err := New(cfg, nil)
		if err != nil {
			panic(err)
		}

		if err := db.AutoMigrate(&TestUser{}); err != nil {
			panic(err)
		}

		// Generate a huge WHERE condition to artificially make the query very long
		args := lo.RepeatBy(20000, func(index int) string {
			return fmt.Sprintf("foobar_%d@example.org", index)
		})

		users := []*TestUser{}
		res := db.Select("*").Where("email IN (?)", args).Find(&users)
		require.NoError(t, res.Error)
	})

	t.Run("transaction_many_queries", func(t *testing.T) {
		t.Cleanup(func() {
			// The DB is not in memory here
			if err := os.Remove("timeout_many_queries_test.db"); err != nil {
				panic(err)
			}
		})
		cfg := config.LoadDefault()
		cfg.Set("app.debug", false)
		cfg.Set("database.connection", "sqlite3_timeout_test")
		cfg.Set("database.name", "timeout_many_queries_test.db")
		cfg.Set("database.defaultReadQueryTimeout", 200)
		cfg.Set("database.defaultWriteQueryTimeout", 200)
		db, err := New(cfg, nil)
		if err != nil {
			panic(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
		if err := db.WithContext(ctx).AutoMigrate(&TestUser{}); err != nil {
			panic(err)
		}
		defer cancel()

		author := userGenerator()
		if err := db.Create(author).Error; err != nil {
			panic(err)
		}

		// The timeout should be per query
		// If we execute a lot of queries that take a cumulated time
		// superior to the configured timeout, we should have no error.

		// Generate a huge WHERE condition to artificially make the query long
		args := lo.RepeatBy(1000, func(index int) string {
			return fmt.Sprintf("foobar_%d@example.org", index)
		})
		err = db.Transaction(func(_ *gorm.DB) error {
			for i := 0; i < 5000; i++ {
				users := []*TestUser{}
				res := db.Select("*").Where("email IN (?)", args).Find(&users)
				if res.Error != nil {
					return res.Error
				}
			}
			return nil
		})
		require.NoError(t, err)
	})
}
