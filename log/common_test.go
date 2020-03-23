package log

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/System-Glitch/goyave/v2"
)

type CommonTestSuite struct {
	goyave.TestSuite
}

func (suite *CommonTestSuite) TestCommonFormatter() {
	ts, _ := time.Parse("2006-01-02T15:04:05.000Z", "2020-03-23T13:58:26.371Z")
	response := suite.CreateTestResponse(httptest.NewRecorder())
	response.Status(http.StatusNoContent)
	request := suite.CreateTestRequest(httptest.NewRequest("GET", "/log", nil))
	suite.Equal("192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 204 5", CommonLogFormatter(ts, response, request, make([]byte, 5)))

	request = suite.CreateTestRequest(httptest.NewRequest("GET", "http://user@localhost/log", nil))
	suite.Equal("192.0.2.1 - user [23/Mar/2020:13:58:26 +0000] \"GET \"http://user@localhost/log\" HTTP/1.1\" 204 5", CommonLogFormatter(ts, response, request, make([]byte, 5)))

	// Invalid ipv6 URL
	request = suite.CreateTestRequest(httptest.NewRequest("GET", "/log", nil))
	request.Request().RemoteAddr = "[::1"
	suite.Equal("[::1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 204 5", CommonLogFormatter(ts, response, request, make([]byte, 5)))

	// HTTP 2
	request = suite.CreateTestRequest(httptest.NewRequest("CONNECT", "/log", nil))
	request.Request().Proto = "HTTP/2.0"
	request.Request().ProtoMajor = 2
	suite.Equal("192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"CONNECT \"example.com\" HTTP/2.0\" 204 5", CommonLogFormatter(ts, response, request, make([]byte, 5)))

	// No request uri
	request = suite.CreateTestRequest(httptest.NewRequest("GET", "/log", nil))
	request.Request().RequestURI = ""
	suite.Equal("192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 204 5", CommonLogFormatter(ts, response, request, make([]byte, 5)))
}

func (suite *CommonTestSuite) TestCombinedFormatter() {
	ts, _ := time.Parse("2006-01-02T15:04:05.000Z", "2020-03-23T13:58:26.371Z")
	response := suite.CreateTestResponse(httptest.NewRecorder())
	response.Status(http.StatusNoContent)
	request := suite.CreateTestRequest(httptest.NewRequest("GET", "/log", nil))

	referrer := "http://example.com"
	userAgent := "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:74.0) Gecko/20100101 Firefox/74.0"
	request.Header().Set("Referer", referrer)
	request.Header().Set("User-Agent", userAgent)
	suite.Equal("192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 204 5 \""+referrer+"\" \""+userAgent+"\"", CombinedLogFormatter(ts, response, request, make([]byte, 5)))
}

func TestCommonSuite(t *testing.T) {
	goyave.RunTest(t, new(CommonTestSuite))
}
