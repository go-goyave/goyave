package database

import (
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"goyave.dev/goyave/v4/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils/tests"
)

type User struct {
	Name  string `gorm:"type:varchar(100)"`
	Email string `gorm:"type:varchar(100)"`
	ID    uint   `gorm:"primaryKey"`
}

type DatabaseTestSuite struct {
	suite.Suite
	previousEnv string
}

func (suite *DatabaseTestSuite) SetupSuite() {
	if _, ok := dialects["mysql"]; !ok {
		RegisterDialect("mysql", "{username}:{password}@({host}:{port})/{name}?{options}", mysql.Open)
	}
	suite.previousEnv = os.Getenv("GOYAVE_ENV")
	os.Setenv("GOYAVE_ENV", "test")
	if err := config.Load(); err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *DatabaseTestSuite) TestBuildDSN() {
	d := dialect{nil, "{username}:{password}@({host}:{port})/{name}?{options}"}
	suite.Equal("goyave:secret@(127.0.0.1:3306)/goyave?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local", d.buildDSN())
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
	suite.Nil(Close())
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
	config.Set("database.port", 0)
	suite.Panics(func() {
		GetConnection()
	})
	config.Set("database.port", tmpPort)

	tmpConnection = config.Get("database.connection")
	config.Set("database.connection", "notadriver")
	suite.Panics(func() {
		GetConnection()
	})
	config.Set("database.connection", tmpConnection)
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
	RegisterDialect("newdialect", template, nil)
	defer delete(dialects, "newdialect")

	t, ok := dialects["newdialect"]
	suite.True(ok)
	suite.Equal(template, t.template)

	suite.Equal("goyave{username} secret 127.0.0.1:3306 goyave charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local", t.buildDSN())

	suite.Panics(func() {
		RegisterDialect("newdialect", "othertemplate", nil)
	})
	t, ok = dialects["newdialect"]
	suite.True(ok)
	suite.Equal(template, t.template)
}

type DummyDialector struct {
	tests.DummyDialector
	DriverName string
}

func (d DummyDialector) Initialize(db *gorm.DB) error {
	pool, err := sql.Open(d.DriverName, "")
	if err != nil {
		return err
	}
	db.ConnPool = pool
	return d.DummyDialector.Initialize(db)
}

func (suite *DatabaseTestSuite) TestSetConnection() {
	Close()
	defer Close()
	initializerOK := false
	AddInitializer(func(db *gorm.DB) {
		initializerOK = true
	})

	// No connection yet
	db, err := SetConnection(DummyDialector{DriverName: "mysql"})
	suite.Nil(err)
	suite.Same(db, dbConnection)
	suite.True(initializerOK)

	// Connection replaced
	db2, err2 := SetConnection(DummyDialector{DriverName: "mysql"})
	suite.Nil(err2)
	suite.Same(db2, dbConnection)
	suite.NotSame(db, db2)

	// Open error
	db3, err3 := SetConnection(DummyDialector{DriverName: "not a driver"})
	suite.NotNil(err3)
	suite.Nil(db3)
}

func (suite *DatabaseTestSuite) TearDownAllSuite() {
	os.Setenv("GOYAVE_ENV", suite.previousEnv)
}

func TestDatabaseTestSuite(t *testing.T) {
	// Ensure this test is running with a working database service running
	// in the background. Running "run_test.sh" runs a mariadb container.
	suite.Run(t, new(DatabaseTestSuite))
}
