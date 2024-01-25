package session

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
	"goyave.dev/goyave/v5/util/errors"
)

// Session aims at facilitating business transactions while abstracting the underlying mechanism,
// be it a database transaction or another transaction mechanism. This allows services to execute
// multiple business use-cases and easily rollback changes in case of error, without creating a
// dependency to the database layer.
//
// Sessions should be constituted of a root session created with a "New"-type constructor and allow
// the creation of child sessions with `Begin()` and `Transaction()`. Nested transactions should be supported
// as well.
type Session interface {
	// Begin returns a new session with the given context and a started transaction.
	// Using the returned session should have no side-effect on the parent session.
	Begin(ctx context.Context) (Session, error)

	// Transaction executes a transaction. If the given function returns an error, the transaction
	// is rolled back. Otherwise it is automatically committed before `Transaction()` returns.
	Transaction(ctx context.Context, f func(context.Context) error) error

	// Rollback the changes in the transaction. This action is final.
	Rollback() error

	// Commit the changes in the transaction. This action is final.
	Commit() error
}

// Gorm session implementation.
type Gorm struct {
	db        *gorm.DB
	TxOptions *sql.TxOptions
	ctx       context.Context
}

// GORM create a new root session for Gorm.
// The transaction options are optional.
func GORM(db *gorm.DB, opt *sql.TxOptions) Gorm {
	return Gorm{
		db:        db,
		TxOptions: opt,
		ctx:       context.Background(),
	}
}

// Begin returns a new session with the given context and a started DB transaction.
// The returned session has manual controls. Make sure a call to `Rollback()` or `Commit()`
// is executed before the session is expired (eligible for garbage collection).
func (s Gorm) Begin(ctx context.Context) (Session, error) {
	tx := s.db.WithContext(ctx).Begin(s.TxOptions)
	return Gorm{
		ctx:       ctx,
		TxOptions: s.TxOptions,
		db:        tx,
	}, errors.NewSkip(tx.Error, 3)
}

// Rollback the changes in the transaction. This action is final.
func (s Gorm) Rollback() error {
	return errors.NewSkip(s.db.Rollback().Error, 3)
}

// Commit the changes in the transaction. This action is final.
func (s Gorm) Commit() error {
	return errors.NewSkip(s.db.Commit().Error, 3)
}

// dbKey the key used to store the database in the context.
type dbKey struct{}

// Transaction executes a transaction. If the given function returns an error, the transaction
// is rolled back. Otherwise it is automatically committed before `Transaction()` returns.
//
// The Gorm DB associated with this session is injected into the context as a value so `session.DB()`
// can be used to retrieve it.
func (s Gorm) Transaction(ctx context.Context, f func(context.Context) error) error {
	tx := s.db.WithContext(ctx).Begin(s.TxOptions)
	if tx.Error != nil {
		return errors.New(tx.Error)
	}
	c := context.WithValue(ctx, dbKey{}, tx)
	err := f(c)
	if err != nil {
		tx.Rollback()
		return errors.New(err)
	}
	err = tx.Commit().Error
	if err != nil {
		return errors.New(err)
	}
	return nil
}

// DB returns the Gorm instance stored in the given context. Returns the given fallback
// if no Gorm DB could be found in the context.
func DB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	db := ctx.Value(dbKey{})
	if db == nil {
		return fallback
	}
	return db.(*gorm.DB)
}
