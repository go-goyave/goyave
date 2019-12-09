package goyave

import (
	"os"
	"testing"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/lang"
	"github.com/stretchr/testify/assert"
)

type CustomTestSuite struct {
	TestSuite
}

type FailingTestSuite struct {
	TestSuite
}

func (suite *CustomTestSuite) TestEnv() {
	suite.Equal("test", os.Getenv("GOYAVE_ENV"))
	suite.Equal("test", config.GetString("environment"))
	suite.Equal("Malformed JSON", lang.Get("en-US", "malformed-json"))
}

func TestTestSuite(t *testing.T) {
	RunTest(t, new(CustomTestSuite))
}

func TestTestSuiteFail(t *testing.T) {
	os.Rename("config.test.json", "config.test.json.bak")
	mockT := new(testing.T)
	RunTest(mockT, new(FailingTestSuite))
	assert.True(t, mockT.Failed())
	os.Rename("config.test.json.bak", "config.test.json")
}
