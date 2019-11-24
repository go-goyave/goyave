package goyave

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/System-Glitch/goyave/config"
	"github.com/stretchr/testify/suite"
)

type NativeHandlerTestSuite struct {
	suite.Suite
}

func (suite *NativeHandlerTestSuite) SetupSuite() {
	config.Load()
}

func (suite *NativeHandlerTestSuite) TestNativeHandler() {
	request := &Request{
		httpRequest: httptest.NewRequest("GET", "/native", nil),
	}
	recorder := httptest.NewRecorder()
	response := &Response{
		ResponseWriter: recorder,
		empty:          true,
	}
	handler := NativeHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world"))
	}))

	handler(response, request)
	result := recorder.Result()
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(200, result.StatusCode)
	suite.Equal("Hello world", string(body))
	suite.False(response.empty)
}
func (suite *NativeHandlerTestSuite) TestNativeMiddleware() {
	request := &Request{
		httpRequest: httptest.NewRequest("GET", "/native", nil),
	}
	recorder := httptest.NewRecorder()
	response := &Response{
		ResponseWriter: recorder,
		empty:          true,
	}
	middleware := NativeMiddleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello world"))
			next.ServeHTTP(w, r)
		})
	})

	handlerExecuted := false
	handler := func(response *Response, r *Request) {
		result := recorder.Result()
		body, err := ioutil.ReadAll(result.Body)
		if err != nil {
			panic(err)
		}
		suite.Equal(200, result.StatusCode)
		suite.Equal("Hello world", string(body))
		suite.False(response.empty)
		handlerExecuted = true
	}
	middleware(handler)(response, request)
	suite.True(handlerExecuted)
}

func TestNativeHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(NativeHandlerTestSuite))
}
