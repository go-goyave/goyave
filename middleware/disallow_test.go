package middleware

import (
	"io/ioutil"
	"testing"

	"github.com/System-Glitch/goyave/v3"
	"github.com/System-Glitch/goyave/v3/validation"
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
	request.Data = map[string]interface{}{"validated": "hello world"}
	request.Rules = &validation.Rules{
		Fields: validation.FieldMap{
			"validated": {
				Rules: []*validation.Rule{
					{Name: "string"},
				},
			},
		},
	}
	result = suite.Middleware(DisallowNonValidatedFields, request, func(response *goyave.Response, r *goyave.Request) {})
	suite.Equal(204, result.StatusCode)

	request.Data = map[string]interface{}{"non-validated": "hello world", "validated": "hello world"}
	result = suite.Middleware(DisallowNonValidatedFields, request, func(response *goyave.Response, r *goyave.Request) {
		suite.Fail("DisallowNonValidatedFields shouldn't pass.")
	})
	suite.Equal(422, result.StatusCode)

	body, err = ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal("{\"validationError\":{\"non-validated\":[\"Non-validated fields are forbidden.\"]}}\n", string(body))

	// With confirmation
	request.Data = map[string]interface{}{"validated": "hello world", "validated_confirmation": "hello world"}
	request.Rules = &validation.Rules{
		Fields: validation.FieldMap{
			"validated": {
				Rules: []*validation.Rule{
					{Name: "string"},
					{Name: "confirmed"},
				},
			},
		},
	}
	result = suite.Middleware(DisallowNonValidatedFields, request, func(response *goyave.Response, r *goyave.Request) {})
	result.Body.Close()
	suite.Equal(204, result.StatusCode)

	suite.NotPanics(func() {
		request.Data = map[string]interface{}{"non-validated": "hello world"}
		request.Rules = &validation.Rules{
			Fields: nil,
		}
		result := suite.Middleware(DisallowNonValidatedFields, request, func(response *goyave.Response, r *goyave.Request) {})
		suite.Equal(422, result.StatusCode)

		body, err = ioutil.ReadAll(result.Body)
		if err != nil {
			panic(err)
		}
		result.Body.Close()
		suite.Equal("{\"validationError\":{\"non-validated\":[\"Non-validated fields are forbidden.\"]}}\n", string(body))
	})
}

func TestDisallowMiddlewareTestSuite(t *testing.T) {
	goyave.RunTest(t, new(DisallowMiddlewareTestSuite))
}
