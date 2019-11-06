package goyave

import (
	"testing"

	"github.com/System-Glitch/goyave/config"
	"github.com/stretchr/testify/suite"
)

type GoyaveTestSuite struct {
	suite.Suite
}

func (suite *GoyaveTestSuite) SetupSuite() {
	config.LoadConfig()
}

func (suite *GoyaveTestSuite) TestGetAddress() {
	suite.Equal("127.0.0.1:8080", getAddress("http"))
	suite.Equal("127.0.0.1:8081", getAddress("https"))
}

func TestGoyaveTestSuite(t *testing.T) {
	suite.Run(t, new(GoyaveTestSuite))
}
