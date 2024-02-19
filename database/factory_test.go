package database

import (
	"testing"

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
		RegisterDialect("sqlite3_factory_test", "file:{name}?{options}", sqlite.Open)
		t.Cleanup(func() {
			mu.Lock()
			delete(dialects, "sqlite3_factory_test")
			mu.Unlock()
		})

		cfg := config.LoadDefault()
		cfg.Set("app.debug", false)
		cfg.Set("database.connection", "sqlite3_factory_test")
		cfg.Set("database.name", "factory_test.db")
		cfg.Set("database.options", "mode=memory")
		db, err := New(cfg, nil)
		require.NoError(t, err)
		require.NoError(t, db.AutoMigrate(&TestUser{}))

		factory := NewFactory(userGenerator)
		records := factory.Save(db, 3)

		results := []*TestUser{}
		res := db.Find(&results)
		require.NoError(t, res.Error)
		assert.Equal(t, records, results)
	})
}
