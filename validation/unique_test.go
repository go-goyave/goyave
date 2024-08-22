package validation

import (
	"fmt"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/database"
)

type uniqueTestModel struct {
	Name string
	ID   int64
}

func (m uniqueTestModel) TableName() string {
	return "models"
}

func prepareUniqueTest(t *testing.T) *Options {
	dialect := fmt.Sprintf("sqlite3_%s_test", t.Name())
	database.RegisterDialect(dialect, "file:{name}?{options}", sqlite.Open)
	cfg := config.LoadDefault()
	cfg.Set("app.name", t.Name())
	cfg.Set("app.debug", false)
	cfg.Set("database.connection", dialect)
	cfg.Set("database.name", dialect+".db")
	cfg.Set("database.options", "mode=memory")

	db, err := database.New(cfg, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = db.AutoMigrate(&uniqueTestModel{})
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	return &Options{
		DB:     db,
		Config: cfg,
	}
}

func TestUniqueValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Unique(func(db *gorm.DB, _ any) *gorm.DB {
			return db
		})
		assert.NotNil(t, v)
		assert.Equal(t, "unique", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
		assert.NotNil(t, v.Scope)
	})

	cases := []struct {
		value          any
		desc           string
		table          string
		column         string
		records        []uniqueTestModel
		expectedErrors []string
		valid          bool
		expected       bool
	}{
		{
			desc:           "OK",
			value:          "johndoe",
			table:          "models",
			column:         "name",
			records:        []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 3, Name: "c"}},
			valid:          true,
			expected:       true,
			expectedErrors: []string{},
		},
		{
			desc:           "NOK",
			value:          "johndoe",
			table:          "models",
			column:         "name",
			records:        []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "johndoe"}, {ID: 3, Name: "c"}},
			valid:          true,
			expected:       false,
			expectedErrors: []string{},
		},
		{
			desc:           "error",
			value:          "johndoe",
			table:          "models",
			column:         "not_a_column",
			records:        []uniqueTestModel{{ID: 1, Name: "a"}},
			valid:          true,
			expected:       false,
			expectedErrors: []string{"no such column: not_a_column"},
		},
		{
			desc:           "ctx_invalid",
			value:          "johndoe",
			table:          "models",
			column:         "name",
			records:        []uniqueTestModel{{ID: 1, Name: "johndoe"}},
			valid:          false,
			expected:       true,
			expectedErrors: []string{},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			opts := prepareUniqueTest(t)
			if len(c.records) > 0 {
				if err := opts.DB.Create(c.records).Error; err != nil {
					assert.FailNow(t, err.Error())
				}
			}

			v := Unique(func(db *gorm.DB, val any) *gorm.DB {
				return db.Model(&uniqueTestModel{}).Where(c.column, val)
			})
			v.init(opts)

			ctx := &Context{
				Invalid: !c.valid,
				Value:   c.value,
			}
			assert.Equal(t, c.expected, v.Validate(ctx))
			assert.Equal(t, c.expectedErrors, lo.Map(ctx.errors, func(e error, _ int) string { return e.Error() }))
		})
	}
}

func TestExistsValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Exists(func(db *gorm.DB, _ any) *gorm.DB {
			return db
		})
		assert.NotNil(t, v)
		assert.Equal(t, "exists", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
		assert.NotNil(t, v.Scope)
	})

	cases := []struct {
		value          any
		desc           string
		table          string
		column         string
		records        []uniqueTestModel
		expectedErrors []string
		valid          bool
		expected       bool
	}{
		{
			desc:           "OK",
			value:          "johndoe",
			table:          "models",
			column:         "name",
			records:        []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "johndoe"}, {ID: 3, Name: "c"}},
			valid:          true,
			expected:       true,
			expectedErrors: []string{},
		},
		{
			desc:           "NOK",
			value:          "johndoe",
			table:          "models",
			column:         "name",
			records:        []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 3, Name: "c"}},
			valid:          true,
			expected:       false,
			expectedErrors: []string{},
		},
		{
			desc:           "error",
			value:          "johndoe",
			table:          "models",
			column:         "not_a_column",
			records:        []uniqueTestModel{{ID: 1, Name: "johndoe"}},
			valid:          true,
			expected:       true, // It doesn't matter that it's true because the validator will exit with errors
			expectedErrors: []string{"no such column: not_a_column"},
		},
		{
			desc:           "ctx_invalid",
			value:          "johndoe",
			table:          "models",
			column:         "name",
			records:        []uniqueTestModel{{ID: 1, Name: "johndoe"}},
			valid:          false,
			expected:       false,
			expectedErrors: []string{},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			opts := prepareUniqueTest(t)
			if len(c.records) > 0 {
				if err := opts.DB.Create(c.records).Error; err != nil {
					assert.FailNow(t, err.Error())
				}
			}

			v := Exists(func(db *gorm.DB, val any) *gorm.DB {
				return db.Model(&uniqueTestModel{}).Where(c.column, val)
			})
			v.init(opts)

			ctx := &Context{
				Invalid: !c.valid,
				Value:   c.value,
			}
			assert.Equal(t, c.expected, v.Validate(ctx))
			assert.Equal(t, c.expectedErrors, lo.Map(ctx.errors, func(e error, _ int) string { return e.Error() }))
		})
	}
}

func TestUniqueArrayValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := UniqueArray[int]("table", "column", nil)
		assert.NotNil(t, v)
		assert.Equal(t, "unique", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
		assert.Equal(t, "table", v.Table)
		assert.Equal(t, "column", v.Column)
	})

	cases := []struct {
		value                      any
		desc                       string
		table                      string
		column                     string
		transform                  func(val int) clause.Expr
		records                    []uniqueTestModel
		expectedErrors             []string
		expectedArrayElementErrors []int
		valid                      bool
		expected                   bool
	}{
		{
			desc:           "OK",
			value:          []int{7, 5},
			table:          "models",
			column:         "id",
			records:        []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 3, Name: "c"}},
			valid:          true,
			transform:      nil,
			expected:       true,
			expectedErrors: []string{},
		},
		{

			desc:    "OK_with_transform",
			value:   []int{7, 5},
			table:   "models",
			column:  "id",
			records: []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 3, Name: "c"}},
			valid:   true,
			transform: func(val int) clause.Expr {
				return gorm.Expr("?", val-1)
			},
			expected:       true,
			expectedErrors: []string{},
		},
		{
			desc:    "NOK_with_transform",
			value:   []int{7, 4},
			table:   "models",
			column:  "id",
			records: []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 3, Name: "c"}},
			valid:   true,
			transform: func(val int) clause.Expr {
				return gorm.Expr("?", val-1)
			},
			expected:                   true, // Always returns true, the errors are added to child elements
			expectedErrors:             []string{},
			expectedArrayElementErrors: []int{1},
		},
		{
			desc:                       "NOK",
			value:                      []int{3, 5, 2},
			table:                      "models",
			column:                     "id",
			records:                    []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 3, Name: "c"}},
			valid:                      true,
			transform:                  nil,
			expected:                   true, // Always returns true, the errors are added to child elements
			expectedErrors:             []string{},
			expectedArrayElementErrors: []int{0, 2},
		},
		{
			desc:           "not_a_slice_of_int",
			value:          []string{"7", "5"},
			table:          "models",
			column:         "id",
			records:        []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 3, Name: "c"}},
			valid:          true,
			transform:      nil,
			expected:       true,
			expectedErrors: []string{},
		},
		{
			desc:           "ctx_invalid",
			value:          []int{7, 5},
			table:          "models",
			column:         "id",
			records:        []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 3, Name: "c"}},
			valid:          false,
			transform:      nil,
			expected:       true,
			expectedErrors: []string{},
		},
		{
			desc:           "error",
			value:          []int{7, 5},
			table:          "models",
			column:         "not_a_column",
			records:        []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 3, Name: "c"}},
			valid:          true,
			transform:      nil,
			expected:       false,
			expectedErrors: []string{"no such column: models.not_a_column"},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			opts := prepareUniqueTest(t)
			if len(c.records) > 0 {
				if err := opts.DB.Create(c.records).Error; err != nil {
					assert.FailNow(t, err.Error())
				}
			}

			v := UniqueArray[int](c.table, c.column, c.transform)
			v.init(opts)

			ctx := &Context{
				Invalid: !c.valid,
				Value:   c.value,
			}
			assert.Equal(t, c.expected, v.Validate(ctx))
			assert.Equal(t, c.expectedErrors, lo.Map(ctx.errors, func(e error, _ int) string { return e.Error() }))
			assert.Equal(t, c.expectedArrayElementErrors, ctx.arrayElementErrors)
		})
	}

	t.Run("buildQuery", func(t *testing.T) {
		cases := []struct {
			dialect  string
			expected string
		}{
			{dialect: "sqlite3", expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (2,0),(7,1),(6,2)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS NOT NULL"},
			{dialect: "mysql", expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES ROW(2,0),ROW(7,1),ROW(6,2)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS NOT NULL"},
			{dialect: "postgres", expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (2,0::int),(7,1::int),(6,2::int)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS NOT NULL"},
			{dialect: "mssql", expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (2,0),(7,1),(6,2)) t(id,i)) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS NOT NULL"},
			{dialect: "clickhouse", expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES 'id Int64, i Int64', (2,0),(7,1),(6,2))) SELECT i FROM ctx_values INNER JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS NOT NULL"},
		}

		for _, c := range cases {
			t.Run(c.dialect, func(t *testing.T) {
				opts := prepareUniqueTest(t)
				opts.Config.Set("database.connection", c.dialect)
				v := UniqueArray[int]("models", "name", nil)
				v.init(opts)

				tx, _ := v.buildQuery([]int{2, 7, 6}, false)

				sql := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx
				})
				assert.Equal(t, c.expected, sql)
			})
		}
	})
}

func TestExistsArrayValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := ExistsArray[int]("table", "column", nil)
		assert.NotNil(t, v)
		assert.Equal(t, "exists", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
		assert.Equal(t, "table", v.Table)
		assert.Equal(t, "column", v.Column)
	})

	cases := []struct {
		value                      any
		desc                       string
		table                      string
		column                     string
		transform                  func(val int) clause.Expr
		records                    []uniqueTestModel
		expectedErrors             []string
		expectedArrayElementErrors []int
		valid                      bool
		expected                   bool
	}{
		{
			desc:           "OK",
			value:          []int{2, 1},
			table:          "models",
			column:         "id",
			records:        []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 3, Name: "c"}},
			valid:          true,
			transform:      nil,
			expected:       true,
			expectedErrors: []string{},
		},
		{

			desc:    "OK_with_transform",
			value:   []int{2, 4},
			table:   "models",
			column:  "id",
			records: []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 3, Name: "c"}},
			valid:   true,
			transform: func(val int) clause.Expr {
				return gorm.Expr("?", val-1)
			},
			expected:       true,
			expectedErrors: []string{},
		},
		{
			desc:    "NOK_with_transform",
			value:   []int{4, 1},
			table:   "models",
			column:  "id",
			records: []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 3, Name: "c"}},
			valid:   true,
			transform: func(val int) clause.Expr {
				return gorm.Expr("?", val-1)
			},
			expected:                   true, // Always returns true, the errors are added to child elements
			expectedErrors:             []string{},
			expectedArrayElementErrors: []int{1},
		},
		{
			desc:                       "NOK",
			value:                      []int{3, 5, 7},
			table:                      "models",
			column:                     "id",
			records:                    []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 3, Name: "c"}},
			valid:                      true,
			transform:                  nil,
			expected:                   true, // Always returns true, the errors are added to child elements
			expectedErrors:             []string{},
			expectedArrayElementErrors: []int{1, 2},
		},
		{
			desc:           "not_a_slice_of_int",
			value:          []string{"7", "5"},
			table:          "models",
			column:         "id",
			records:        []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 3, Name: "c"}},
			valid:          true,
			transform:      nil,
			expected:       true,
			expectedErrors: []string{},
		},
		{
			desc:           "ctx_invalid",
			value:          []int{2, 3},
			table:          "models",
			column:         "id",
			records:        []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 3, Name: "c"}},
			valid:          false,
			transform:      nil,
			expected:       true,
			expectedErrors: []string{},
		},
		{
			desc:           "error",
			value:          []int{3, 2},
			table:          "models",
			column:         "not_a_column",
			records:        []uniqueTestModel{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 3, Name: "c"}},
			valid:          true,
			transform:      nil,
			expected:       false,
			expectedErrors: []string{"no such column: models.not_a_column"},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			opts := prepareUniqueTest(t)
			if len(c.records) > 0 {
				if err := opts.DB.Create(c.records).Error; err != nil {
					assert.FailNow(t, err.Error())
				}
			}

			v := ExistsArray[int](c.table, c.column, c.transform)
			v.init(opts)

			ctx := &Context{
				Invalid: !c.valid,
				Value:   c.value,
			}
			assert.Equal(t, c.expected, v.Validate(ctx))
			assert.Equal(t, c.expectedErrors, lo.Map(ctx.errors, func(e error, _ int) string { return e.Error() }))
			assert.Equal(t, c.expectedArrayElementErrors, ctx.arrayElementErrors)
		})
	}

	t.Run("buildQuery", func(t *testing.T) {
		cases := []struct {
			dialect  string
			expected string
		}{
			{dialect: "sqlite3", expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (2,0),(7,1),(6,2)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: "mysql", expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES ROW(2,0),ROW(7,1),ROW(6,2)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: "postgres", expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (2,0::int),(7,1::int),(6,2::int)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: "mssql", expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (2,0),(7,1),(6,2)) t(id,i)) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: "clickhouse", expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES 'id Int64, i Int64', (2,0),(7,1),(6,2))) SELECT i FROM ctx_values INNER JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
		}

		for _, c := range cases {
			t.Run(c.dialect, func(t *testing.T) {
				opts := prepareUniqueTest(t)
				opts.Config.Set("database.connection", c.dialect)
				v := ExistsArray[int]("models", "name", nil)
				v.init(opts)

				tx, _ := v.buildQuery([]int{2, 7, 6}, true)

				sql := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx
				})
				assert.Equal(t, c.expected, sql)
			})
		}
	})
}

func TestBuildQueryValidatorWithTransform(t *testing.T) {
	t.Run("buildQuery", func(t *testing.T) {
		cases := []struct {
			dialect  string
			expected string
		}{
			{dialect: "sqlite3", expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (1,0),(6,1),(5,2)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: "mysql", expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES ROW(1,0),ROW(6,1),ROW(5,2)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: "postgres", expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (1,0::int),(6,1::int),(5,2::int)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: "mssql", expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (1,0),(6,1),(5,2)) t(id,i)) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: "clickhouse", expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES 'id Int64, i Int64', (1,0),(6,1),(5,2))) SELECT i FROM ctx_values INNER JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
		}

		for _, c := range cases {
			t.Run(c.dialect, func(t *testing.T) {
				opts := prepareUniqueTest(t)
				opts.Config.Set("database.connection", c.dialect)
				transform := func(val int) clause.Expr {
					return gorm.Expr("?", val-1)
				}
				v := ExistsArray[int]("models", "name", transform)
				v.init(opts)

				tx, _ := v.buildQuery([]int{2, 7, 6}, true)

				sql := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx
				})
				assert.Equal(t, c.expected, sql)
			})
		}
	})
}

func TestClickhouseUnsupportedType(t *testing.T) {
	opts := prepareUniqueTest(t)
	opts.Config.Set("database.connection", "clickhouse")
	v := ExistsArray[struct{}]("models", "name", nil)
	v.init(opts)

	ctx := &Context{
		Value: []struct{}{{}, {}},
	}
	v.validate(ctx, true)
	require.Len(t, ctx.errors, 1)
	assert.Equal(t, "ExistsArray/UniqueArray validator: value of type T (struct {}) is not supported for Clickhouse. You must provide a Transform function", ctx.errors[0].Error())
}
