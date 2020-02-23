package middleware

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/System-Glitch/goyave/v2"
)

type GzipMiddlewareTestSuite struct {
	goyave.TestSuite
}

func (suite *GzipMiddlewareTestSuite) TestGzipMiddleware() {
	handler := func(response *goyave.Response, r *goyave.Request) {
		response.String(http.StatusOK, "hello world")
	}
	rawRequest := httptest.NewRequest("GET", "/", nil)
	request := suite.CreateTestRequest(rawRequest)
	result := suite.Middleware(Gzip(), request, handler)
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal("hello world", string(body)) // Not compressed

	rawRequest = httptest.NewRequest("GET", "/", nil)
	rawRequest.Header.Set("Accept-Encoding", "gzip")
	request = suite.CreateTestRequest(rawRequest)
	result = suite.Middleware(Gzip(), request, handler)
	suite.Equal("gzip", result.Header.Get("Content-Encoding"))

	reader, err := gzip.NewReader(result.Body)
	if err != nil {
		panic(err)
	}
	body, err = ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	suite.Equal("hello world", string(body))
}

func (suite *GzipMiddlewareTestSuite) TestCloseNonCloseable() {
	rawRequest := httptest.NewRequest("GET", "/", nil)
	rawRequest.Header.Set("Accept-Encoding", "gzip")
	recorder := httptest.NewRecorder()
	compressWriter := &gzipWriter{
		Writer:         recorder,
		ResponseWriter: recorder,
	}
	compressWriter.Write([]byte("hello world"))
	compressWriter.Close()

	result := recorder.Result()
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal("hello world", string(body))
}

func (suite *GzipMiddlewareTestSuite) TestGzipMiddlewareInvalidLevel() {
	suite.Panics(func() { GzipLevel(-3) })
	suite.Panics(func() { GzipLevel(10) })
}

func TestGzipMiddlewareTestSuite(t *testing.T) {
	goyave.RunTest(t, new(GzipMiddlewareTestSuite))
}
