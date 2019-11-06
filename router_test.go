package goyave

import (
	"testing"

	"github.com/System-Glitch/goyave/config"
	"github.com/stretchr/testify/suite"
)

type RouterTestSuite struct {
	suite.Suite
}

func (suite *RouterTestSuite) SetupSuite() {
	config.LoadConfig()
}

func (suite *RouterTestSuite) TestCleanStaticPath() {
	suite.Equal("config/index.html", cleanStaticPath("config", "index.html"))
	suite.Equal("config/index.html", cleanStaticPath("config", ""))
	suite.Equal("config/defaults.json", cleanStaticPath("config", "defaults.json"))
	suite.Equal("resources/lang/en-US/locale.json", cleanStaticPath("resources", "lang/en-US/locale.json"))
	suite.Equal("resources/lang/en-US/locale.json", cleanStaticPath("resources", "/lang/en-US/locale.json"))
	suite.Equal("resources/lang/en-US/index.html", cleanStaticPath("resources", "lang/en-US"))
	suite.Equal("resources/lang/en-US/index.html", cleanStaticPath("resources", "lang/en-US/"))
}

func (suite *RouterTestSuite) TestStaticHandler() {
	// TODO test static handler
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}
