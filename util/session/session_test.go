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
	"goyave.dev/goyave/v5/util/errors"
)

type testKey struct{}

type testCommitter struct {
	gorm.ConnPool
	beginError  error
	commitError error
	committed   bool
	rolledback  bool
}

func (c *testCommitter) Commit() error {
	c.committed = true
	return c.commitError
}

func (c *testCommitter) Rollback() error {
	c.rolledback = true
	return nil
}

func (c *testCommitter) BeginTx(_ context.Context, _ *sql.TxOptions) (gorm.ConnPool, error) {
	return c, c.beginError
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
		tx, err := session.Begin(ctx)
		assert.NoError(t, err)
		assert.NotEqual(t, session, tx)
		assert.Equal(t, opts, tx.(Gorm).TxOptions)
		assert.Equal(t, tx.(Gorm).ctx, tx.Context())
		assert.Equal(t, "testvalue", tx.Context().Value(testKey{}))
		assert.Equal(t, tx.(Gorm).db, tx.Context().Value(dbKey{}))

		assert.NoError(t, tx.Commit())
		assert.True(t, committer.committed)

		assert.NoError(t, tx.Rollback())
		assert.True(t, committer.rolledback)
	})

	t.Run("Begin_error", func(t *testing.T) {
		db, err := database.NewFromDialector(cfg, nil, tests.DummyDialector{})
		if !assert.NoError(t, err) {
			return
		}
		beginErr := fmt.Errorf("begin error")
		committer := &testCommitter{
			beginError: beginErr,
		}
		db.Statement.ConnPool = committer
		session := GORM(db, nil)

		tx, err := session.Begin(context.Background())
		assert.ErrorIs(t, err, beginErr)
		assert.Nil(t, tx)

		err = session.Transaction(context.Background(), func(ctx context.Context) error {
			return nil
		})
		assert.ErrorIs(t, err, beginErr)
	})

	t.Run("Nested_manual", func(t *testing.T) {
		db, err := database.NewFromDialector(cfg, nil, tests.DummyDialector{})
		if !assert.NoError(t, err) {
			return
		}
		committer := &testCommitter{}
		db.Statement.ConnPool = committer
		session := GORM(db, nil)

		ctx := context.WithValue(context.Background(), testKey{}, "testvalue")
		tx, err := session.Begin(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, tx)

		subtx, err := session.Begin(tx.Context())
		assert.NoError(t, err)
		assert.Equal(t, "testvalue", subtx.(Gorm).db.Statement.Context.Value(testKey{})) // Parent context is kept
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
		assert.Equal(t, errors.New(fmt.Errorf("test err")).Error(), err.Error())
		assert.Equal(t, "testvalue", ctxValue)
		assert.True(t, committer.rolledback)
		assert.False(t, committer.committed)
	})

	t.Run("Transaction_Commit_error", func(t *testing.T) {
		db, err := database.NewFromDialector(cfg, nil, tests.DummyDialector{})
		if !assert.NoError(t, err) {
			return
		}
		commitErr := fmt.Errorf("commit error")
		committer := &testCommitter{
			commitError: commitErr,
		}
		db.Statement.ConnPool = committer
		session := GORM(db, nil)

		err = session.Transaction(context.Background(), func(ctx context.Context) error {
			return nil
		})
		assert.ErrorIs(t, err, commitErr)
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
