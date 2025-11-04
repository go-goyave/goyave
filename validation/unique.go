package validation

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"goyave.dev/goyave/v5/util/errors"
)

// Mapping of Go types to Clickhouse types can be found here:
// https://github.com/ClickHouse/clickhouse-go/blob/main/TYPES.md
// Go types uint and int are not specified, default to 'UInt64' and 'Int64', respectively
var clickhouseTypes = map[reflect.Type]string{
	reflect.TypeFor[uint64]():    "UInt64",
	reflect.TypeFor[uint32]():    "UInt32",
	reflect.TypeFor[uint16]():    "UInt16",
	reflect.TypeFor[uint8]():     "UInt8",
	reflect.TypeFor[uint]():      "UInt64",
	reflect.TypeFor[int64]():     "Int64",
	reflect.TypeFor[int32]():     "Int32",
	reflect.TypeFor[int16]():     "Int16",
	reflect.TypeFor[int8]():      "Int8",
	reflect.TypeFor[int]():       "Int64",
	reflect.TypeFor[float32]():   "Float32",
	reflect.TypeFor[float64]():   "Float64",
	reflect.TypeFor[string]():    "String",
	reflect.TypeFor[bool]():      "Bool",
	reflect.TypeFor[uuid.UUID](): "UUID",
	reflect.TypeFor[time.Time](): "DateTime64",
	reflect.TypeFor[*big.Int]():  "Int256",
}

// UniqueValidator validates the field under validation must have a unique value in database
// according to the provided database scope. Uniqueness is checked using a COUNT query.
type UniqueValidator struct {
	Scope func(db *gorm.DB, val any) *gorm.DB // TODO v6: change val to validation.Context
	BaseValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *UniqueValidator) Validate(ctx *Context) bool {
	if ctx.Invalid {
		return true
	}
	count := int64(0)

	if err := v.Scope(v.DB(), ctx.Value).Count(&count).Error; err != nil {
		ctx.AddError(errors.New(err))
		return false
	}
	return count == 0
}

// Name returns the string name of the validator.
func (v *UniqueValidator) Name() string { return "unique" }

// Unique validates the field under validation must have a unique value in database
// according to the provided database scope. Uniqueness is checked using a COUNT query.
//
//	 v.Unique(func(db *gorm.DB, val any) *gorm.DB {
//		return db.Model(&model.User{}).Where(clause.PrimaryKey, val)
//	 })
//
//	 v.Unique(func(db *gorm.DB, val any) *gorm.DB {
//		// Unique email excluding the currently authenticated user
//		return db.Model(&model.User{}).Where("email", val).Where("email != ?", request.User.(*model.User).Email)
//	 })
func Unique(scope func(db *gorm.DB, val any) *gorm.DB) *UniqueValidator {
	return &UniqueValidator{Scope: scope}
}

//------------------------------

// ExistsValidator validates the field under validation must exist in database
// according to the provided database scope. Existence is checked using a COUNT query.
type ExistsValidator struct {
	UniqueValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *ExistsValidator) Validate(ctx *Context) bool {
	return !v.UniqueValidator.Validate(ctx)
}

// Name returns the string name of the validator.
func (v *ExistsValidator) Name() string { return "exists" }

// Exists validates the field under validation must have exist database
// according to the provided database scope. Existence is checked using a COUNT query.
//
//	 v.Exists(func(db *gorm.DB, val any) *gorm.DB {
//		return db.Model(&model.User{}).Where(clause.PrimaryKey, val)
//	 })
func Exists(scope func(db *gorm.DB, val any) *gorm.DB) *ExistsValidator {
	return &ExistsValidator{UniqueValidator: UniqueValidator{Scope: scope}}
}

//------------------------------

// ExistsArrayValidator validates the field under validation must be an array and all
// of its elements must exist. The type `T` is the type of the elements of the array
// under validation.
//
// This is preferable to use this validation rule on the array instead of `Exists` on
// each array element because this rule will only execute a single SQL query instead of
// as many as there are elements in the array.
//
// If provided, the `Transform` function is called on every array element to transform
// them into a raw expression. For example to transform a number into `(123::int)` for
// Postgres to prevent some type errors.
type ExistsArrayValidator[T any] struct {
	BaseValidator
	Transform func(val T) clause.Expr
	Table     string
	Column    string
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *ExistsArrayValidator[T]) Validate(ctx *Context) bool {
	return v.validate(ctx, true)
}

func (v *ExistsArrayValidator[T]) buildQuery(values []T, condition bool) (*gorm.DB, error) {
	questionMarks := []string{}
	params := []any{}

	db := v.DB()
	dialectorName := db.Name()

	if dialectorName == "clickhouse" {
		return v.buildClickhouseQuery(values, condition)
	}

	isPostgres := dialectorName == "postgres"

	for i, val := range values {
		questionMarks = append(questionMarks, "?")
		var transformedValue any = val
		if v.Transform != nil {
			transformedValue = v.Transform(val)
		}
		switch dialectorName {
		case "mysql":
			params = append(params, gorm.Expr("ROW(?,?)", transformedValue, i))
		case "bigquery":
			params = append(params, gorm.Expr("STRUCT(? AS id, ? AS i)", transformedValue, i))
		default:
			params = append(params, gorm.Expr(
				"(?,?)",
				transformedValue,
				lo.Ternary[any](isPostgres, gorm.Expr("?::int", i), i),
			))
		}
	}

	table := db.Statement.Quote(v.Table)
	column := db.Statement.Quote(v.Column)

	var cte string
	switch dialectorName {
	case "bigquery":
		cte = fmt.Sprintf("WITH ctx_values AS (SELECT * FROM UNNEST([%s]))", strings.Join(questionMarks, ","))
	default:
		cte = fmt.Sprintf("WITH ctx_values(id, i) AS (SELECT * FROM (VALUES %s) t%s)",
			strings.Join(questionMarks, ","),
			lo.Ternary(dialectorName == "sqlserver", "(id,i)", ""),
		)
	}

	sql := fmt.Sprintf(
		"%s SELECT i FROM ctx_values LEFT JOIN %s ON %s.%s = ctx_values.id WHERE %s.%s IS %s NULL",
		cte,
		table,
		table, column,
		table, column,
		lo.Ternary(condition, "", "NOT"),
	)
	return db.Raw(sql, params...), nil
}

func (v *ExistsArrayValidator[T]) buildClickhouseQuery(values []T, condition bool) (*gorm.DB, error) {
	questionMarks := []string{}
	params := []any{}

	var zeroVal T
	paramType, ok := clickhouseTypes[reflect.TypeOf(zeroVal)]
	if !ok && v.Transform == nil {
		return nil, errors.Errorf("ExistsArray/UniqueArray validator: value of type T (%T) is not supported for Clickhouse. You must provide a Transform function", zeroVal)
	}

	for i, val := range values {
		questionMarks = append(questionMarks, "?")
		var transformedValue any = val
		if v.Transform != nil {
			transformedValue = v.Transform(val)
		}
		params = append(params, gorm.Expr("(?,?)", transformedValue, i))
	}

	db := v.DB()
	table := db.Statement.Quote(v.Table)
	column := db.Statement.Quote(v.Column)

	sql := fmt.Sprintf(
		"WITH ctx_values(id, i) AS (SELECT * FROM (VALUES 'id %s, i Int64', %s)) SELECT i FROM ctx_values INNER JOIN %s ON %s.%s = ctx_values.id WHERE %s.%s IS %s NULL",
		paramType,
		strings.Join(questionMarks, ","),
		table,
		table, column,
		table, column,
		lo.Ternary(condition, "", "NOT"),
	)
	return db.Raw(sql, params...), nil
}

func (v *ExistsArrayValidator[T]) validate(ctx *Context, condition bool) bool {
	values, ok := ctx.Value.([]T)
	if ctx.Invalid || !ok {
		return true
	}

	db, err := v.buildQuery(values, condition)
	if err != nil {
		ctx.AddError(errors.New(err))
		return false
	}

	timeout := v.Config().GetInt("database.defaultReadQueryTimeout")
	if _, hasDeadline := db.Statement.Context.Deadline(); !hasDeadline && timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(db.Statement.Context, time.Duration(timeout)*time.Millisecond)
		defer cancel()
		db = db.WithContext(timeoutCtx)
	}

	results := []int{}
	if err := db.Find(&results).Error; err != nil {
		ctx.AddError(errors.New(err))
		return false
	}

	ctx.AddArrayElementValidationErrors(results...)
	return true
}

// Name returns the string name of the validator.
func (v *ExistsArrayValidator[T]) Name() string { return "exists" }

// ExistsArray validates the field under validation must be an array and all
// of its elements must exist. The type `T` is the type of the elements of the array
// under validation.
//
// This is preferable to use this validation rule on the array instead of `Exists` on
// each array element because this rule will only execute a single SQL query instead of
// as many as there are elements in the array.
//
// If provided, the `Transform` function is called on every array element to transform
// them into a raw expression. For example to transform a number into `(123::int)` for
// Postgres to prevent some type errors.
func ExistsArray[T any](table, column string, transform func(val T) clause.Expr) *ExistsArrayValidator[T] {
	return &ExistsArrayValidator[T]{
		Table:     table,
		Column:    column,
		Transform: transform,
	}
}

//------------------------------

// UniqueArrayValidator validates the field under validation must be an array and all
// of its elements must not already exist. The type `T` is the type of the elements of the array
// under validation.
//
// This is preferable to use this validation rule on the array instead of `Unique` on
// each array element because this rule will only execute a single SQL query instead of
// as many as there are elements in the array.
//
// If provided, the `Transform` function is called on every array element to transform
// them into a raw expression. For example to transform a number into `(123::int)` for
// Postgres to prevent some type errors.
type UniqueArrayValidator[T any] struct {
	ExistsArrayValidator[T]
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *UniqueArrayValidator[T]) Validate(ctx *Context) bool {
	return v.validate(ctx, false)
}

// Name returns the string name of the validator.
func (v *UniqueArrayValidator[T]) Name() string { return "unique" }

// UniqueArray validates the field under validation must be an array and all
// of its elements must not already exist. The type `T` is the type of the elements of the array
// under validation.
//
// This is preferable to use this validation rule on the array instead of `Unique` on
// each array element because this rule will only execute a single SQL query instead of
// as many as there are elements in the array.
//
// If provided, the `Transform` function is called on every array element to transform
// them into a raw expression. For example to transform a number into `(123::int)` for
// Postgres to prevent some type errors.
func UniqueArray[T any](table, column string, transform func(val T) clause.Expr) *UniqueArrayValidator[T] {
	return &UniqueArrayValidator[T]{
		ExistsArrayValidator: ExistsArrayValidator[T]{
			Table:     table,
			Column:    column,
			Transform: transform,
		},
	}
}
