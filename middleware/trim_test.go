package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/System-Glitch/goyave/v2"
	"github.com/System-Glitch/goyave/v2/validation"
	"github.com/stretchr/testify/assert"
)

func TestTrimMiddleware(t *testing.T) {
	request := &goyave.Request{
		Data:   map[string]interface{}{"text": " \t  trimmed\n  \t"},
		Rules:  validation.RuleSet{},
		Params: map[string]string{},
	}
	recorder := httptest.NewRecorder()
	response := goyave.CreateTestResponse(recorder)
	Trim(func(response *goyave.Response, r *goyave.Request) {
		assert.Equal(t, "trimmed", r.String("text"))
	})(response, request)
}
