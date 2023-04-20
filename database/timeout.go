package database

import (
	"context"
	"time"

	"gorm.io/gorm"
)

const (
	timeoutCallbackBeforeName = "goyave:timeout_before"
	timeoutCallbackAfterName  = "goyave:timeout_after"
)

type timeoutContext struct {
	context.Context
	cancel context.CancelFunc
}

// TimeoutPlugin GORM plugin adding a default timeout to SQL queries if none is applied
// on the statement already. It works by replacing the statement's context with a child
// context having the configured timeout. The context is replaced in a "before" callback
// on all GORM operations. In a "after" callback, the new context is canceled.
type TimeoutPlugin struct {
	Timeout time.Duration
}

// Name returns the name of the plugin
func (p *TimeoutPlugin) Name() string {
	return "goyave:timeout"
}

// Initialize registers the callbacks for all operations.
func (p *TimeoutPlugin) Initialize(db *gorm.DB) error {
	// TODO test it works well if hooks

	createCallback := db.Callback().Create()
	if err := createCallback.Before("*").Register(timeoutCallbackBeforeName, p.timeoutBefore); err != nil {
		return err
	}
	if err := createCallback.After("*").Register(timeoutCallbackAfterName, p.timeoutAfter); err != nil {
		return err
	}

	queryCallback := db.Callback().Query()
	if err := queryCallback.Before("*").Register(timeoutCallbackBeforeName, p.timeoutBefore); err != nil {
		return err
	}
	if err := queryCallback.After("*").Register(timeoutCallbackAfterName, p.timeoutAfter); err != nil {
		return err
	}

	deleteCallback := db.Callback().Delete()
	if err := deleteCallback.Before("*").Register(timeoutCallbackBeforeName, p.timeoutBefore); err != nil {
		return err
	}
	if err := deleteCallback.After("*").Register(timeoutCallbackAfterName, p.timeoutAfter); err != nil {
		return err
	}

	updateCallback := db.Callback().Update()
	if err := updateCallback.Before("*").Register(timeoutCallbackBeforeName, p.timeoutBefore); err != nil {
		return err
	}
	if err := updateCallback.After("*").Register(timeoutCallbackAfterName, p.timeoutAfter); err != nil {
		return err
	}

	rowCallback := db.Callback().Row()
	if err := rowCallback.Before("*").Register(timeoutCallbackBeforeName, p.timeoutBefore); err != nil {
		return err
	}
	if err := rowCallback.After("*").Register(timeoutCallbackAfterName, p.timeoutAfter); err != nil {
		return err
	}

	rawCallback := db.Callback().Raw()
	if err := rawCallback.Before("*").Register(timeoutCallbackBeforeName, p.timeoutBefore); err != nil {
		return err
	}
	return rawCallback.After("*").Register(timeoutCallbackAfterName, p.timeoutAfter)
}

func (p *TimeoutPlugin) timeoutBefore(db *gorm.DB) {
	if db.Statement.Context == nil {
		return
	}
	if _, hasDeadline := db.Statement.Context.Deadline(); hasDeadline {
		return
	}
	ctx, cancel := context.WithTimeout(db.Statement.Context, p.Timeout)
	db.Statement.Context = &timeoutContext{
		Context: ctx,
		cancel:  cancel,
	}
}

func (p *TimeoutPlugin) timeoutAfter(db *gorm.DB) {
	ctx, ok := db.Statement.Context.(*timeoutContext)
	if !ok || ctx.cancel == nil {
		return
	}
	ctx.cancel()
}
