package goyave

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/database"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
)

type ResponseTestSuite struct {
	suite.Suite
	previousEnv string
}

func (suite *ResponseTestSuite) SetupSuite() {
	suite.previousEnv = os.Getenv("GOYAVE_ENV")
	os.Setenv("GOYAVE_ENV", "test")
	if err := config.Load(); err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *ResponseTestSuite) TestResponseStatus() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := newResponse(httptest.NewRecorder(), rawRequest)
	response.Status(403)
	resp := response.responseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(200, resp.StatusCode) // Not written yet
	suite.True(response.empty)
	suite.Equal(403, response.GetStatus())

	rawRequest = httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response = newResponse(httptest.NewRecorder(), rawRequest)
	response.String(403, "test")
	resp = response.responseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(403, resp.StatusCode)
	suite.False(response.empty)
	suite.Equal(403, response.GetStatus())

	rawRequest = httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response = newResponse(httptest.NewRecorder(), rawRequest)
	response.Status(403)
	response.Status(200) // Should have no effect
	resp = response.responseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(200, resp.StatusCode) // Not written yet
	suite.True(response.empty)
	suite.Equal(403, response.GetStatus())
}

func (suite *ResponseTestSuite) TestResponseHeader() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := newResponse(httptest.NewRecorder(), rawRequest)
	response.Header().Set("Content-Type", "application/json")
	response.Status(200)
	resp := response.responseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(200, resp.StatusCode)
	suite.Equal("application/json", resp.Header.Get("Content-Type"))
	suite.True(response.empty)
	suite.Equal(200, response.status)
}

func (suite *ResponseTestSuite) TestResponseError() {
	config.Set("debug", true)
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := newResponse(httptest.NewRecorder(), rawRequest)
	response.Error(fmt.Errorf("random error"))
	resp := response.responseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(500, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal("{\"error\":\"random error\"}\n", string(body))
	suite.False(response.empty)
	suite.Equal(500, response.status)
	suite.NotNil(response.err)

	rawRequest = httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response = newResponse(httptest.NewRecorder(), rawRequest)
	response.Error("random error")
	resp = response.responseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(500, resp.StatusCode)
	suite.NotNil(response.err)

	body, err = ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal("{\"error\":\"random error\"}\n", string(body))
	suite.False(response.empty)
	suite.Equal(500, response.status)
	suite.NotNil(response.err)

	config.Set("debug", false)
	rawRequest = httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response = newResponse(httptest.NewRecorder(), rawRequest)
	response.Error("random error")
	resp = response.responseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(500, response.GetStatus())

	body, err = ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Empty("", string(body))
	suite.True(response.empty)
	suite.Equal("random error", response.GetError())
	suite.Equal(500, response.status)
	config.Set("debug", true)
}

func (suite *ResponseTestSuite) TestResponseFile() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := newResponse(httptest.NewRecorder(), rawRequest)

	response.File("config/config.test.json")
	resp := response.responseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(200, resp.StatusCode)
	suite.Equal("inline", resp.Header.Get("Content-Disposition"))
	suite.Equal("application/json", resp.Header.Get("Content-Type"))
	suite.Equal("29", resp.Header.Get("Content-Length"))
	suite.False(response.empty)
	suite.Equal(200, response.status)

	// Test no Content-Type override
	rawRequest = httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response = newResponse(httptest.NewRecorder(), rawRequest)
	response.Header().Set("Content-Type", "text/plain")
	response.File("config/config.test.json")
	resp = response.responseWriter.(*httptest.ResponseRecorder).Result()
	suite.Equal("text/plain", resp.Header.Get("Content-Type"))
}

func (suite *ResponseTestSuite) TestResponseFilePanic() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := newResponse(httptest.NewRecorder(), rawRequest)

	suite.Panics(func() {
		response.File("doesn'texist")
	})
}

func (suite *ResponseTestSuite) TestResponseDownload() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := newResponse(httptest.NewRecorder(), rawRequest)

	response.Download("config/config.test.json", "config.json")
	resp := response.responseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(200, resp.StatusCode)
	suite.Equal("attachment; filename=\"config.json\"", resp.Header.Get("Content-Disposition"))
	suite.Equal("application/json", resp.Header.Get("Content-Type"))
	suite.Equal("29", resp.Header.Get("Content-Length"))
	suite.False(response.empty)
	suite.Equal(200, response.status)
}

func (suite *ResponseTestSuite) TestResponseRedirect() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := newResponse(httptest.NewRecorder(), rawRequest)

	response.Redirect("https://www.google.com")
	resp := response.responseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(308, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal("<a href=\"https://www.google.com\">Permanent Redirect</a>.\n\n", string(body))
	suite.False(response.empty)
	suite.Equal(308, response.status)
}

func (suite *ResponseTestSuite) TestResponseTemporaryRedirect() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := newResponse(httptest.NewRecorder(), rawRequest)

	response.TemporaryRedirect("https://www.google.com")
	resp := response.responseWriter.(*httptest.ResponseRecorder).Result()

	suite.Equal(307, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal("<a href=\"https://www.google.com\">Temporary Redirect</a>.\n\n", string(body))
	suite.False(response.empty)
	suite.Equal(307, response.status)
}

func (suite *ResponseTestSuite) TestResponseCookie() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := newResponse(httptest.NewRecorder(), rawRequest)
	response.Cookie(&http.Cookie{
		Name:  "cookie-name",
		Value: "test",
	})

	resp := response.responseWriter.(*httptest.ResponseRecorder).Result()
	cookies := resp.Cookies()
	suite.Equal(1, len(cookies))
	suite.Equal("cookie-name", cookies[0].Name)
	suite.Equal("test", cookies[0].Value)
}

func (suite *ResponseTestSuite) TestResponseWrite() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	response := newResponse(httptest.NewRecorder(), rawRequest)
	response.Write([]byte("byte array"))
	resp := response.responseWriter.(*httptest.ResponseRecorder).Result()
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
		suite.Equal(recorder, response.responseWriter)
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

func (suite *ResponseTestSuite) TestHandleDatabaseError() {
	type TestRecord struct {
		gorm.Model
	}
	config.Set("dbConnection", "mysql")
	defer config.Set("dbConnection", "none")
	db := database.GetConnection()

	response := newResponse(httptest.NewRecorder(), nil)
	suite.False(response.HandleDatabaseError(db.Find(&TestRecord{})))

	suite.Equal(http.StatusInternalServerError, response.status)

	db.AutoMigrate(&TestRecord{})
	defer db.DropTable(&TestRecord{})
	response = newResponse(httptest.NewRecorder(), nil)
	suite.False(response.HandleDatabaseError(db.Where("id = ?", -1).Find(&TestRecord{})))

	suite.Equal(http.StatusNotFound, response.status)

	response = newResponse(httptest.NewRecorder(), nil)
	suite.True(response.HandleDatabaseError(db.Exec("SHOW TABLES;")))

	suite.Equal(0, response.status)

	response = newResponse(httptest.NewRecorder(), nil)
	results := []TestRecord{}
	suite.True(response.HandleDatabaseError(db.Find(&results))) // Get all but empty result should not be an error
	suite.Equal(0, response.status)
}

// ------------------------

type testWriter struct {
	result *string
	id     string
	io.Writer
	closed bool
}

func (w *testWriter) Write(b []byte) (int, error) {
	*w.result += w.id + string(b)
	return w.Writer.Write(b)
}

func (w *testWriter) Close() error {
	w.closed = true
	return fmt.Errorf("Test close error")
}

func (suite *ResponseTestSuite) TestChainedWriter() {
	writer := httptest.NewRecorder()
	response := newResponse(writer, nil)
	result := ""
	testWr := &testWriter{&result, "0", response.Writer(), false}
	response.SetWriter(testWr)

	response.String(http.StatusOK, "hello world")

	suite.Equal("0hello world", result)
	suite.Equal(200, response.status)
	suite.True(response.wroteHeader)
	suite.False(response.empty)

	suite.Equal("Test close error", response.close().Error())
	suite.True(testWr.closed)

	resp := writer.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	suite.Equal("hello world", string(body))

	// Test double chained writer
	writer = httptest.NewRecorder()
	response = newResponse(writer, nil)
	result = ""
	testWr = &testWriter{&result, "0", response.Writer(), false}
	testWr2 := &testWriter{&result, "1", testWr, false}
	response.SetWriter(testWr2)

	response.String(http.StatusOK, "hello world")
	suite.Equal("1hello world0hello world", result)
	suite.Equal(200, response.status)
	suite.True(response.wroteHeader)
	suite.False(response.empty)
	resp = writer.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	suite.Equal("hello world", string(body))
}

func (suite *ResponseTestSuite) TearDownAllSuite() {
	os.Setenv("GOYAVE_ENV", suite.previousEnv)
}

func TestResponseTestSuite(t *testing.T) {
	suite.Run(t, new(ResponseTestSuite))
}
