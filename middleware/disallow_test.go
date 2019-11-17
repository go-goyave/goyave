package middleware

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/System-Glitch/goyave"
	"github.com/System-Glitch/goyave/lang"
	"github.com/System-Glitch/goyave/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DisallowMiddlewareTestSuite struct {
	suite.Suite
}

func (suite *DisallowMiddlewareTestSuite) SetupSuite() {
	lang.LoadDefault()
}

func TestDisallowMiddleware(t *testing.T) {
	request := &goyave.Request{
		Data:   map[string]interface{}{"non-validated": "hello world"},
		Rules:  validation.RuleSet{},
		Params: map[string]string{},
	}
	recorder := httptest.NewRecorder()
	response := goyave.CreateTestResponse(recorder)
	DisallowNonValidatedFields(func(response *goyave.Response, r *goyave.Request) {
		result := recorder.Result()
		assert.Equal(t, 422, result.StatusCode)

		body, err := ioutil.ReadAll(result.Body)
		if err != nil {
			panic(err)
		}
		assert.Equal(t, "{\"validationErrors\":{\"non-validated\":[\"Non-validated fields are forbidden.\"]}}\n", string(body))
	})(response, request)

	request = &goyave.Request{
		Data:   map[string]interface{}{"non-validated": "hello world"},
		Rules:  validation.RuleSet{"non-validated": {"string"}},
		Params: map[string]string{},
	}
	recorder = httptest.NewRecorder()
	response = goyave.CreateTestResponse(recorder)
	DisallowNonValidatedFields(func(response *goyave.Response, r *goyave.Request) {
		result := recorder.Result()
		assert.Equal(t, 200, result.StatusCode)
	})(response, request)
}

func TestDisallowMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(DisallowMiddlewareTestSuite))
}
