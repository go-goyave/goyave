package database

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/bigquery"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/utils/tests"
)

var (
	dialectorNameSQLite     = sqlite.Dialector{}.Name()
	dialectorNamePostgres   = postgres.Dialector{}.Name()
	dialectorNameMySQL      = mysql.Dialector{}.Name()
	dialectorNameMSSQL      = sqlserver.Dialector{}.Name()
	dialectorNameClickhouse = clickhouse.Dialector{}.Name()
	dialectorNameBigquery   = bigquery.Dialector{}.Name()
)

type CustomDialector struct {
	tests.DummyDialector
	name string
}

func (d CustomDialector) Name() string {
	return d.name
}

func prepareExistTest(t *testing.T, dialectorName string) *gorm.DB {
	db, err := gorm.Open(&CustomDialector{name: dialectorName})
	require.NoError(t, err)
	return db
}

func prepareMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	dialector := &sqlite.Dialector{
		DriverName: "sqlite3",
		DSN:        "file:exist_test.db?mode=memory",
		Conn:       mockDB,
	}

	// The SQLite dialector selects the sqlite version first to know which callback clauses it can use.
	mock.ExpectQuery(regexp.QuoteMeta(`select sqlite_version()`)).WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("3.53.4"))

	t.Cleanup(func() {
		mock.ExpectClose()
		assert.NoError(t, mockDB.Close())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	db, err := gorm.Open(dialector)
	require.NoError(t, err)

	return db, mock
}

func TestExist(t *testing.T) {
	t.Run("buildSliceQuery", func(t *testing.T) {
		cases := []struct {
			dialect  string
			expected string
		}{
			{dialect: dialectorNameSQLite, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (2,0),(7,1),(6,2)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: dialectorNameBigquery, expected: "WITH ctx_values AS (SELECT * FROM UNNEST([STRUCT(2 AS id, 0 AS i),STRUCT(7 AS id, 1 AS i),STRUCT(6 AS id, 2 AS i)])) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: dialectorNameMySQL, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES ROW(2,0),ROW(7,1),ROW(6,2)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: dialectorNamePostgres, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (2,0::int),(7,1::int),(6,2::int)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: dialectorNameMSSQL, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (2,0),(7,1),(6,2)) t(id,i)) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: dialectorNameClickhouse, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES 'id Int64, i Int64', (2,0),(7,1),(6,2))) SELECT i FROM ctx_values INNER JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
		}

		for _, c := range cases {
			t.Run(c.dialect, func(t *testing.T) {
				db := prepareExistTest(t, c.dialect)
				e := NewExist[int]("models", "name")

				tx, _ := e.buildSliceQuery(db, []int{2, 7, 6}, true)

				sql := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx
				})
				assert.Equal(t, c.expected, sql)
			})
		}
	})

	t.Run("buildSliceQuery_withTransform", func(t *testing.T) {
		cases := []struct {
			dialect  string
			expected string
		}{
			{dialect: dialectorNameSQLite, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (1,0),(6,1),(5,2)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: dialectorNameBigquery, expected: "WITH ctx_values AS (SELECT * FROM UNNEST([STRUCT(1 AS id, 0 AS i),STRUCT(6 AS id, 1 AS i),STRUCT(5 AS id, 2 AS i)])) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: dialectorNameMySQL, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES ROW(1,0),ROW(6,1),ROW(5,2)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: dialectorNamePostgres, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (1,0::int),(6,1::int),(5,2::int)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: dialectorNameMSSQL, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (1,0),(6,1),(5,2)) t(id,i)) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
			{dialect: dialectorNameClickhouse, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES 'id Int64, i Int64', (1,0),(6,1),(5,2))) SELECT i FROM ctx_values INNER JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS  NULL"},
		}

		for _, c := range cases {
			t.Run(c.dialect, func(t *testing.T) {
				db := prepareExistTest(t, c.dialect)
				e := NewExist[int]("models", "name")
				e.Transform = func(val int) clause.Expr {
					return gorm.Expr("?", val-1)
				}

				tx, _ := e.buildSliceQuery(db, []int{2, 7, 6}, true)

				sql := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx
				})
				assert.Equal(t, c.expected, sql)
			})
		}
	})

	t.Run("clickhouse_unsupported_type", func(t *testing.T) {
		db := prepareExistTest(t, dialectorNameClickhouse)
		e := NewExist[struct{}]("table", "column")
		indexes, err := e.CheckSlice(db, []struct{}{})
		assert.Nil(t, indexes)
		assert.ErrorContains(t, err, "database.Exist: value of type T (struct {}) is not supported for Clickhouse. You must provide a Transform function")
	})

	t.Run("Check", func(t *testing.T) {
		cases := []struct {
			desc          string
			key           string
			wantKey       string
			transform     func(val string) clause.Expr
			returnedCount int
			sqlErr        error
			want          bool
		}{
			{
				desc:          "exists",
				key:           "johndoe@example.org",
				wantKey:       "johndoe@example.org",
				transform:     nil,
				returnedCount: 1,
				sqlErr:        nil,
				want:          true,
			},
			{
				desc:          "does_not_exist",
				key:           "johndoe@example.org",
				wantKey:       "johndoe@example.org",
				transform:     nil,
				returnedCount: 0,
				sqlErr:        nil,
				want:          false,
			},
			{
				desc:          "exists_with_transform",
				key:           "johndoe@example.org",
				wantKey:       "transformed",
				transform:     func(_ string) clause.Expr { return gorm.Expr("?", "transformed") },
				returnedCount: 1,
				sqlErr:        nil,
				want:          true,
			},
			{
				desc:          "error",
				key:           "johndoe@example.org",
				wantKey:       "johndoe@example.org",
				transform:     nil,
				returnedCount: 0,
				sqlErr:        fmt.Errorf("test error"),
				want:          false,
			},
		}

		for _, c := range cases {
			t.Run(c.desc, func(t *testing.T) {
				db, mock := prepareMockDB(t)

				expectedQuery := mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users` WHERE `email` = ?")).WithArgs(c.wantKey)
				if c.sqlErr != nil {
					expectedQuery.WillReturnError(c.sqlErr)
				} else {
					expectedQuery.WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(c.returnedCount))
				}

				e := NewExist[string]("users", "email")
				e.Transform = c.transform
				exist, err := e.Check(db, c.key)
				if c.sqlErr != nil {
					assert.ErrorIs(t, err, c.sqlErr)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, c.want, exist)
			})
		}
	})

	t.Run("CheckSlice", func(t *testing.T) {
		// In this test we only want to check that the results of the query are correctely returned.
		// The query itself is tested in the "buildSliceQuery" test case.
		db, mock := prepareMockDB(t)

		emails := []string{"johndoe@example.org", "test@dummy.com", "bob@alice.fr"}

		mock.ExpectQuery(regexp.QuoteMeta("WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (?,?),(?,?),(?,?)) t) SELECT i FROM ctx_values LEFT JOIN `users` ON `users`.`email` = ctx_values.id WHERE `users`.`email` IS NULL")).
			WithArgs(emails[0], 0, emails[1], 1, emails[2], 2).
			WillReturnRows(sqlmock.NewRows([]string{"i"}).AddRow(1).AddRow(2))

		e := NewExist[string]("users", "email")
		indexes, err := e.CheckSlice(db, emails)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2}, indexes)
	})
}

func TestUnique(t *testing.T) {
	t.Run("buildSliceQuery", func(t *testing.T) {
		cases := []struct {
			dialect  string
			expected string
		}{
			{dialect: dialectorNameSQLite, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (2,0),(7,1),(6,2)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS NOT NULL"},
			{dialect: dialectorNameBigquery, expected: "WITH ctx_values AS (SELECT * FROM UNNEST([STRUCT(2 AS id, 0 AS i),STRUCT(7 AS id, 1 AS i),STRUCT(6 AS id, 2 AS i)])) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS NOT NULL"},
			{dialect: dialectorNameMySQL, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES ROW(2,0),ROW(7,1),ROW(6,2)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS NOT NULL"},
			{dialect: dialectorNamePostgres, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (2,0::int),(7,1::int),(6,2::int)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS NOT NULL"},
			{dialect: dialectorNameMSSQL, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (2,0),(7,1),(6,2)) t(id,i)) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS NOT NULL"},
			{dialect: dialectorNameClickhouse, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES 'id Int64, i Int64', (2,0),(7,1),(6,2))) SELECT i FROM ctx_values INNER JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS NOT NULL"},
		}

		for _, c := range cases {
			t.Run(c.dialect, func(t *testing.T) {
				db := prepareExistTest(t, c.dialect)
				e := NewUnique[int]("models", "name")

				tx, _ := e.buildSliceQuery(db, []int{2, 7, 6}, false)

				sql := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx
				})
				assert.Equal(t, c.expected, sql)
			})
		}
	})

	t.Run("buildSliceQuery_withTransform", func(t *testing.T) {
		cases := []struct {
			dialect  string
			expected string
		}{
			{dialect: dialectorNameSQLite, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (1,0),(6,1),(5,2)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS NOT NULL"},
			{dialect: dialectorNameBigquery, expected: "WITH ctx_values AS (SELECT * FROM UNNEST([STRUCT(1 AS id, 0 AS i),STRUCT(6 AS id, 1 AS i),STRUCT(5 AS id, 2 AS i)])) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS NOT NULL"},
			{dialect: dialectorNameMySQL, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES ROW(1,0),ROW(6,1),ROW(5,2)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS NOT NULL"},
			{dialect: dialectorNamePostgres, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (1,0::int),(6,1::int),(5,2::int)) t) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS NOT NULL"},
			{dialect: dialectorNameMSSQL, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (1,0),(6,1),(5,2)) t(id,i)) SELECT i FROM ctx_values LEFT JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS NOT NULL"},
			{dialect: dialectorNameClickhouse, expected: "WITH ctx_values(id, i) AS (SELECT * FROM (VALUES 'id Int64, i Int64', (1,0),(6,1),(5,2))) SELECT i FROM ctx_values INNER JOIN `models` ON `models`.`name` = ctx_values.id WHERE `models`.`name` IS NOT NULL"},
		}

		for _, c := range cases {
			t.Run(c.dialect, func(t *testing.T) {
				db := prepareExistTest(t, c.dialect)
				e := NewUnique[int]("models", "name")
				e.Transform = func(val int) clause.Expr {
					return gorm.Expr("?", val-1)
				}

				tx, _ := e.buildSliceQuery(db, []int{2, 7, 6}, false)

				sql := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx
				})
				assert.Equal(t, c.expected, sql)
			})
		}
	})

	t.Run("clickhouse_unsupported_type", func(t *testing.T) {
		db := prepareExistTest(t, dialectorNameClickhouse)
		e := NewUnique[struct{}]("table", "column")
		indexes, err := e.CheckSlice(db, []struct{}{})
		assert.Nil(t, indexes)
		assert.ErrorContains(t, err, "database.Exist: value of type T (struct {}) is not supported for Clickhouse. You must provide a Transform function")
	})

	t.Run("Check", func(t *testing.T) {
		cases := []struct {
			desc          string
			key           string
			wantKey       string
			transform     func(val string) clause.Expr
			returnedCount int
			sqlErr        error
			want          bool
		}{
			{
				desc:          "unique",
				key:           "johndoe@example.org",
				wantKey:       "johndoe@example.org",
				transform:     nil,
				returnedCount: 0,
				sqlErr:        nil,
				want:          true,
			},
			{
				desc:          "not_unique",
				key:           "johndoe@example.org",
				wantKey:       "johndoe@example.org",
				transform:     nil,
				returnedCount: 1,
				sqlErr:        nil,
				want:          false,
			},
			{
				desc:          "unique_with_transform",
				key:           "johndoe@example.org",
				wantKey:       "transformed",
				transform:     func(_ string) clause.Expr { return gorm.Expr("?", "transformed") },
				returnedCount: 0,
				sqlErr:        nil,
				want:          true,
			},
			{
				desc:          "error",
				key:           "johndoe@example.org",
				wantKey:       "johndoe@example.org",
				transform:     nil,
				returnedCount: 0,
				sqlErr:        fmt.Errorf("test error"),
				want:          false,
			},
		}

		for _, c := range cases {
			t.Run(c.desc, func(t *testing.T) {
				db, mock := prepareMockDB(t)

				expectedQuery := mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users` WHERE `email` = ?")).WithArgs(c.wantKey)
				if c.sqlErr != nil {
					expectedQuery.WillReturnError(c.sqlErr)
				} else {
					expectedQuery.WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(c.returnedCount))
				}

				e := NewUnique[string]("users", "email")
				e.Transform = c.transform
				unique, err := e.Check(db, c.key)
				if c.sqlErr != nil {
					assert.ErrorIs(t, err, c.sqlErr)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, c.want, unique)
			})
		}
	})

	t.Run("CheckSlice", func(t *testing.T) {
		// In this test we only want to check that the results of the query are correctely returned.
		// The query itself is tested in the "buildSliceQuery" test case.
		db, mock := prepareMockDB(t)

		emails := []string{"johndoe@example.org", "test@dummy.com", "bob@alice.fr"}

		mock.ExpectQuery(regexp.QuoteMeta("WITH ctx_values(id, i) AS (SELECT * FROM (VALUES (?,?),(?,?),(?,?)) t) SELECT i FROM ctx_values LEFT JOIN `users` ON `users`.`email` = ctx_values.id WHERE `users`.`email` IS NOT NULL")).
			WithArgs(emails[0], 0, emails[1], 1, emails[2], 2).
			WillReturnRows(sqlmock.NewRows([]string{"i"}).AddRow(1).AddRow(2))

		e := NewUnique[string]("users", "email")
		indexes, err := e.CheckSlice(db, emails)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2}, indexes)
	})
}
