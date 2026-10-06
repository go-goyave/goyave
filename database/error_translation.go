package database

import (
	"errors"

	"gorm.io/gorm"
	"goyave.dev/goyave/v6"
	"goyave.dev/goyave/v6/util/errwrap"
)

const (
	errorTranslationCallbackAfterName = "goyave:error_translation_after"
)

// ErrorTranslationPlugin GORM plugin joining [gorm.ErrRecordNotFound] errors with
// [goyave.NotFound] errors for compatibility with the rest of the framework.
type ErrorTranslationPlugin struct{}

// Name returns the name of the plugin
func (p *ErrorTranslationPlugin) Name() string {
	return "goyave:error_translation"
}

// Initialize registers the callbacks for all operations.
func (p *ErrorTranslationPlugin) Initialize(db *gorm.DB) error {
	createCallback := db.Callback().Create()
	if err := createCallback.After("*").Register(errorTranslationCallbackAfterName, p.translate); err != nil {
		return errwrap.New(err)
	}

	queryCallback := db.Callback().Query()
	if err := queryCallback.After("*").Register(errorTranslationCallbackAfterName, p.translate); err != nil {
		return errwrap.New(err)
	}

	deleteCallback := db.Callback().Delete()
	if err := deleteCallback.After("*").Register(errorTranslationCallbackAfterName, p.translate); err != nil {
		return errwrap.New(err)
	}

	updateCallback := db.Callback().Update()
	if err := updateCallback.After("*").Register(errorTranslationCallbackAfterName, p.translate); err != nil {
		return errwrap.New(err)
	}

	rowCallback := db.Callback().Row()
	if err := rowCallback.After("*").Register(errorTranslationCallbackAfterName, p.translate); err != nil {
		return errwrap.New(err)
	}

	rawCallback := db.Callback().Raw()
	if err := rawCallback.After("*").Register(errorTranslationCallbackAfterName, p.translate); err != nil {
		return errwrap.New(err)
	}
	return nil
}

func (p *ErrorTranslationPlugin) translate(db *gorm.DB) {
	if errors.Is(db.Error, gorm.ErrRecordNotFound) {
		db.Error = errors.Join(goyave.ErrNotFound, db.Error)
	}
}
