package config

import (
	"os"
	"reflect"
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

func (suite *ConfigTestSuite) TestLocalOverride() {
	os.Setenv("GOYAVE_ENV", "test")
	config = nil
	if err := Load(); err != nil {
		suite.FailNow(err.Error())
	}
	suite.Equal("test", Get("environment"))
	Set("environment", "test_override")
	suite.Equal("test_override", Get("environment"))
}

func (suite *ConfigTestSuite) TestGet() {
	suite.Equal("goyave", Get("appName"))
	suite.Panics(func() {
		Get("missingKey")
	})

	suite.Equal("goyave", GetString("appName"))
	suite.Panics(func() {
		GetString("missingKey")
	})
	suite.Panics(func() {
		GetString("debug") // Not a string
	})

	suite.Equal(true, GetBool("debug"))
	suite.Panics(func() {
		GetBool("missingKey")
	})
	suite.Panics(func() {
		GetBool("appName") // Not a bool
	})
}

func (suite *ConfigTestSuite) TestHas() {
	suite.False(Has("not_a_config_entry"))
	suite.True(Has("appName"))
}

func (suite *ConfigTestSuite) TestRegister() {
	Set("register_test", "value")
	suite.Panics(func() {
		Register("register_test", reflect.Struct)
	})
	delete(configValidation, "register_test")
	delete(config, "register_test")

	type configStruct struct{}
	Register("register_test", reflect.Struct)
	Set("register_test", configStruct{})
	suite.Panics(func() {
		Set("register_test", "value")
	})
	suite.Panics(func() { // Already registered
		Register("register_test", reflect.Struct)
	})
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

func (suite *ConfigTestSuite) TestInvalidConfig() {
	val := Get("appName")

	config["appName"] = true
	err := validateConfig()
	suite.NotNil(err)
	suite.Equal("Invalid config:\n\t- \"appName\" type must be string", err.Error())
	config["appName"] = val

	suite.Panics(func() {
		Set("appName", true)
	})

	val = Get("dbConnection")

	config["dbConnection"] = "not a driver"
	err = validateConfig()
	suite.NotNil(err)
	suite.Equal("Invalid config:\n\t- \"dbConnection\" must have one of the following values: none, mysql, postgres, sqlite3, mssql", err.Error())
	config["dbConnection"] = val

	suite.Panics(func() {
		Set("protocol", "ftp") // Unsupported protocol
	})

	os.Setenv("GOYAVE_ENV", "test_invalid")
	config = nil
	suite.NotNil(Load())
	os.Setenv("GOYAVE_ENV", "test")
}

func (suite *ConfigTestSuite) TearDownAllSuite() {
	config = map[string]interface{}{}
	os.Setenv("GOYAVE_ENV", suite.previousEnv)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
