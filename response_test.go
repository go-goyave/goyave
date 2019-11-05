package goyave

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/System-Glitch/goyave/config"
	"github.com/stretchr/testify/suite"
)

type ResponseTestSuite struct {
	suite.Suite
}

func (suite *ResponseTestSuite) SetupSuite() {
	config.LoadConfig()
}

func createTestResponse(rawRequest *http.Request) *Response {
	response := &Response{
		writer: httptest.NewRecorder(),
		empty:  true,
	}

	return response
}

func (suite *ResponseTestSuite) TestResponseStatus() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := createTestResponse(rawRequest)
	response.Status(403)
	resp := response.writer.(*httptest.ResponseRecorder).Result()

	suite.Equal(403, resp.StatusCode)
}

func (suite *ResponseTestSuite) TestResponseHeader() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := createTestResponse(rawRequest)
	response.Header().Set("Content-Type", "application/json")
	response.Status(200)
	resp := response.writer.(*httptest.ResponseRecorder).Result()

	suite.Equal(200, resp.StatusCode)
	suite.Equal("application/json", resp.Header.Get("Content-Type"))
}

func (suite *ResponseTestSuite) TestResponseError() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := createTestResponse(rawRequest)
	response.Error(fmt.Errorf("random error"))
	resp := response.writer.(*httptest.ResponseRecorder).Result()

	suite.Equal(500, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal("{\"error\":\"random error\"}\n", string(body))

	rawRequest = httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response = createTestResponse(rawRequest)
	response.Error("random error")
	resp = response.writer.(*httptest.ResponseRecorder).Result()

	suite.Equal(500, resp.StatusCode)

	body, err = ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal("{\"error\":\"random error\"}\n", string(body))
}

func (suite *ResponseTestSuite) TestResponseFile() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := createTestResponse(rawRequest)

	response.File("config/defaults.json")
	resp := response.writer.(*httptest.ResponseRecorder).Result()

	suite.Equal(200, resp.StatusCode)
	suite.Equal("inline", resp.Header.Get("Content-Disposition"))
	suite.Equal("text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
	suite.Equal("546", resp.Header.Get("Content-Length"))
}

func (suite *ResponseTestSuite) TestResponseFilePanic() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := createTestResponse(rawRequest)

	suite.Panics(func() {
		response.File("doesn'texist")
	})
}

func (suite *ResponseTestSuite) TestResponseDownload() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := createTestResponse(rawRequest)

	response.Download("config/defaults.json", "config.json")
	resp := response.writer.(*httptest.ResponseRecorder).Result()

	suite.Equal(200, resp.StatusCode)
	suite.Equal("attachment; filename=\"config.json\"", resp.Header.Get("Content-Disposition"))
	suite.Equal("text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
	suite.Equal("546", resp.Header.Get("Content-Length"))
}

func TestResponseTestSuite(t *testing.T) {
	suite.Run(t, new(ResponseTestSuite))
}
