package middleware

import (
	"testing"

	"goyave.dev/goyave/v3"
)

type TrimMiddlewareTestSuite struct {
	goyave.TestSuite
}

func (suite *TrimMiddlewareTestSuite) TestTrimMiddleware() {
	request := suite.CreateTestRequest(nil)
	request.Data = map[string]interface{}{"text": " \t  trimmed\n  \t"}
	suite.Middleware(Trim, request, func(response *goyave.Response, r *goyave.Request) {
		suite.Equal("trimmed", r.String("text"))
	})
}

func TestTrimMiddlewareTestSuite(t *testing.T) {
	goyave.RunTest(t, new(TrimMiddlewareTestSuite))
}
