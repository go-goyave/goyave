package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	previousEnv string
}

func (suite *ConfigTestSuite) SetupSuite() {
	suite.previousEnv = os.Getenv("GOYAVE_ENV")
	os.Setenv("GOYAVE_ENV", "test")
	if err := Load(); err != nil {
		suite.FailNow(err.Error())
	}
	suite.True(IsLoaded())
}

func (suite *ConfigTestSuite) TestLoadDefaults() {
	// TODO test loadDefaults
}

// TODO test override

// TODO test category Get

func (suite *ConfigTestSuite) TestLocalOverride() {
	os.Setenv("GOYAVE_ENV", "test")
	Clear()
	if err := Load(); err != nil {
		suite.FailNow(err.Error())
	}
	suite.Equal("root level content", Get("rootLevel"))
	Set("rootLevel", "root level content override")
	suite.Equal("root level content override", Get("rootLevel"))

	suite.Equal("test", Get("app.environment"))
	Set("app.environment", "test_override")
	suite.Equal("test_override", Get("app.environment"))
}

func (suite *ConfigTestSuite) TestGet() {
	suite.Equal("goyave", Get("app.name"))
	suite.Panics(func() {
		Get("missingKey") // TODO test with subcategory too
	})

	suite.Equal("goyave", GetString("app.name"))
	suite.Panics(func() {
		GetString("missingKey")
	})
	suite.Panics(func() {
		GetString("app.debug") // Not a string
	})

	suite.Equal(true, GetBool("app.debug"))
	suite.Panics(func() {
		GetBool("missingKey")
	})
	suite.Panics(func() {
		GetBool("app.name") // Not a bool
	})

	suite.Equal(8080, GetInt("server.port"))
	suite.Panics(func() {
		GetInt("missingKey")
	})
	suite.Panics(func() {
		GetInt("app.name") // Not an int
	})

	Set("testFloat", 1.42)
	suite.Equal(1.42, GetFloat("testFloat"))
	suite.Panics(func() {
		GetFloat("missingKey")
	})
	suite.Panics(func() {
		GetFloat("app.name") // Not a float
	})

	// TODO test with several depth levels
}

func (suite *ConfigTestSuite) TestHas() {
	suite.False(Has("not_a_config_entry"))
	suite.True(Has("rootLevel"))
	suite.True(Has("app.name"))
}

func (suite *ConfigTestSuite) TestGetEnv() {
	os.Setenv("GOYAVE_ENV", "localhost")
	suite.Equal("config.json", getConfigFilePath())

	os.Setenv("GOYAVE_ENV", "test")
	suite.Equal("config.test.json", getConfigFilePath())

	os.Setenv("GOYAVE_ENV", "production")
	suite.Equal("config.production.json", getConfigFilePath())

	os.Setenv("GOYAVE_ENV", "test")
}

func (suite *ConfigTestSuite) TestInvalidConfig() { // TODO add custom entry validation
	// val := Get("app.name")

	// TODO re-enable config validation tests
	// config["app.name"] = true
	// err := validateConfig()
	// suite.NotNil(err)
	// suite.Equal("Invalid config:\n\t- \"app.name\" type must be string", err.Error())
	// config["app.name"] = val

	// suite.Panics(func() {
	// 	Set("app.name", true)
	// })

	// val = Get("database.connection")

	// TODO re-enable config validation tests
	// config["database.connection"] = "not a driver"
	// err = validateConfig()
	// suite.NotNil(err)
	// suite.Equal("Invalid config:\n\t- \"database.connection\" must have one of the following values: none, mysql, postgres, sqlite3, mssql", err.Error())
	// config["database.connection"] = val

	// suite.Panics(func() {
	// 	Set("server.protocol", "ftp") // Unsupported protocol
	// })

	// os.Setenv("GOYAVE_ENV", "test_invalid")
	// config = nil
	// suite.NotNil(Load())
	// os.Setenv("GOYAVE_ENV", "test")
}

func (suite *ConfigTestSuite) TearDownAllSuite() {
	config = map[string]interface{}{}
	os.Setenv("GOYAVE_ENV", suite.previousEnv)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
