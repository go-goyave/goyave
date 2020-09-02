package database

import (
	"os"
	"testing"

	"github.com/System-Glitch/goyave/v3/config"
	"github.com/stretchr/testify/suite"

	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	ID    uint   `gorm:"primary_key"`
	Name  string `gorm:"type:varchar(100)"`
	Email string `gorm:"type:varchar(100)"`
}

type DatabaseTestSuite struct {
	suite.Suite
	previousEnv string
}

func (suite *DatabaseTestSuite) SetupSuite() {
	suite.previousEnv = os.Getenv("GOYAVE_ENV")
	os.Setenv("GOYAVE_ENV", "test")
	if err := config.Load(); err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *DatabaseTestSuite) TestBuildConnectionOptions() {
	suite.Equal("goyave:secret@(127.0.0.1:3306)/goyave?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local", buildConnectionOptions("mysql"))
	suite.Equal("host=127.0.0.1 port=3306 user=goyave dbname=goyave password=secret charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local", buildConnectionOptions("postgres"))
	suite.Equal("goyave", buildConnectionOptions("sqlite3"))
	suite.Equal("sqlserver://goyave:secret@127.0.0.1:3306?database=goyave&charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local", buildConnectionOptions("mssql"))

	suite.Panics(func() {
		buildConnectionOptions("test")
	})
}

func (suite *DatabaseTestSuite) TestGetConnection() {
	db := GetConnection()
	suite.NotNil(db)
	suite.NotNil(dbConnection)
	suite.Equal(dbConnection, db)
	suite.Nil(Close())
	suite.Nil(dbConnection)

	db = Conn()
	suite.NotNil(db)
	suite.Equal(dbConnection, db)
	Close()
}

func (suite *DatabaseTestSuite) TestLogLevel() {
	db := GetConnection()
	suite.Equal(logger.Default.LogMode(logger.Silent), db.Logger)
	Close()
	prev := config.Get("app.debug")
	config.Set("app.debug", true)
	defer config.Set("app.debug", prev)
	db = GetConnection()
	suite.Equal(logger.Default.LogMode(logger.Info), db.Logger)
	Close()
}

func (suite *DatabaseTestSuite) TestGetConnectionPanic() {
	tmpConnection := config.Get("database.connection")
	config.Set("database.connection", "none")
	suite.Panics(func() {
		GetConnection()
	})
	config.Set("database.connection", tmpConnection)

	tmpPort := config.Get("database.port")
	config.Set("database.port", 0.0)
	suite.Panics(func() {
		GetConnection()
	})
	config.Set("database.port", tmpPort)
}

func (suite *DatabaseTestSuite) TestModelAndMigrate() {
	ClearRegisteredModels()
	RegisterModel(&User{})
	suite.Len(models, 1)

	registeredModels := GetRegisteredModels()
	suite.Len(registeredModels, 1)
	suite.Same(models[0], registeredModels[0])

	Migrate()
	ClearRegisteredModels()
	suite.Equal(0, len(models))

	db := GetConnection()
	defer db.Migrator().DropTable(&User{})

	rows, err := db.Raw("SHOW TABLES;").Rows()
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		name := ""
		if err := rows.Scan(&name); err != nil {
			panic(err)
		}
		if name == "users" {
			found = true
			break
		}
	}

	suite.True(found)
}

func (suite *DatabaseTestSuite) TestInitializers() {
	initializer := func(db *gorm.DB) {
		db.Config.SkipDefaultTransaction = true
	}
	AddInitializer(initializer)

	suite.Len(initializers, 1)

	db := GetConnection()
	suite.True(db.Config.SkipDefaultTransaction)

	suite.Nil(Close())

	AddInitializer(func(db *gorm.DB) {
		db.Statement.Settings.Store("gorm:table_options", "ENGINE=InnoDB")
		// TODO update docs for initializers
	})
	suite.Len(initializers, 2)

	db = GetConnection()
	suite.True(db.Config.SkipDefaultTransaction)
	val, ok := db.Get("gorm:table_options")
	suite.True(ok)
	suite.Equal("ENGINE=InnoDB", val)

	suite.Nil(Close())

	ClearInitializers()
	suite.Empty(initializers)
}

func (suite *DatabaseTestSuite) TestRegisterDialect() {
	template := "{username}{username} {password} {host}:{port} {name} {options}"
	RegisterDialect("newdialect", template)
	defer delete(dialectOptions, "newdialect")

	t, ok := dialectOptions["newdialect"]
	suite.True(ok)
	suite.Equal(template, t)

	suite.Equal("goyave{username} secret 127.0.0.1:3306 goyave charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local", buildConnectionOptions("newdialect"))

	suite.Panics(func() {
		RegisterDialect("newdialect", "othertemplate")
	})
	t, ok = dialectOptions["newdialect"]
	suite.True(ok)
	suite.Equal(template, t)
}

func (suite *DatabaseTestSuite) TearDownAllSuite() {
	os.Setenv("GOYAVE_ENV", suite.previousEnv)
}

func TestDatabaseTestSuite(t *testing.T) {
	// Ensure this test is running with a working database service running
	// in the background. Running "run_test.sh" runs a mariadb container.
	suite.Run(t, new(DatabaseTestSuite))
}
