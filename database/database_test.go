package database

import (
	"os"
	"testing"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/stretchr/testify/suite"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

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
	suite.Equal("goyave:secret@(127.0.0.1:3306)/goyave?charset=utf8&parseTime=true&loc=Local", buildConnectionOptions("mysql"))
	suite.Equal("host=127.0.0.1 port=3306 user=goyave dbname=goyave password=secret options='charset=utf8&parseTime=true&loc=Local'", buildConnectionOptions("postgres"))
	suite.Equal("goyave", buildConnectionOptions("sqlite3"))
	suite.Equal("sqlserver://goyave:secret@127.0.0.1:3306?database=goyave&charset=utf8&parseTime=true&loc=Local", buildConnectionOptions("mssql"))

	suite.Panics(func() {
		buildConnectionOptions("test")
	})
}

func (suite *DatabaseTestSuite) testGetConnection() { // TODO re-enable test
	db := GetConnection()
	suite.NotNil(db)
	suite.NotNil(dbConnection)
	suite.Equal(dbConnection, db)
	Close()
	suite.Nil(dbConnection)
}

func (suite *DatabaseTestSuite) testGetConnectionPanic() { // TODO re-enable test
	tmpConnection := config.Get("dbConnection")
	config.Set("dbConnection", "none")
	suite.Panics(func() {
		GetConnection()
	})
	config.Set("dbConnection", tmpConnection)

	tmpPort := config.Get("dbPort")
	config.Set("dbPort", 0.0)
	suite.Panics(func() {
		GetConnection()
	})
	config.Set("dbPort", tmpPort)
}

func (suite *DatabaseTestSuite) TearDownAllSuite() {
	os.Setenv("GOYAVE_ENV", suite.previousEnv)
}

func TestDatabaseTestSuite(t *testing.T) {
	// Ensure this test is running with a working database service running
	// in the background. Running "run_test.sh" runs a mariadb container.
	suite.Run(t, new(DatabaseTestSuite))
}
