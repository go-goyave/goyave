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
	LoadConfig()
}

func (suite *ConfigTestSuite) TestLocalOverride() {
	suite.Equal("test", Get("environment"))
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
	suite.False(validateConfig())
	config["appName"] = val

	suite.Panics(func() {
		Set("appName", true)
	})

	suite.Panics(func() {
		Set("protocol", "ftp") // Unsupported protocol
	})
}

func (suite *ConfigTestSuite) TearDownAllSuite() {
	config = map[string]interface{}{}
	os.Setenv("GOYAVE_ENV", suite.previousEnv)

}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
