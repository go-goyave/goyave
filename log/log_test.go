package log

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/System-Glitch/goyave/v2"
)

type LogMiddlewareTestSuite struct {
	goyave.TestSuite
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

func (suite *LogMiddlewareTestSuite) TestWrite() {
	// now := time.Now()
	request := suite.CreateTestRequest(httptest.NewRequest("GET", "/log", nil))

	result := suite.Middleware(CommonLogMiddleware(), request, func(response *goyave.Response, request *goyave.Request) {
		response.String(http.StatusOK, "message")
	})

	suite.Equal(200, result.StatusCode)

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal("message", string(body))

	// TODO test log output
	// suite.Equal("192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 204 5", CommonLogFormatter(ts, response, request, make([]byte, 5)))
}

func TestLogMiddlewareSuite(t *testing.T) {
	goyave.RunTest(t, new(LogMiddlewareTestSuite))
}
