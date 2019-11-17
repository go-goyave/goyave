package database

import (
	"testing"

	"github.com/System-Glitch/goyave/config"
	"github.com/stretchr/testify/suite"
)

type DatabaseTestSuite struct {
	suite.Suite
}

func (suite *DatabaseTestSuite) SetupSuite() {
	config.LoadConfig()
}

func (suite *DatabaseTestSuite) TestBuildConnectionOptions() {
	suite.Equal("root:root@(127.0.0.1:3306)/goyave?charset=utf8&parseTime=true&loc=Local", buildConnectionOptions("mysql"))
	suite.Equal("host=127.0.0.1 port=3306 user=root dbname=goyave password=root options='charset=utf8&parseTime=true&loc=Local'", buildConnectionOptions("postgres"))
	suite.Equal("goyave", buildConnectionOptions("sqlite3"))
	suite.Equal("sqlserver://root:root@127.0.0.1:3306?database=goyave&charset=utf8&parseTime=true&loc=Local", buildConnectionOptions("mssql"))

	suite.Panics(func() {
		buildConnectionOptions("test")
	})
}

func TestDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}
