package testutil

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCtxKey struct{}

func TestSession(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		session := NewTestSession()
		assert.Equal(t, context.Background(), session.Context())
		assert.Nil(t, session.cancel)
		assert.NotNil(t, session.children)
		assert.Empty(t, session.children)
		assert.Equal(t, SessionCreated, session.Status())
	})

	t.Run("manual_commit", func(t *testing.T) {
		session := NewTestSession()
		ctx := context.WithValue(context.Background(), testCtxKey{}, "test-value")
		child, err := session.Begin(ctx)
		require.NoError(t, err)
		assert.Equal(t, SessionCreated, session.Status()) // Parent session status unchanged

		// Check the child session has been correctly created
		childSession, ok := child.(*Session)
		require.True(t, ok)
		assert.Equal(t, SessionCreated, childSession.Status())
		assert.NotNil(t, childSession.cancel)

		// Uses the parent context
		childCtx := childSession.Context()
		assert.NotEqual(t, ctx, childCtx)
		assert.Equal(t, "test-value", childCtx.Value(testCtxKey{}))
		assert.Equal(t, child, childCtx.Value(ctxKey{}))

		err = childSession.Commit()
		require.NoError(t, err)
		assert.Equal(t, SessionCommitted, childSession.Status())
		assert.ErrorIs(t, childCtx.Err(), context.Canceled) // Make sure the child context has been canceled

		assert.NoError(t, session.Context().Err())        // The parent context should not be canceled
		assert.Equal(t, SessionCreated, session.Status()) // Parent session status unchanged
	})

	t.Run("manual_rollback", func(t *testing.T) {
		session := NewTestSession()
		ctx := context.WithValue(context.Background(), testCtxKey{}, "test-value")
		child, err := session.Begin(ctx)
		require.NoError(t, err)

		childSession, ok := child.(*Session)
		require.True(t, ok)
		childCtx := childSession.Context()

		err = childSession.Rollback()
		require.NoError(t, err)
		assert.Equal(t, SessionRolledBack, childSession.Status())
		assert.ErrorIs(t, childCtx.Err(), context.Canceled) // Make sure the child context has been canceled

		assert.NoError(t, session.Context().Err())        // The parent context should not be canceled
		assert.Equal(t, SessionCreated, session.Status()) // Parent session status unchanged
	})

	t.Run("manual_children_added", func(t *testing.T) {
		session := NewTestSession()
		child1, err := session.Begin(context.Background())
		require.NoError(t, err)
		child2, err := session.Begin(context.Background())
		require.NoError(t, err)

		assert.Equal(t, []*Session{child1.(*Session), child2.(*Session)}, session.Children())

		nested, err := child1.Begin(child1.Context())
		require.NoError(t, err)
		assert.Equal(t, []*Session{nested.(*Session)}, child1.(*Session).Children())

		nested2, err := session.Begin(child1.Context()) // Parent is child1 because of the context, not root session
		require.NoError(t, err)
		assert.Equal(t, []*Session{nested.(*Session), nested2.(*Session)}, child1.(*Session).Children())
		assert.Equal(t, []*Session{child1.(*Session), child2.(*Session)}, session.Children()) // Root session children unchanged
	})

	t.Run("full_tx_children_added", func(t *testing.T) {
		session := NewTestSession()
		ctx := context.WithValue(context.Background(), testCtxKey{}, "test-value")
		for range 2 {
			err := session.Transaction(ctx, func(_ context.Context) error {
				return nil
			})
			require.NoError(t, err)
		}

		// Both children were added to the children slice
		// and both should be committed.
		children := session.Children()
		assert.Len(t, children, 2)
		for _, c := range children {
			assert.Equal(t, SessionCommitted, c.Status())

			// Uses the parent context
			childCtx := c.Context()
			assert.NotEqual(t, ctx, childCtx)
			assert.Equal(t, "test-value", childCtx.Value(testCtxKey{}))
			assert.ErrorIs(t, childCtx.Err(), context.Canceled) // Make sure the child context has been canceled
		}
		assert.NoError(t, session.Context().Err())        // The parent context should not be canceled
		assert.Equal(t, SessionCreated, session.Status()) // Parent session status unchanged
	})

	t.Run("full_tx_rollback", func(t *testing.T) {
		testError := errors.New("test error")
		session := NewTestSession()
		err := session.Transaction(context.Background(), func(_ context.Context) error {
			return testError
		})
		require.ErrorIs(t, err, testError)
		assert.NotEqual(t, testError, err) // Error should be wrapped

		// Children should be rolled back.
		children := session.Children()
		assert.Len(t, children, 1)
		for _, c := range children {
			assert.Equal(t, SessionRolledBack, c.Status())
			assert.ErrorIs(t, c.Context().Err(), context.Canceled) // Make sure the child context has been canceled
		}

		assert.NoError(t, session.Context().Err())        // The parent context should not be canceled
		assert.Equal(t, SessionCreated, session.Status()) // Parent session status unchanged
	})

	t.Run("cannot_commit_root_session", func(t *testing.T) {
		session := NewTestSession()
		assert.ErrorIs(t, session.Commit(), ErrEndRootSession)
	})

	t.Run("cannot_rollback_root_session", func(t *testing.T) {
		session := NewTestSession()
		assert.ErrorIs(t, session.Rollback(), ErrEndRootSession)
	})

	t.Run("cannot_begin_from_ended_session", func(t *testing.T) {
		session := NewTestSession()
		child, err := session.Begin(context.Background())
		require.NoError(t, err)
		require.NoError(t, child.Commit())

		c, err := child.Begin(context.Background())
		assert.Nil(t, c)
		assert.ErrorIs(t, err, ErrSessionEnded)
	})

	t.Run("cannot_commit_committed_session", func(t *testing.T) {
		session := NewTestSession()
		child, err := session.Begin(context.Background())
		require.NoError(t, err)
		require.NoError(t, child.Commit())
		assert.ErrorIs(t, child.Commit(), ErrSessionEnded)
		assert.Equal(t, SessionCommitted, child.(*Session).Status())
	})

	t.Run("cannot_commit_rolledback_session", func(t *testing.T) {
		session := NewTestSession()
		child, err := session.Begin(context.Background())
		require.NoError(t, err)
		require.NoError(t, child.Rollback())
		assert.ErrorIs(t, child.Commit(), ErrSessionEnded)
		assert.Equal(t, SessionRolledBack, child.(*Session).Status())
	})

	t.Run("cannot_rollback_rolledback_session", func(t *testing.T) {
		session := NewTestSession()
		child, err := session.Begin(context.Background())
		require.NoError(t, err)
		require.NoError(t, child.Rollback())
		assert.ErrorIs(t, child.Rollback(), ErrSessionEnded)
		assert.Equal(t, SessionRolledBack, child.(*Session).Status())
	})

	t.Run("cannot_rollback_committed_session", func(t *testing.T) {
		session := NewTestSession()
		child, err := session.Begin(context.Background())
		require.NoError(t, err)
		require.NoError(t, child.Commit())
		assert.ErrorIs(t, child.Rollback(), ErrSessionEnded)
		assert.Equal(t, SessionCommitted, child.(*Session).Status())
	})

	t.Run("cannot_commit_if_child_is_running", func(t *testing.T) {
		session := NewTestSession()
		mainSession, err := session.Begin(context.Background())
		require.NoError(t, err)

		committedChild, err := mainSession.Begin(mainSession.Context())
		require.NoError(t, err)
		require.NoError(t, committedChild.Commit())
		rolledBackchild, err := mainSession.Begin(mainSession.Context())
		require.NoError(t, err)
		require.NoError(t, rolledBackchild.Rollback())

		_, err = mainSession.Begin(mainSession.Context()) // Running child
		require.NoError(t, err)

		err = mainSession.Commit()
		require.ErrorIs(t, err, ErrChildRunning)
		assert.Equal(t, SessionCreated, mainSession.(*Session).Status()) // Status unchanged
	})

	t.Run("cannot_rollback_if_child_is_running", func(t *testing.T) {
		session := NewTestSession()
		mainSession, err := session.Begin(context.Background())
		require.NoError(t, err)

		committedChild, err := mainSession.Begin(mainSession.Context())
		require.NoError(t, err)
		require.NoError(t, committedChild.Commit())
		rolledBackchild, err := mainSession.Begin(mainSession.Context())
		require.NoError(t, err)
		require.NoError(t, rolledBackchild.Rollback())

		_, err = mainSession.Begin(mainSession.Context()) // Running child
		require.NoError(t, err)

		err = mainSession.Rollback()
		require.ErrorIs(t, err, ErrChildRunning)
		assert.Equal(t, SessionCreated, mainSession.(*Session).Status()) // Status unchanged
	})

	t.Run("cannot_begin_if_parent_context_canceled", func(t *testing.T) {
		session := NewTestSession()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := session.Begin(ctx)
		require.ErrorIs(t, err, context.Canceled)

		ctx, cancel = context.WithCancel(context.Background())
		mainSession, err := session.Begin(ctx)
		require.NoError(t, err)
		cancel()
		_, err = mainSession.Begin(context.WithValue(ctx, testCtxKey{}, "test-value"))
		require.ErrorIs(t, err, context.Canceled)
		require.NotErrorIs(t, err, ErrSessionEnded)
	})

	t.Run("cannot_full_tx_if_parent_context_canceled", func(t *testing.T) {
		session := NewTestSession()
		ctx, cancel := context.WithCancel(context.Background())
		mainSession, err := session.Begin(ctx)
		require.NoError(t, err)
		cancel()
		err = mainSession.Transaction(context.WithValue(ctx, testCtxKey{}, "test-value"), func(_ context.Context) error {
			return nil
		})
		require.ErrorIs(t, err, context.Canceled)
		require.NotErrorIs(t, err, ErrSessionEnded)
	})

	t.Run("cannot_commit_if_parent_context_canceled", func(t *testing.T) {
		session := NewTestSession()
		ctx, cancel := context.WithCancel(context.Background())
		mainSession, err := session.Begin(ctx)
		require.NoError(t, err)
		cancel()
		err = mainSession.Commit()
		require.ErrorIs(t, err, context.Canceled)
		require.NotErrorIs(t, err, ErrSessionEnded)
	})

	t.Run("can_rollback_if_parent_context_canceled", func(t *testing.T) {
		session := NewTestSession()
		ctx, cancel := context.WithCancel(context.Background())
		mainSession, err := session.Begin(ctx)
		require.NoError(t, err)
		cancel()
		err = mainSession.Rollback()
		require.NoError(t, err)
	})

	t.Run("begin_context_should_be_from_parent_session", func(t *testing.T) {
		session := NewTestSession()
		mainSession, err := session.Begin(context.Background())
		require.NoError(t, err)

		_, err = mainSession.Begin(context.Background())
		require.ErrorIs(t, err, ErrNotParentContext)

		ctx := context.WithValue(mainSession.Context(), testCtxKey{}, "test-value")
		_, err = mainSession.Begin(ctx)
		require.NoError(t, err)
	})
}
