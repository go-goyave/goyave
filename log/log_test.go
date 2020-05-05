package log

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/System-Glitch/goyave/v2"
)

type LogMiddlewareTestSuite struct {
	goyave.TestSuite
}

type testWriter struct {
	io.Writer
	closed bool
}

func (w *testWriter) Close() error {
	w.closed = true
	return fmt.Errorf("Test close error")
}

func (suite *LogMiddlewareTestSuite) TestNewWriter() {
	now := time.Now()
	recorder := httptest.NewRecorder()
	response := suite.CreateTestResponse(recorder)
	request := suite.CreateTestRequest(httptest.NewRequest("GET", "/log", nil))
	writer := NewWriter(response, request, CommonLogFormatter)

	suite.Equal(now.Format("2006-01-02T15:04"), writer.now.Format("2006-01-02T15:04"))
	suite.Equal(request, writer.request)
	suite.Equal(response, writer.response)
	suite.Equal(recorder, writer.writer)
}

func (suite *LogMiddlewareTestSuite) TestCommonWrite() {
	buffer := bytes.NewBufferString("")
	goyave.AccessLogger.SetOutput(buffer)
	defer func() {
		goyave.AccessLogger.SetOutput(os.Stdout)
	}()

	now := time.Now()
	request := suite.CreateTestRequest(httptest.NewRequest("GET", "/log", nil))

	result := suite.Middleware(CommonLogMiddleware(), request, func(response *goyave.Response, request *goyave.Request) {
		response.Writer().(*Writer).now = now
		response.String(http.StatusOK, "message")
	})

	suite.Equal(200, result.StatusCode)

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal("message", string(body))

	suite.Equal("192.0.2.1 - - ["+now.Format(TimestampFormat)+"] \"GET \"/log\" HTTP/1.1\" 200 7\n", buffer.String())
}

func (suite *LogMiddlewareTestSuite) TestCombinedWrite() {
	buffer := bytes.NewBufferString("")
	goyave.AccessLogger.SetOutput(buffer)
	defer func() {
		goyave.AccessLogger.SetOutput(os.Stdout)
	}()

	now := time.Now()
	request := suite.CreateTestRequest(httptest.NewRequest("GET", "/log", nil))
	referrer := "http://example.com"
	userAgent := "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:74.0) Gecko/20100101 Firefox/74.0"
	request.Header().Set("Referer", referrer)
	request.Header().Set("User-Agent", userAgent)

	result := suite.Middleware(CombinedLogMiddleware(), request, func(response *goyave.Response, request *goyave.Request) {
		response.Writer().(*Writer).now = now
		response.String(http.StatusOK, "message")
	})

	suite.Equal(200, result.StatusCode)

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal("message", string(body))

	suite.Equal("192.0.2.1 - - ["+now.Format(TimestampFormat)+"] \"GET \"/log\" HTTP/1.1\" 200 7 \""+referrer+"\" \""+userAgent+"\"\n", buffer.String())
}

func (suite *LogMiddlewareTestSuite) TestCloseChildWriter() {
	closeableWriter := &testWriter{closed: false}
	suite.RunServer(func(router *goyave.Router) {
		router.Middleware(func(next goyave.Handler) goyave.Handler {
			return func(response *goyave.Response, r *goyave.Request) {
				closeableWriter.Writer = response.Writer()
				response.SetWriter(closeableWriter)
				next(response, r)
			}
		})
		router.Middleware(CombinedLogMiddleware())
		router.Route("GET", "/test", func(response *goyave.Response, request *goyave.Request) {
			response.String(http.StatusOK, "message")
		}, nil)
	}, func() {
		resp, err := suite.Get("/test", nil)
		if err != nil {
			suite.Fail(err.Error())
		}
		resp.Body.Close()
		suite.True(closeableWriter.closed)
	})
}

func TestLogMiddlewareSuite(t *testing.T) {
	goyave.RunTest(t, new(LogMiddlewareTestSuite))
}
