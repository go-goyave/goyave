package database

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"goyave.dev/goyave/v5/config"
)

type TestUser struct {
	Name  string `gorm:"type:varchar(100)"`
	Email string `gorm:"type:varchar(100)"`
	ID    uint   `gorm:"primaryKey"`
}

func userGenerator() *TestUser {
	return &TestUser{
		Name:  "John Doe",
		Email: "johndoe@example.org",
	}
}

func TestFactory(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		factory := NewFactory(userGenerator)

		if !assert.NotNil(t, factory) {
			return
		}
		assert.Equal(t, 100, factory.BatchSize)
		assert.Nil(t, factory.override)
	})

	t.Run("Generate", func(t *testing.T) {
		factory := NewFactory(userGenerator)

		records := factory.Generate(3)
		expected := []*TestUser{
			userGenerator(),
			userGenerator(),
			userGenerator(),
		}
		assert.Equal(t, expected, records)

		records = factory.Generate(0)
		assert.Equal(t, []*TestUser{}, records)
	})

	t.Run("Override", func(t *testing.T) {
		factory := NewFactory(userGenerator)
		factory.Override(&TestUser{Name: "name override"})

		records := factory.Generate(1)
		expected := []*TestUser{{
			Name:  "name override",
			Email: "johndoe@example.org",
		}}
		assert.Equal(t, expected, records)
	})

	t.Run("Save", func(t *testing.T) {
		cfg := &config.DatabaseConnection{
			Dialect:            "sqlmock",
			DatabaseName:       "paginator_test.db",
			MaxIdleConnections: 1,
			GORM:               config.GORM{}, // Disabling PrepareStmt is important to avoid errors caused by mock
		}

		mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		require.NoError(t, err)

		dialector := &sqlite.Dialector{
			DriverName: "sqlite3_timeout_test",
			DSN:        fmt.Sprintf("file:%s?%s", cfg.DatabaseName, cfg.Options),
			Conn:       mockDB,
		}

		// The SQLite dialector selects the sqlite version first to know which callback clauses it can use.
		mock.ExpectQuery(regexp.QuoteMeta(`select sqlite_version()`)).WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("3.53.4"))

		t.Cleanup(func() {
			mock.ExpectClose()
			assert.NoError(t, mockDB.Close())
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		db, err := NewFromDialector(cfg, nil, dialector)
		require.NoError(t, err)

		want := []*TestUser{
			userGenerator(),
			userGenerator(),
			userGenerator(),
		}
		mock.ExpectBegin() // Factory creates in batches so GORM does it in a transaction
		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO `test_users` (`name`,`email`) VALUES (?,?),(?,?),(?,?) RETURNING `id`")).
			WithArgs(want[0].Name, want[0].Email, want[1].Name, want[1].Email, want[2].Name, want[2].Email).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2).AddRow(3))
		mock.ExpectCommit()

		factory := NewFactory(userGenerator)
		records, err := factory.Save(db, 3)
		require.NoError(t, err)

		for i, u := range want {
			u.ID = uint(i + 1)
		}
		assert.Equal(t, want, records)
	})
}
