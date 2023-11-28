package session

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/database"
)

type testKey struct{}

type testCommitter struct {
	gorm.ConnPool
	committed  bool
	rolledback bool
}

func (c *testCommitter) Commit() error {
	c.committed = true
	return nil
}

func (c *testCommitter) Rollback() error {
	c.rolledback = true
	return nil
}

func (c *testCommitter) BeginTx(_ context.Context, _ *sql.TxOptions) (gorm.ConnPool, error) {
	return c, nil
}

func TestGormSession(t *testing.T) {
	cfg := config.LoadDefault()

	t.Run("New", func(t *testing.T) {
		db, err := database.NewFromDialector(cfg, nil, tests.DummyDialector{})
		if !assert.NoError(t, err) {
			return
		}

		opts := &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  true,
		}
		session := GORM(db, opts)

		assert.Equal(t, Gorm{
			ctx:       context.Background(),
			db:        db,
			TxOptions: opts,
		}, session)
	})

	t.Run("Manual", func(t *testing.T) {
		db, err := database.NewFromDialector(cfg, nil, tests.DummyDialector{})
		if !assert.NoError(t, err) {
			return
		}
		committer := &testCommitter{}
		db.Statement.ConnPool = committer
		opts := &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  true,
		}
		session := GORM(db, opts)

		ctx := context.WithValue(context.Background(), testKey{}, "testvalue")
		tx := session.Begin(ctx)
		assert.NotEqual(t, session, tx)
		assert.Equal(t, opts, tx.(Gorm).TxOptions)

		assert.NoError(t, tx.Commit())
		assert.True(t, committer.committed)

		assert.NoError(t, tx.Rollback())
		assert.True(t, committer.rolledback)
	})

	t.Run("Transaction", func(t *testing.T) {
		db, err := database.NewFromDialector(cfg, nil, tests.DummyDialector{})
		if !assert.NoError(t, err) {
			return
		}
		committer := &testCommitter{}
		db.Statement.ConnPool = committer
		session := GORM(db, nil)

		var ctxValue any
		ctx := context.WithValue(context.Background(), testKey{}, "testvalue")
		err = session.Transaction(ctx, func(ctx context.Context) error {
			ctxValue = ctx.Value(testKey{})
			db := ctx.Value(dbKey{})
			assert.NotNil(t, db)
			_, ok := db.(*gorm.DB)
			assert.True(t, ok)
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "testvalue", ctxValue)
		assert.True(t, committer.committed)
		assert.False(t, committer.rolledback)
	})

	t.Run("TransactionError", func(t *testing.T) {
		db, err := database.NewFromDialector(cfg, nil, tests.DummyDialector{})
		if !assert.NoError(t, err) {
			return
		}
		committer := &testCommitter{}
		db.Statement.ConnPool = committer
		session := GORM(db, nil)

		var ctxValue any
		ctx := context.WithValue(context.Background(), testKey{}, "testvalue")
		err = session.Transaction(ctx, func(ctx context.Context) error {
			ctxValue = ctx.Value(testKey{})
			return fmt.Errorf("test err")
		})
		assert.Error(t, err)
		assert.Equal(t, fmt.Errorf("test err"), err)
		assert.Equal(t, "testvalue", ctxValue)
		assert.True(t, committer.rolledback)
		assert.False(t, committer.committed)
	})

	t.Run("DB", func(t *testing.T) {
		db, err := database.NewFromDialector(cfg, nil, tests.DummyDialector{})
		if !assert.NoError(t, err) {
			return
		}
		fallback := &gorm.DB{}

		cases := []struct {
			ctx    context.Context
			expect *gorm.DB
			desc   string
		}{
			{
				desc:   "missing_from_context",
				ctx:    context.Background(),
				expect: fallback,
			},
			{
				desc:   "fallback",
				ctx:    context.Background(),
				expect: fallback,
			},
			{
				desc:   "found",
				ctx:    context.WithValue(context.Background(), dbKey{}, db),
				expect: db,
			},
		}

		for _, c := range cases {
			c := c
			t.Run(c.desc, func(t *testing.T) {
				db := DB(c.ctx, fallback)
				assert.Equal(t, c.expect, db)
			})
		}
	})
}
