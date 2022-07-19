package validation

import (
	"time"

	"gorm.io/gorm"
	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/lang"
)

type Options struct {
	Data      any
	Rules     Ruler
	IsJSON    bool
	Languages *lang.Languages
	DB        *gorm.DB
	Config    *config.Config
	Lang      string
	Extra     map[string]any
}

type ContextV5 struct {
	Options *Options
	Data    any
	Extra   map[string]any
	Value   any
	Parent  any
	Field   *Field
	Rule    *Rule
	Now     time.Time

	// The name of the field under validation
	Name string

	errors []error
}

// DB get the database instance given through the validation Options.
// Panics if there is none.
func (c *ContextV5) DB() *gorm.DB {
	if c.Options.DB == nil {
		panic("DB is not set in validation options")
	}
	return c.Options.DB
}

// Config get the configuration given through the validation Options.
// Panics if there is none.
func (c *ContextV5) Config() *config.Config {
	if c.Options.Config == nil {
		panic("Config is not set in validation options")
	}
	return c.Options.Config
}

// AddError adds an error to the validation context. This is NOT supposed
// to be used when the field under validation doesn't match the rule, but rather
// when there has been an operation error (such as a database error).
func (c *ContextV5) AddError(err error) {
	c.errors = append(c.errors, err)
}

func ValidateV5(options *Options) (Errors, []error) {
	// TODO implement validateV5

	return nil, nil
}
