package database

import (
	"context"
	"time"

	"gorm.io/gorm"
	"goyave.dev/goyave/v5/util/errors"
)

const (
	timeoutCallbackBeforeName = "goyave:timeout_before"
	timeoutCallbackAfterName  = "goyave:timeout_after"
)

type timeoutContext struct {
	context.Context

	parentContext context.Context

	// We store the pointer to the original statement
	// so we can cancel the context only if the original
	// statement is completely finished. This prevents
	// sub-statements (such as preloads) to cancel the context
	// when they are done, despite the parent statement not being
	// executed yet.
	statement *gorm.Statement

	cancel context.CancelFunc
}

// TimeoutPlugin GORM plugin adding a default timeout to SQL queries if none is applied
// on the statement already. It works by replacing the statement's context with a child
// context having the configured timeout. The context is replaced in a "before" callback
// on all GORM operations. In a "after" callback, the new context is canceled.
//
// The `ReadTimeout` is applied on the `Query` and `Raw` GORM callbacks. The `WriteTimeout`
// is applied on the rest of the callbacks.
//
// Supports all GORM operations except `Scan()`.
//
// A timeout duration inferior or equal to 0 disables the plugin for the relevant operations.
type TimeoutPlugin struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Name returns the name of the plugin
func (p *TimeoutPlugin) Name() string {
	return "goyave:timeout"
}

// Initialize registers the callbacks for all operations.
func (p *TimeoutPlugin) Initialize(db *gorm.DB) error {
	createCallback := db.Callback().Create()
	if err := createCallback.Before("*").Register(timeoutCallbackBeforeName, p.writeTimeoutBefore); err != nil {
		return errors.New(err)
	}
	if err := createCallback.After("*").Register(timeoutCallbackAfterName, p.timeoutAfter); err != nil {
		return errors.New(err)
	}

	queryCallback := db.Callback().Query()
	if err := queryCallback.Before("*").Register(timeoutCallbackBeforeName, p.readTimeoutBefore); err != nil {
		return errors.New(err)
	}
	if err := queryCallback.After("*").Register(timeoutCallbackAfterName, p.timeoutAfter); err != nil {
		return errors.New(err)
	}

	deleteCallback := db.Callback().Delete()
	if err := deleteCallback.Before("*").Register(timeoutCallbackBeforeName, p.writeTimeoutBefore); err != nil {
		return errors.New(err)
	}
	if err := deleteCallback.After("*").Register(timeoutCallbackAfterName, p.timeoutAfter); err != nil {
		return errors.New(err)
	}

	updateCallback := db.Callback().Update()
	if err := updateCallback.Before("*").Register(timeoutCallbackBeforeName, p.writeTimeoutBefore); err != nil {
		return errors.New(err)
	}
	if err := updateCallback.After("*").Register(timeoutCallbackAfterName, p.timeoutAfter); err != nil {
		return errors.New(err)
	}

	// Cannot use it with `Row()` because context is canceled before the call of `rows.Next()`, causing an error.
	// rowCallback := db.Callback().Row()
	// if err := rowCallback.Before("*").Register(timeoutCallbackBeforeName, p.readTimeoutBefore); err != nil {
	// 	return errors.New(err)
	// }
	// if err := rowCallback.After("*").Register(timeoutCallbackAfterName, p.timeoutAfter); err != nil {
	// 	return errors.New(err)
	// }

	rawCallback := db.Callback().Raw()
	if err := rawCallback.Before("*").Register(timeoutCallbackBeforeName, p.writeTimeoutBefore); err != nil {
		return errors.New(err)
	}
	if err := rawCallback.After("*").Register(timeoutCallbackAfterName, p.timeoutAfter); err != nil {
		return errors.New(err)
	}
	return nil
}

func (p *TimeoutPlugin) readTimeoutBefore(db *gorm.DB) {
	p.timeoutBefore(db, p.ReadTimeout)
}

func (p *TimeoutPlugin) writeTimeoutBefore(db *gorm.DB) {
	p.timeoutBefore(db, p.WriteTimeout)
}

func (p *TimeoutPlugin) timeoutBefore(db *gorm.DB, timeout time.Duration) {
	if timeout <= 0 || db.Statement.Context == nil {
		return
	}
	if tc, ok := db.Statement.Context.(*timeoutContext); ok {
		// The statement is re-used, replace the context with a new one
		ctx, cancel := context.WithTimeout(tc.parentContext, timeout)
		db.Statement.Context = &timeoutContext{
			Context:       ctx,
			parentContext: tc.parentContext,
			statement:     db.Statement,
			cancel:        cancel,
		}
		return
	}
	if _, hasDeadline := db.Statement.Context.Deadline(); hasDeadline {
		return
	}
	ctx, cancel := context.WithTimeout(db.Statement.Context, timeout)
	db.Statement.Context = &timeoutContext{
		Context:       ctx,
		parentContext: db.Statement.Context,
		statement:     db.Statement,
		cancel:        cancel,
	}
}

func (p *TimeoutPlugin) timeoutAfter(db *gorm.DB) {
	ctx, ok := db.Statement.Context.(*timeoutContext)
	if !ok || ctx.cancel == nil || db.Statement != ctx.statement {
		return
	}
	ctx.cancel()
}
