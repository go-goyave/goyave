package goyave

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/stretchr/testify/suite"
)

type ResponseTestSuite struct {
	suite.Suite
}

func (suite *ResponseTestSuite) SetupSuite() {
	config.Load()
}

func createTestResponse(rawRequest *http.Request) *Response {
	response := &Response{
		ResponseWriter: httptest.NewRecorder(),
		httpRequest:    rawRequest,
		empty:          true,
		emptyStatus:    true,
	}

	return response
}

func (suite *ResponseTestSuite) TestResponseStatus() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := createTestResponse(rawRequest)
	response.Status(403)
	resp := response.ResponseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(403, resp.StatusCode)
	suite.True(response.empty)
	suite.False(response.emptyStatus)

	rawRequest = httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response = createTestResponse(rawRequest)
	response.String(403, "test")
	resp = response.ResponseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(403, resp.StatusCode)
	suite.False(response.empty)
	suite.False(response.emptyStatus)
}

func (suite *ResponseTestSuite) TestResponseHeader() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := createTestResponse(rawRequest)
	response.Header().Set("Content-Type", "application/json")
	response.Status(200)
	resp := response.ResponseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(200, resp.StatusCode)
	suite.Equal("application/json", resp.Header.Get("Content-Type"))
	suite.True(response.empty)
	suite.False(response.emptyStatus)
}

func (suite *ResponseTestSuite) TestResponseError() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := createTestResponse(rawRequest)
	response.Error(fmt.Errorf("random error"))
	resp := response.ResponseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(500, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal("{\"error\":\"random error\"}\n", string(body))
	suite.False(response.empty)
	suite.False(response.emptyStatus)

	rawRequest = httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response = createTestResponse(rawRequest)
	response.Error("random error")
	resp = response.ResponseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(500, resp.StatusCode)

	body, err = ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal("{\"error\":\"random error\"}\n", string(body))
	suite.False(response.empty)
	suite.False(response.emptyStatus)
}

func (suite *ResponseTestSuite) TestResponseFile() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := createTestResponse(rawRequest)

	response.File("config/config.test.json")
	resp := response.ResponseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(200, resp.StatusCode)
	suite.Equal("inline", resp.Header.Get("Content-Disposition"))
	suite.Equal("application/json", resp.Header.Get("Content-Type"))
	suite.Equal("29", resp.Header.Get("Content-Length"))
	suite.False(response.empty)
	suite.False(response.emptyStatus)
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

	response.Download("config/config.test.json", "config.json")
	resp := response.ResponseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(200, resp.StatusCode)
	suite.Equal("attachment; filename=\"config.json\"", resp.Header.Get("Content-Disposition"))
	suite.Equal("application/json", resp.Header.Get("Content-Type"))
	suite.Equal("29", resp.Header.Get("Content-Length"))
	suite.False(response.empty)
	suite.False(response.emptyStatus)
}

func (suite *ResponseTestSuite) TestResponseRedirect() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := createTestResponse(rawRequest)

	response.Redirect("https://www.google.com")
	resp := response.ResponseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(308, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal("<a href=\"https://www.google.com\">Permanent Redirect</a>.\n\n", string(body))
	suite.False(response.empty)
	suite.False(response.emptyStatus)
}

func (suite *ResponseTestSuite) TestResponseTemporaryRedirect() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := createTestResponse(rawRequest)

	response.TemporaryRedirect("https://www.google.com")
	resp := response.ResponseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(307, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal("<a href=\"https://www.google.com\">Temporary Redirect</a>.\n\n", string(body))
	suite.False(response.empty)
	suite.False(response.emptyStatus)
}

func (suite *ResponseTestSuite) TestResponseCookie() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := createTestResponse(rawRequest)
	response.Cookie(&http.Cookie{
		Name:  "cookie-name",
		Value: "test",
	})

	resp := response.ResponseWriter.(*httptest.ResponseRecorder).Result()
	cookies := resp.Cookies()
	suite.Equal(1, len(cookies))
	suite.Equal("cookie-name", cookies[0].Name)
	suite.Equal("test", cookies[0].Value)
}

func (suite *ResponseTestSuite) TestResponseWrite() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := createTestResponse(rawRequest)
	response.Write([]byte("byte array"))
	resp := response.ResponseWriter.(*httptest.ResponseRecorder).Result()
	body, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal("byte array", string(body))
	suite.False(response.empty)
}

func (suite *ResponseTestSuite) TestCreateTestResponse() {
	recorder := httptest.NewRecorder()
	response := CreateTestResponse(recorder)
	suite.NotNil(response)
	if response != nil {
		suite.Equal(recorder, response.ResponseWriter)
	}
}

func (suite *ResponseTestSuite) TestRender() {
	// With map data
	recorder := httptest.NewRecorder()
	response := CreateTestResponse(recorder)

	mapData := map[string]interface{}{
		"Status":  http.StatusNotFound,
		"Message": "Not Found.",
	}
	suite.Nil(response.Render(http.StatusNotFound, "error.txt", mapData))
	resp := recorder.Result()
	suite.Equal(404, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal("Error 404: Not Found.", string(body))

	// With struct data
	recorder = httptest.NewRecorder()
	response = CreateTestResponse(recorder)

	structData := struct {
		Status  int
		Message string
	}{
		Status:  http.StatusNotFound,
		Message: "Not Found.",
	}
	suite.Nil(response.Render(http.StatusNotFound, "error.txt", structData))
	resp = recorder.Result()
	suite.Equal(404, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal("Error 404: Not Found.", string(body))

	// Non-existing template and exec error
	recorder = httptest.NewRecorder()
	response = CreateTestResponse(recorder)

	suite.NotNil(response.Render(http.StatusNotFound, "non-existing-template", nil))
	suite.NotNil(response.Render(http.StatusNotFound, "invalid.txt", nil))
}

func (suite *ResponseTestSuite) TestRenderHTML() {
	// With map data
	recorder := httptest.NewRecorder()
	response := CreateTestResponse(recorder)

	mapData := map[string]interface{}{
		"Status":  http.StatusNotFound,
		"Message": "Not Found.",
	}
	suite.Nil(response.RenderHTML(http.StatusNotFound, "error.html", mapData))
	resp := recorder.Result()
	suite.Equal(404, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal("<html>\n    <head></head>\n    <body>\n        <p>Error 404: Not Found.</p>\n    </body>\n</html>", string(body))

	// With struct data
	recorder = httptest.NewRecorder()
	response = CreateTestResponse(recorder)

	structData := struct {
		Status  int
		Message string
	}{
		Status:  http.StatusNotFound,
		Message: "Not Found.",
	}
	suite.Nil(response.RenderHTML(http.StatusNotFound, "error.html", structData))
	resp = recorder.Result()
	suite.Equal(404, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal("<html>\n    <head></head>\n    <body>\n        <p>Error 404: Not Found.</p>\n    </body>\n</html>", string(body))

	// Non-existing template and exec error
	recorder = httptest.NewRecorder()
	response = CreateTestResponse(recorder)

	suite.NotNil(response.RenderHTML(http.StatusNotFound, "non-existing-template", nil))
	suite.NotNil(response.RenderHTML(http.StatusNotFound, "invalid.txt", nil))
}

func TestResponseTestSuite(t *testing.T) {
	suite.Run(t, new(ResponseTestSuite))
}
