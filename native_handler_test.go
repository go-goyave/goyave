package goyave

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type NativeHandlerTestSuite struct {
	TestSuite
}

func (suite *NativeHandlerTestSuite) TestNativeHandler() {
	request := &Request{
		httpRequest: httptest.NewRequest("GET", "/native", nil),
	}
	recorder := httptest.NewRecorder()
	response := newResponse(recorder, nil)
	handler := NativeHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world"))
	}))

	handler(response, request)
	result := recorder.Result()
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	result.Body.Close()
	suite.Equal(200, result.StatusCode)
	suite.Equal("Hello world", string(body))
	suite.False(response.empty)
}

func (suite *NativeHandlerTestSuite) TestNativeHandlerBody() {

	handler := NativeHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		suite.Equal("request=content", string(res))
		w.WriteHeader(http.StatusNoContent)
	}))

	suite.RunServer(func(router *Router) {
		router.Route("POST", "/native", handler)
	}, func() {
		headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded; param=value"}
		resp, err := suite.Post("/native", headers, strings.NewReader("request=content"))
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		suite.Equal(http.StatusNoContent, resp.StatusCode)
	})
}

func (suite *NativeHandlerTestSuite) TestNativeHandlerBodyJSON() {

	handler := NativeHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		suite.Equal("{\"request\":\"content\"}", string(res))
		w.WriteHeader(http.StatusNoContent)
	}))

	suite.RunServer(func(router *Router) {
		router.Route("POST", "/native", handler)
	}, func() {
		headers := map[string]string{"Content-Type": "application/json"}
		resp, err := suite.Post("/native", headers, strings.NewReader("{\"request\":\"content\"}"))
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		suite.Equal(http.StatusNoContent, resp.StatusCode)
	})
}

func (suite *NativeHandlerTestSuite) TestNativeMiddleware() {
	request := &Request{
		httpRequest: httptest.NewRequest("GET", "/native", nil),
	}
	recorder := httptest.NewRecorder()
	response := newResponse(recorder, nil)
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
		result.Body.Close()
		suite.Equal(200, result.StatusCode)
		suite.Equal("Hello world", string(body))
		suite.False(response.empty)
		handlerExecuted = true
	}
	middleware(handler)(response, request)
	suite.True(handlerExecuted)

	middleware = NativeMiddleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Blocking middleware
		})
	})

	handlerExecuted = false
	handler = func(response *Response, r *Request) {
		handlerExecuted = true
	}
	middleware(handler)(response, request)
	suite.False(handlerExecuted)
}

func (suite *NativeHandlerTestSuite) TestNativeMiddlewareReplacesRequest() {
	request := &Request{
		httpRequest: httptest.NewRequest("GET", "/native", nil),
	}
	recorder := httptest.NewRecorder()
	response := newResponse(recorder, nil)
	var requestWithContext *http.Request
	middleware := NativeMiddleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			type key int
			ctx := context.WithValue(r.Context(), key(0), "value")
			requestWithContext = r.WithContext(ctx)
			next.ServeHTTP(w, requestWithContext)
		})
	})

	handlerExecuted := false
	handler := func(response *Response, r *Request) {
		suite.Same(requestWithContext, r.Request())
		handlerExecuted = true
	}
	middleware(handler)(response, request)
	suite.True(handlerExecuted)
}

func TestNativeHandlerTestSuite(t *testing.T) {
	RunTest(t, new(NativeHandlerTestSuite))
}
