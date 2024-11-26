package testutil

import (
	"context"
	stderrors "errors"
	"sync/atomic"

	"github.com/samber/lo"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/session"
)

const (
	SessionCreated uint32 = iota
	SessionCommitted
	SessionRolledBack
)

var (
	ErrSessionEnded     = stderrors.New("testutil.Session: session already ended")
	ErrEndRootSession   = stderrors.New("testutil.Session: cannot commit/rollback root session")
	ErrChildRunning     = stderrors.New("testutil.Session: cannot commit/rollback if a child session is still running")
	ErrNotParentContext = stderrors.New("testutil.Session: cannot create a child session with an unrelated context. Parent context should be the context or a child context of the parent session.")
)

// ctxKey the key used to store the `*Session` into a context value.
type ctxKey struct{}

// Session is an advanced mock for the `session.Session` interface. This implementation is designed to
// provide a realistic, observable transaction system and help identify incorrect usage.
//
// Each transaction created with this implementation has a cancellable context created from its parent.
// The context is canceled when the session is committed or rolled back. This helps detecting cases where
// your code tries to use a terminated transaction.
//
// A transaction cannot be committed or rolled back several times. It cannot be committed after being rolled back
// or the other way around.
//
// For nested transactions, all child sessions should be ended (committed or rolled back) before the parent can be ended.
// Moreover, the context given on `Begin()` should be the context or a child context of the parent session.
//
// A child session cannot be created or committed if its parent context is done.
//
// The root transaction cannot be committed or rolledback. This helps detecting cases where your codes
// uses the root session without creating a child session.
//
// This implementation is not meant to be used concurrently. You should create a new instance for each test.
type Session struct {
	ctx      context.Context
	cancel   func()
	children []*Session
	status   atomic.Uint32
}

// NewTestSession create a new root session with the `context.Background()`.
func NewTestSession() *Session {
	return &Session{
		ctx:      context.Background(),
		cancel:   nil,
		children: []*Session{},
		status:   atomic.Uint32{},
	}
}

// Begin returns a new child session with the given context.
func (s *Session) Begin(ctx context.Context) (session.Session, error) {
	if s.status.Load() != SessionCreated {
		return nil, ErrSessionEnded
	}
	if err := ctx.Err(); err != nil {
		return nil, errors.New(err)
	}

	parent := ctx.Value(ctxKey{})
	if s.cancel != nil && parent != s {
		// Not the root session, we are creating a nested transaction
		// The given context should belong to the parent session.
		return nil, errors.New(ErrNotParentContext)
	}

	childCtx, cancel := context.WithCancel(ctx)
	tx := &Session{
		cancel:   cancel,
		children: []*Session{},
		status:   atomic.Uint32{},
	}
	tx.ctx = context.WithValue(childCtx, ctxKey{}, tx)
	if parent != nil {
		parentSession := parent.(*Session)
		parentSession.children = append(parentSession.children, tx)
	} else {
		s.children = append(s.children, tx)
	}
	return tx, nil
}

// Transaction executes a transaction. If the given function returns an error, the transaction
// is rolled back. Otherwise it is automatically committed before `Transaction()` returns.
// The underlying transaction mechanism is injected into the context as a value.
func (s *Session) Transaction(ctx context.Context, f func(context.Context) error) error {
	tx, err := s.Begin(ctx)
	if err != nil {
		return errors.New(err)
	}

	err = errors.New(f(tx.Context()))
	if err != nil {
		rollbackErr := errors.New(tx.Rollback())
		return errors.New([]error{err, rollbackErr})
	}
	return errors.New(tx.Commit())
}

// Rollback the transaction. For this test utility, it only sets the sessions status to `SessionRollbedBack`.
// If the session status is not `testutil.SessionCreated`, returns `testutil.ErrSessionEnded`.
//
// It is not possible to roll back the root session. In this case, `testutil.ErrEndRootSession` is returned.
//
// This action is final.
func (s *Session) Rollback() error {
	if s.cancel == nil {
		return errors.New(ErrEndRootSession)
	}
	if s.hasRunningChild() {
		return errors.New(ErrChildRunning)
	}
	swapped := s.status.CompareAndSwap(SessionCreated, SessionRolledBack)
	if !swapped {
		return errors.New(ErrSessionEnded)
	}
	s.cancel()
	return nil
}

// Commit the transaction. For this test utility, it only sets the sessions status to `SessionCommitted`.
// If the session status is not `testutil.SessionCreated`, returns `testutil.ErrSessionEnded`.
//
// It is not possible to commit the root session. In this case, `testutil.ErrEndRootSession` is returned.
//
// This action is final.
func (s *Session) Commit() error {
	if s.cancel == nil {
		return errors.New(ErrEndRootSession)
	}
	if err := s.ctx.Err(); err != nil {
		if s.Status() != SessionCreated {
			return errors.New([]error{err, ErrSessionEnded})
		}
		return errors.New(err)
	}
	if s.hasRunningChild() {
		return errors.New(ErrChildRunning)
	}
	swapped := s.status.CompareAndSwap(SessionCreated, SessionCommitted)
	if !swapped {
		return errors.New(ErrSessionEnded)
	}
	s.cancel()
	return nil
}

// Context returns the session's context.
func (s *Session) Context() context.Context {
	return s.ctx
}

// Status returns the session status. The value will be equal to `testutil.SessionCreated`,
// `testutil.SessionCommitted` or `testutil.SessionRolledBack`.
func (s *Session) Status() uint32 {
	return s.status.Load()
}

// Children returns the direct child sessions. You can use the returned values for your test assertions.
// The returned values are sorted in the order in which the child transactions were started.
//
// To access nested transactions, call `Children()` on the returned values.
//
// This method always returns a non-nil value but can return an empty slice.
func (s *Session) Children() []*Session {
	return s.children
}

func (s *Session) hasRunningChild() bool {
	return lo.SomeBy(s.children, func(child *Session) bool {
		return child.Status() == SessionCreated
	})
}
