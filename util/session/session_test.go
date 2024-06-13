package session

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/utils/tests"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/database"
	"goyave.dev/goyave/v5/util/errors"
)

var (
	errAlreadyCommitted  = fmt.Errorf("transaction has already been committed")
	errAlreadyRolledBack = fmt.Errorf("transaction has already been rolled back")
)

type testKey struct{}

type testConnPool struct {
	gorm.ConnPool
	beginError      error
	testTxCommitter *testTxCommitter
}

func newTestConnPool() *testConnPool {
	return &testConnPool{
		testTxCommitter: &testTxCommitter{},
	}
}

func (c *testConnPool) BeginTx(_ context.Context, _ *sql.TxOptions) (gorm.ConnPool, error) {
	return c.testTxCommitter, c.beginError
}

type testTxCommitter struct {
	gorm.ConnPool
	commitError error
	committed   bool
	rolledback  bool
}

func (c *testTxCommitter) Commit() error {
	if c.committed {
		return errAlreadyCommitted
	}
	c.committed = true
	return c.commitError
}

func (c *testTxCommitter) Rollback() error {
	if c.rolledback {
		return errAlreadyRolledBack
	}
	c.rolledback = true
	return nil
}

type testDialector struct {
	tests.DummyDialector
	savepointErr error
	rollbackErr  error

	savepoint    string
	rolledbackTo string
}

func (d *testDialector) SavePoint(_ *gorm.DB, name string) error {
	d.savepoint = name
	return d.savepointErr
}

func (d *testDialector) RollbackTo(_ *gorm.DB, name string) error {
	d.rolledbackTo = name
	return d.rollbackErr
}

func TestGormSession(t *testing.T) {
	cfg := config.LoadDefault()
	cfg.Set("database.config.disableAutomaticPing", true)

	t.Run("New", func(t *testing.T) {
		db, err := database.NewFromDialector(cfg, nil, &testDialector{})
		require.NoError(t, err)

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
		db, err := database.NewFromDialector(cfg, nil, &testDialector{})
		require.NoError(t, err)
		committer := newTestConnPool()
		db.Statement.ConnPool = committer
		opts := &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  true,
		}
		session := GORM(db, opts)

		ctx := context.WithValue(context.Background(), testKey{}, "testvalue")
		tx, err := session.Begin(ctx)
		require.NoError(t, err)
		assert.NotEqual(t, session, tx)
		assert.Equal(t, opts, tx.(Gorm).TxOptions)
		assert.Equal(t, tx.(Gorm).ctx, tx.Context())
		assert.Equal(t, "testvalue", tx.Context().Value(testKey{}))
		assert.Equal(t, tx.(Gorm).db, tx.Context().Value(dbKey{}))

		require.NoError(t, tx.Commit())
		assert.True(t, committer.testTxCommitter.committed)

		require.NoError(t, tx.Rollback())
		assert.True(t, committer.testTxCommitter.rolledback)
	})

	t.Run("Begin_error", func(t *testing.T) {
		db, err := database.NewFromDialector(cfg, nil, &testDialector{})
		require.NoError(t, err)
		beginErr := fmt.Errorf("begin error")
		committer := &testConnPool{
			beginError:      beginErr,
			testTxCommitter: &testTxCommitter{},
		}
		db.Statement.ConnPool = committer
		session := GORM(db, nil)

		tx, err := session.Begin(context.Background())
		require.ErrorIs(t, err, beginErr)
		assert.Nil(t, tx)

		err = session.Transaction(context.Background(), func(_ context.Context) error {
			return nil
		})
		require.ErrorIs(t, err, beginErr)
	})

	t.Run("Nested_manual", func(t *testing.T) {
		dialector := &testDialector{}
		db, err := database.NewFromDialector(cfg, nil, dialector)
		require.NoError(t, err)
		committer := newTestConnPool()
		db.Statement.ConnPool = committer
		session := GORM(db, nil)

		ctx := context.WithValue(context.Background(), testKey{}, "testvalue")
		tx, err := session.Begin(ctx)
		tx.(Gorm).db.Statement.Clauses["testclause"] = clause.Clause{} // Use this to check the nested db is based on the parent DB
		require.NoError(t, err)
		assert.NotNil(t, tx)

		subtx, err := session.Begin(tx.Context())
		require.NoError(t, err)
		assert.NotEmpty(t, subtx.(Gorm).savepoint)
		assert.Equal(t, subtx.(Gorm).savepoint, dialector.savepoint)
		assert.Equal(t, "testvalue", subtx.(Gorm).db.Statement.Context.Value(testKey{})) // Parent context is kept
		assert.Contains(t, subtx.(Gorm).db.Statement.Clauses, "testclause")              // Parent DB is used

		// subtx commit is no-op
		assert.NoError(t, subtx.Commit())
		assert.False(t, committer.testTxCommitter.committed)
		assert.NoError(t, tx.Commit())
		assert.True(t, committer.testTxCommitter.committed)

		// subtx only rolls back to savepoint
		assert.NoError(t, subtx.Rollback())
		assert.False(t, committer.testTxCommitter.rolledback)
		assert.Equal(t, subtx.(Gorm).savepoint, dialector.rolledbackTo)
		assert.NoError(t, tx.Rollback())
		assert.True(t, committer.testTxCommitter.rolledback)
	})

	t.Run("Nested_manual_DisableNestedTransactions", func(t *testing.T) {
		dialector := &testDialector{}
		db, err := database.NewFromDialector(cfg, nil, dialector)
		db.DisableNestedTransaction = true
		require.NoError(t, err)
		committer := newTestConnPool()
		db.Statement.ConnPool = committer
		session := GORM(db, nil)

		ctx := context.WithValue(context.Background(), testKey{}, "testvalue")
		tx, err := session.Begin(ctx)
		tx.(Gorm).db.Statement.Clauses["testclause"] = clause.Clause{} // Use this to check the nested db is based on the parent DB
		require.NoError(t, err)
		assert.NotNil(t, tx)

		subtx, err := session.Begin(tx.Context())
		require.NoError(t, err)
		assert.NotEmpty(t, subtx.(Gorm).savepoint)
		assert.Equal(t, "testvalue", subtx.(Gorm).db.Statement.Context.Value(testKey{})) // Parent context is kept
		assert.Contains(t, subtx.(Gorm).db.Statement.Clauses, "testclause")              // Parent DB is used

		// no savepoint
		assert.Empty(t, dialector.savepoint)
		assert.NoError(t, subtx.Commit())
		assert.False(t, committer.testTxCommitter.committed) // subtx commit is no-op
		assert.NoError(t, tx.Commit())
		assert.True(t, committer.testTxCommitter.committed)

		assert.NoError(t, subtx.Rollback()) // subtx rollback is no-op because nested transactions are disabled
		assert.False(t, committer.testTxCommitter.rolledback)
		assert.Empty(t, dialector.rolledbackTo)
		assert.NoError(t, tx.Rollback())
		assert.True(t, committer.testTxCommitter.rolledback)
	})

	t.Run("Nested_Begin_error", func(t *testing.T) {
		savepointErr := fmt.Errorf("savepoint error")
		dialector := &testDialector{
			savepointErr: savepointErr,
		}
		db, err := database.NewFromDialector(cfg, nil, dialector)
		require.NoError(t, err)
		committer := newTestConnPool()
		db.Statement.ConnPool = committer
		session := GORM(db, nil)

		tx, err := session.Begin(context.Background())
		require.NoError(t, err)

		subtx, err := session.Begin(tx.Context())
		assert.ErrorIs(t, err, savepointErr)
		assert.Nil(t, subtx)
	})

	t.Run("Transaction", func(t *testing.T) {
		db, err := database.NewFromDialector(cfg, nil, &testDialector{})
		require.NoError(t, err)
		committer := newTestConnPool()
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
		require.NoError(t, err)
		assert.Equal(t, "testvalue", ctxValue)
		assert.True(t, committer.testTxCommitter.committed)
		assert.False(t, committer.testTxCommitter.rolledback)
	})

	t.Run("Nested_Transaction", func(t *testing.T) {
		dialector := &testDialector{}
		db, err := database.NewFromDialector(cfg, nil, dialector)
		require.NoError(t, err)
		committer := newTestConnPool()
		db.Statement.ConnPool = committer
		session := GORM(db, nil)

		ctx := context.WithValue(context.Background(), testKey{}, "testvalue")
		tx, err := session.Begin(ctx)
		tx.(Gorm).db.Statement.Clauses["testclause"] = clause.Clause{} // Use this to check the nested db is based on the parent DB
		require.NoError(t, err)
		assert.NotNil(t, tx)

		var savepoint string
		f := func(ctx context.Context) error {
			db := DB(ctx, nil)
			assert.Equal(t, savepoint, dialector.savepoint)
			assert.NotNil(t, db)
			assert.Contains(t, db.Statement.Clauses, "testclause") // Parent DB is used
			return nil
		}
		savepoint = fmt.Sprintf("sp%p", f)
		err = session.Transaction(tx.Context(), f)
		require.NoError(t, err)

		f = func(_ context.Context) error {
			return fmt.Errorf("rollback")
		}
		savepoint = fmt.Sprintf("sp%p", f)
		err = session.Transaction(tx.Context(), f)
		require.Error(t, err)
		assert.Equal(t, savepoint, dialector.rolledbackTo)
	})

	t.Run("Nested_Transaction_savepoint_err", func(t *testing.T) {
		dialector := &testDialector{
			savepointErr: fmt.Errorf("savepoint err"),
		}
		db, err := database.NewFromDialector(cfg, nil, dialector)
		require.NoError(t, err)
		committer := newTestConnPool()
		db.Statement.ConnPool = committer
		session := GORM(db, nil)

		ctx := context.WithValue(context.Background(), testKey{}, "testvalue")
		tx, err := session.Begin(ctx)
		require.NoError(t, err)

		err = session.Transaction(tx.Context(), func(_ context.Context) error {
			return nil
		})
		require.ErrorIs(t, err, dialector.savepointErr)
	})

	t.Run("Nested_Transaction_DisableNestedTransaction", func(t *testing.T) {
		dialector := &testDialector{}
		db, err := database.NewFromDialector(cfg, nil, dialector)
		db.DisableNestedTransaction = true
		require.NoError(t, err)
		committer := newTestConnPool()
		db.Statement.ConnPool = committer
		session := GORM(db, nil)

		ctx := context.WithValue(context.Background(), testKey{}, "testvalue")
		tx, err := session.Begin(ctx)
		tx.(Gorm).db.Statement.Clauses["testclause"] = clause.Clause{} // Use this to check the nested db is based on the parent DB
		require.NoError(t, err)
		assert.NotNil(t, tx)

		err = session.Transaction(tx.Context(), func(ctx context.Context) error {
			db := DB(ctx, nil)
			assert.Empty(t, dialector.savepoint)
			assert.NotNil(t, db)
			assert.Contains(t, db.Statement.Clauses, "testclause") // Parent DB is used
			return nil
		})
		require.NoError(t, err)

		err = session.Transaction(tx.Context(), func(_ context.Context) error {
			return fmt.Errorf("rollback")
		})
		require.Error(t, err)
		assert.Empty(t, dialector.rolledbackTo)
	})

	t.Run("TransactionError", func(t *testing.T) {
		db, err := database.NewFromDialector(cfg, nil, &testDialector{})
		require.NoError(t, err)
		committer := newTestConnPool()
		db.Statement.ConnPool = committer
		session := GORM(db, nil)

		var ctxValue any
		ctx := context.WithValue(context.Background(), testKey{}, "testvalue")
		err = session.Transaction(ctx, func(ctx context.Context) error {
			ctxValue = ctx.Value(testKey{})
			return fmt.Errorf("test err")
		})
		require.Error(t, err)
		assert.Equal(t, errors.New(fmt.Errorf("test err")).Error(), err.Error())
		assert.Equal(t, "testvalue", ctxValue)
		assert.True(t, committer.testTxCommitter.rolledback)
		assert.False(t, committer.testTxCommitter.committed)
	})

	t.Run("Transaction_Commit_error", func(t *testing.T) {
		db, err := database.NewFromDialector(cfg, nil, &testDialector{})
		require.NoError(t, err)
		commitErr := fmt.Errorf("commit error")
		committer := newTestConnPool()
		committer.testTxCommitter.commitError = commitErr
		db.Statement.ConnPool = committer
		session := GORM(db, nil)

		err = session.Transaction(context.Background(), func(_ context.Context) error {
			return nil
		})
		require.ErrorIs(t, err, commitErr)
	})

	t.Run("DB", func(t *testing.T) {
		db, err := database.NewFromDialector(cfg, nil, &testDialector{})
		require.NoError(t, err)
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
