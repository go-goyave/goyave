package middleware

import (
	"io/ioutil"
	"testing"

	"github.com/System-Glitch/goyave/v2"
	"github.com/System-Glitch/goyave/v2/validation"
)

type DisallowMiddlewareTestSuite struct {
	goyave.TestSuite
}

func (suite *DisallowMiddlewareTestSuite) TestDisallowMiddleware() {
	request := suite.CreateTestRequest(nil)
	request.Data = map[string]interface{}{"non-validated": "hello world"}
	result := suite.Middleware(DisallowNonValidatedFields, request, func(response *goyave.Response, r *goyave.Request) {
		suite.Fail("DisallowNonValidatedFields shouldn't pass.")
	})
	suite.Equal(422, result.StatusCode)

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal("{\"validationError\":{\"non-validated\":[\"Non-validated fields are forbidden.\"]}}\n", string(body))

	request = suite.CreateTestRequest(nil)
	request.Data = map[string]interface{}{"non-validated": "hello world"}
	request.Rules = validation.RuleSet{"non-validated": {"string"}}
	result = suite.Middleware(DisallowNonValidatedFields, request, func(response *goyave.Response, r *goyave.Request) {})
	suite.Equal(200, result.StatusCode)
}

func TestDisallowMiddlewareTestSuite(t *testing.T) {
	goyave.RunTest(t, new(DisallowMiddlewareTestSuite))
}
