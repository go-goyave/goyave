package database

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"
	"uuid"

	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"goyave.dev/goyave/v5/util/errors"
)

// clickhouseTypes mapping of Go types to Clickhouse types can be found here:
// https://github.com/ClickHouse/clickhouse-go/blob/main/TYPES.md
// Go types uint and int are not specified, default to 'UInt64' and 'Int64', respectively.
// For clickhouse, if the key is a type that is not part of this list, a [Exist.Transform] is required.
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

// Exist database helper to check the existence of records.
//
// This helper is for convenience: it provides a one-liner method to use
// in repositories rather than having to implement the query yourself.
//
// The T type is the type of the column used for checking existence.
type Exist[T any] struct {
	// Transform if provided, this function is called on query parameters to transform
	// them into a raw expression. For example to transform a number into `(123::int)` for
	// Postgres to prevent some type errors when using "CheckSlice" due to the underlying
	// driver not being able to automatically detect the typing.
	Transform func(val T) clause.Expr
	Table     string
	Column    string
}

// NewExist create a database helper to check the existence of records.
// Often used in repositories wired to [validation.ExistsValidator].
//
// The T type is the type of the column used for checking existence.
func NewExist[T any](table, column string) Exist[T] {
	return Exist[T]{
		Table:  table,
		Column: column,
	}
}

// Check if a single record exists using a count query.
func (e Exist[T]) Check(db *gorm.DB, key T) (bool, error) {
	exists, err := e.check(db, key)
	if err != nil {
		return false, errors.New(err)
	}
	return exists, nil
}

func (e Exist[T]) check(db *gorm.DB, key T) (bool, error) {
	count := int64(0)

	var transformedValue any = key
	if e.Transform != nil {
		transformedValue = e.Transform(key)
	}
	err := db.Table(e.Table).Where(e.Column, transformedValue).Count(&count).Error
	if err != nil {
		return false, errors.New(err)
	}
	return count != 0, nil
}

// CheckSlice check if all given keys have an associated record in database.
// A single query is executed.
//
// Returns the indexes of the keys that didn't match any existing record.
// For example, if the second and third "keys" elements do not match, the
// returned slice will be `[]int{1,2}`.
func (e Exist[T]) CheckSlice(db *gorm.DB, keys []T) ([]int, error) {
	return e.checkSlice(db, keys, true)
}

func (e Exist[T]) checkSlice(db *gorm.DB, keys []T, exist bool) ([]int, error) {
	query, err := e.buildSliceQuery(db, keys, exist)
	if err != nil {
		return nil, errors.New(err)
	}

	var results []int
	if err := query.Find(&results).Error; err != nil {
		return nil, errors.New(err)
	}
	return results, nil
}

// buildSliceQuery creates a raw query to check if a slice of elements exists in the database.
// The query returns the indexes of elements that don't exist.
//
// If `exist` is true, check for existence and returns the indexes of the elements that don't exist.
//
// If `exist` is false, it inverts the condition in order to check that the elements in the slice don't exist.
// This is essentially a uniqueness validation. The query will return the indexes of the elements that exist.
func (e Exist[T]) buildSliceQuery(db *gorm.DB, keys []T, exist bool) (*gorm.DB, error) {
	questionMarks := []string{}
	params := []any{}

	dialectorName := db.Name()

	if dialectorName == "clickhouse" {
		return e.buildClickhouseSliceQuery(db, keys, exist)
	}

	isPostgres := dialectorName == "postgres"

	for i, key := range keys {
		questionMarks = append(questionMarks, "?")
		var transformedValue any = key
		if e.Transform != nil {
			transformedValue = e.Transform(key)
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

	table := db.Statement.Quote(e.Table)
	column := db.Statement.Quote(e.Column)

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
		lo.Ternary(exist, "", "NOT"),
	)
	return db.Raw(sql, params...), nil
}

func (e Exist[T]) buildClickhouseSliceQuery(db *gorm.DB, values []T, condition bool) (*gorm.DB, error) {
	questionMarks := []string{}
	params := []any{}

	var zeroVal T
	paramType, ok := clickhouseTypes[reflect.TypeOf(zeroVal)]
	if !ok && e.Transform == nil {
		return nil, errors.Errorf("database.Exist: value of type T (%T) is not supported for Clickhouse. You must provide a Transform function", zeroVal)
	}

	for i, val := range values {
		questionMarks = append(questionMarks, "?")
		var transformedValue any = val
		if e.Transform != nil {
			transformedValue = e.Transform(val)
		}
		params = append(params, gorm.Expr("(?,?)", transformedValue, i))
	}

	table := db.Statement.Quote(e.Table)
	column := db.Statement.Quote(e.Column)

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

// Unique database helper to check if an input doesn't match any existing record
// in the database.
//
// This helper is for convenience: it provides a one-liner method to use
// in repositories rather than having to implement the query yourself.
//
// The T type is the type of the column used for checking existence.
type Unique[T any] struct {
	Exist[T]
}

// NewUnique create a database helper to check the uniqueness of an input.
// In other words, it checks that no record in the database match an input.
// Often used in repositories wired to [validation.UniqueValidator].
//
// The T type is the type of the column used for checking existence.
func NewUnique[T any](table, column string) Unique[T] {
	return Unique[T]{
		Table:  table,
		Column: column,
	}
}

// Check if the given key doesn't match any existing record using a count query.
func (u Unique[T]) Check(db *gorm.DB, key T) (bool, error) {
	exists, err := u.check(db, key)
	if err != nil {
		return false, errors.New(err)
	}
	return !exists, nil
}

// CheckSlice check if all given keys do not match any record in database.
// A single query is executed.
//
// Returns the indexes of the keys that matched an existing record.
// For example, if the second and third "keys" elements match an existing record,
// the returned slice will be `[]int{1,2}`.
func (u Unique[T]) CheckSlice(db *gorm.DB, keys []T) ([]int, error) {
	return u.checkSlice(db, keys, false)
}
