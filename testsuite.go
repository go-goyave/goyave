package goyave

import (
	"os"
	"testing"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/lang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	testify "github.com/stretchr/testify/suite"
)

// TestSuite is an extension of testify's Suite for
// Goyave-specific testing.
type TestSuite struct {
	suite.Suite
	previousEnv string
}

// RunTest run a test suite with prior initialization of a test environment.
// The GOYAVE_ENV environment variable is automatically set to "test" and restored
// to its original value at the end of the test run.
func RunTest(t *testing.T, suite suite.TestingSuite) {
	oldEnv := os.Getenv("GOYAVE_ENV")
	os.Setenv("GOYAVE_ENV", "test")
	defer os.Setenv("GOYAVE_ENV", oldEnv)
	if err := config.Load(); err != nil {
		assert.Fail(t, "Failed to load config", err)
	}
	lang.LoadDefault()
	lang.LoadAllAvailableLanguages()

	testify.Run(t, suite)
}
