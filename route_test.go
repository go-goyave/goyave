package goyave

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RouteTestSuite struct {
	suite.Suite
}

func (suite *RouteTestSuite) TestSomething() {

}

func TestRouteTestSuite(t *testing.T) {
	suite.Run(t, new(RouteTestSuite))
}
