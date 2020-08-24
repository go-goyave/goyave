package middleware

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/System-Glitch/goyave/v3"
)

type GzipMiddlewareTestSuite struct {
	goyave.TestSuite
}

type closeableChildWriter struct {
	io.Writer
	closed bool
}

func (w *closeableChildWriter) Close() error {
	w.closed = true
	return nil
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
	writer, _ := gzip.NewWriterLevel(recorder, gzip.BestCompression)
	compressWriter := &gzipWriter{
		Writer:         writer,
		ResponseWriter: recorder,
	}
	if _, err := compressWriter.Write([]byte("hello world")); err != nil {
		panic(err)
	}
	compressWriter.Close()

	result := recorder.Result()
	reader, err := gzip.NewReader(result.Body)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}

	suite.Equal("hello world", string(body))
}

func (suite *GzipMiddlewareTestSuite) TestCloseChild() {
	closeableWriter := &closeableChildWriter{closed: false}
	suite.RunServer(func(router *goyave.Router) {
		router.Middleware(func(next goyave.Handler) goyave.Handler {
			return func(response *goyave.Response, r *goyave.Request) {
				closeableWriter.Writer = response.Writer()
				response.SetWriter(closeableWriter)
				next(response, r)
			}
		})
		router.Middleware(Gzip())
		router.Route("GET", "/test", func(response *goyave.Response, r *goyave.Request) {
			response.String(http.StatusOK, "hello world")
		})
	}, func() {
		resp, err := suite.Get("/test", nil)
		if err != nil {
			suite.Fail(err.Error())
		}
		resp.Body.Close()
		suite.True(closeableWriter.closed)
	})
}

func (suite *GzipMiddlewareTestSuite) TestGzipMiddlewareInvalidLevel() {
	suite.Panics(func() { GzipLevel(-3) })
	suite.Panics(func() { GzipLevel(10) })
}

func TestGzipMiddlewareTestSuite(t *testing.T) {
	goyave.RunTest(t, new(GzipMiddlewareTestSuite))
}
