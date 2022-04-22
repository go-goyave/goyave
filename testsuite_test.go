package goyave

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/database"
	"goyave.dev/goyave/v4/lang"
	"goyave.dev/goyave/v4/util/fsutil"
)

type CustomTestSuite struct {
	TestSuite
}

type MigratingTestSuite struct {
	TestSuite
}

type FailingTestSuite struct {
	TestSuite
}

type ConcurrentTestSuite struct {
	res *int
	TestSuite
}

type TestModel struct {
	Name string `gorm:"type:varchar(100)"`
	ID   uint   `gorm:"primaryKey"`
}

type TestViewModel struct {
	database.View
	Name string `gorm:"type:varchar(100)"`
	ID   uint   `gorm:"primaryKey"`
}

func genericHandler(message string) func(response *Response, request *Request) {
	return func(response *Response, request *Request) {
		response.String(http.StatusOK, message)
	}
}

func (suite *CustomTestSuite) TestEnv() {
	suite.Equal("test", os.Getenv("GOYAVE_ENV"))
	suite.Equal("test", config.GetString("app.environment"))
	suite.Equal("Malformed JSON", lang.Get("en-US", "malformed-json"))
}

func (suite *CustomTestSuite) TestCreateTestResponse() {
	writer := httptest.NewRecorder()
	response := suite.CreateTestResponse(writer)
	suite.Equal(writer, response.writer)
	suite.Equal(writer, response.responseWriter)

	rawRequest := httptest.NewRequest("POST", "/test-route", strings.NewReader("body"))
	response = suite.CreateTestResponseWithRequest(writer, rawRequest)
	suite.Equal(writer, response.writer)
	suite.Equal(writer, response.responseWriter)
	suite.Equal(rawRequest, response.httpRequest)
}

func (suite *CustomTestSuite) TestCreateTestRequest() {
	request := suite.CreateTestRequest(nil)
	suite.Nil(request.route)
	suite.Nil(request.Data)
	suite.Nil(request.Rules)
	suite.Equal("en-US", request.Lang)
	suite.NotNil(request.Params)
	suite.NotNil(request.Extra)
	suite.NotNil(request.httpRequest)
	suite.Equal("GET", request.httpRequest.Method)
	suite.Equal("/", request.httpRequest.RequestURI)

	rawRequest := httptest.NewRequest("POST", "/test-route", nil)
	request = suite.CreateTestRequest(rawRequest)
	suite.Equal("POST", request.httpRequest.Method)
	suite.Equal("/test-route", request.httpRequest.RequestURI)
}

func (suite *CustomTestSuite) TestRunServer() {
	suite.RunServer(func(router *Router) {
		router.Route("GET", "/hello", func(response *Response, request *Request) {
			response.String(http.StatusOK, "Hi!")
		})
	}, func() {
		resp, err := suite.Get("/hello", nil)
		suite.Nil(err)
		if err != nil {
			suite.Fail(err.Error())
		}

		suite.NotNil(resp)
		if resp != nil {
			defer resp.Body.Close()
			suite.Equal(200, resp.StatusCode)
			suite.Equal("Hi!", string(suite.GetBody(resp)))
		}
	})
	suite.Empty(startupHooks)
}

func (suite *CustomTestSuite) TestRunServerTimeout() {
	suite.SetTimeout(time.Second)
	oldT := suite.T()
	suite.SetT(new(testing.T))
	suite.RunServer(func(router *Router) {}, func() {
		time.Sleep(suite.Timeout() + 1*time.Second)
	})
	assert.True(oldT, suite.T().Failed())
	suite.SetTimeout(5 * time.Second)
	suite.SetT(oldT)
	suite.Empty(startupHooks)
}

func (suite *CustomTestSuite) TestRunServerError() {
	config.Clear()
	oldT := suite.T()
	suite.SetT(new(testing.T))
	prevEnv := os.Getenv("GOYAVE_ENV")
	if err := os.Setenv("GOYAVE_ENV", "notanenv"); err != nil {
		suite.Fail(err.Error())
	}
	suite.RunServer(func(router *Router) {}, func() {})
	assert.True(oldT, suite.T().Failed())
	suite.SetT(oldT)
	if err := os.Setenv("GOYAVE_ENV", prevEnv); err != nil {
		suite.Fail(err.Error())
	}
	suite.Empty(startupHooks)
}

func (suite *CustomTestSuite) TestMiddleware() {
	rawRequest := httptest.NewRequest("GET", "/test-route", nil)
	rawRequest.Header.Set("Content-Type", "application/json")
	request := suite.CreateTestRequest(rawRequest)

	result := suite.Middleware(func(next Handler) Handler {
		return func(response *Response, request *Request) {
			response.Status(http.StatusTeapot)
			next(response, request)
		}
	}, request, func(response *Response, request *Request) {
		suite.Equal("application/json", request.Header().Get("Content-Type"))
	})

	result.Body.Close()
	suite.Equal(418, result.StatusCode)
}

func (suite *CustomTestSuite) TestRequests() {
	suite.RunServer(func(router *Router) {
		router.Route("GET", "/get", genericHandler("get"))
		router.Route("POST", "/post", genericHandler("post"))
		router.Route("PUT", "/put", genericHandler("put"))
		router.Route("PATCH", "/patch", genericHandler("patch"))
		router.Route("DELETE", "/delete", genericHandler("delete"))
		router.Route("GET", "/headers", func(response *Response, request *Request) {
			response.String(http.StatusOK, request.Header().Get("Accept-Language"))
		})
	}, func() {
		resp, err := suite.Get("/get", nil)
		suite.Nil(err)
		if err == nil {
			suite.Equal("get", string(suite.GetBody(resp)))
			resp.Body.Close()
		}
		resp, err = suite.Get("/post", nil)
		suite.Nil(err)
		if err == nil {
			suite.Equal(http.StatusMethodNotAllowed, resp.StatusCode)
			resp.Body.Close()
		}
		resp, err = suite.Get("/nonexistent-route", nil)
		suite.Nil(err)
		if err == nil {
			suite.Equal(http.StatusNotFound, resp.StatusCode)
			resp.Body.Close()
		}
		resp, err = suite.Post("/post", nil, strings.NewReader("field=value"))
		suite.Nil(err)
		if err == nil {
			suite.Equal("post", string(suite.GetBody(resp)))
		}
		resp, err = suite.Put("/put", nil, strings.NewReader("field=value"))
		suite.Nil(err)
		if err == nil {
			suite.Equal("put", string(suite.GetBody(resp)))
		}
		resp, err = suite.Patch("/patch", nil, strings.NewReader("field=value"))
		suite.Nil(err)
		if err == nil {
			suite.Equal("patch", string(suite.GetBody(resp)))
		}
		resp, err = suite.Delete("/delete", nil, strings.NewReader("field=value"))
		suite.Nil(err)
		if err == nil {
			suite.Equal("delete", string(suite.GetBody(resp)))
		}

		// Headers
		resp, err = suite.Get("/headers", map[string]string{"Accept-Language": "en-US"})
		suite.Nil(err)
		if err == nil {
			suite.Equal("en-US", string(suite.GetBody(resp)))
			resp.Body.Close()
		}

		// Errors
		resp, err = suite.Get("invalid", nil)
		suite.NotNil(err)
		suite.Nil(resp)
		if resp != nil {
			resp.Body.Close()
		}
	})
}

func (suite *CustomTestSuite) TestJSON() {
	suite.RunServer(func(router *Router) {
		router.Route("GET", "/invalid", genericHandler("get"))
		router.Route("GET", "/get", func(response *Response, request *Request) {
			response.JSON(http.StatusOK, map[string]interface{}{
				"field":  "value",
				"number": 42,
			})
		})
	}, func() {
		resp, err := suite.Get("/get", nil)
		suite.Nil(err)
		if err == nil {
			defer resp.Body.Close()
			json := map[string]interface{}{}
			err := suite.GetJSONBody(resp, &json)
			suite.Nil(err)
			if err == nil {
				suite.Equal("value", json["field"])
				suite.Equal(float64(42), json["number"])
			}
		}

		resp, err = suite.Get("/invalid", nil)
		suite.Nil(err)
		if err == nil {
			defer resp.Body.Close()
			oldT := suite.T()
			suite.SetT(new(testing.T))
			json := map[string]interface{}{}
			err := suite.GetJSONBody(resp, &json)
			assert.True(oldT, suite.T().Failed())
			suite.SetT(oldT)
			suite.NotNil(err)
		}
	})
}

func (suite *CustomTestSuite) TestJSONSlice() {
	suite.RunServer(func(router *Router) {
		router.Route("GET", "/get", func(response *Response, request *Request) {
			response.JSON(http.StatusOK, []map[string]interface{}{
				{"field": "value", "number": 42},
				{"field": "other value", "number": 12},
			})
		})
	}, func() {
		resp, err := suite.Get("/get", nil)
		suite.Nil(err)
		if err == nil {
			defer resp.Body.Close()
			json := []map[string]interface{}{}
			err := suite.GetJSONBody(resp, &json)
			suite.Nil(err)
			suite.Len(json, 2)
			if err == nil {
				suite.Equal("value", json[0]["field"])
				suite.Equal(float64(42), json[0]["number"])
				suite.Equal("other value", json[1]["field"])
				suite.Equal(float64(12), json[1]["number"])
			}
		}
	})
}

func (suite *CustomTestSuite) TestCreateTestFiles() {
	err := os.WriteFile("test-file.txt", []byte("test-content"), 0644)
	if err != nil {
		panic(err)
	}
	defer fsutil.Delete("test-file.txt")
	files := suite.CreateTestFiles("test-file.txt")
	suite.Equal(1, len(files))
	suite.Equal("test-file.txt", files[0].Header.Filename)
	body, err := io.ReadAll(files[0].Data)
	if err != nil {
		panic(err)
	}
	suite.Equal("test-content", string(body))

	oldT := suite.T()
	suite.SetT(new(testing.T))
	files = suite.CreateTestFiles("doesn't exist")
	assert.True(oldT, suite.T().Failed())
	suite.SetT(oldT)
	suite.Equal(0, len(files))
}

func (suite *CustomTestSuite) TestMultipartForm() {
	const path = "test-file.txt"
	err := os.WriteFile(path, []byte("test-content"), 0644)
	if err != nil {
		panic(err)
	}
	defer fsutil.Delete(path)

	suite.RunServer(func(router *Router) {
		router.Route("POST", "/post", func(response *Response, request *Request) {
			content, err := io.ReadAll(request.File("file")[0].Data)
			if err != nil {
				panic(err)
			}
			response.JSON(http.StatusOK, map[string]interface{}{
				"file":  string(content),
				"field": request.String("field"),
			})
		})
	}, func() {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		suite.WriteFile(writer, path, "file", filepath.Base(path))
		suite.WriteField(writer, "field", "hello world")
		if err := writer.Close(); err != nil {
			panic(err)
		}
		resp, err := suite.Post("/post", map[string]string{"Content-Type": writer.FormDataContentType()}, body)
		suite.Nil(err)
		if err == nil {
			json := map[string]interface{}{}
			err := suite.GetJSONBody(resp, &json)
			suite.Nil(err)
			if err == nil {
				suite.Equal("test-content", json["file"])
				suite.Equal("hello world", json["field"])
			}
		}
	})
}

func (suite *CustomTestSuite) TestClearDatabase() {
	config.Set("database.connection", "mysql")
	db := database.GetConnection()
	db.AutoMigrate(&TestModel{})

	for i := 0; i < 5; i++ {
		db.Create(&TestModel{Name: fmt.Sprintf("Test %d", i)})
	}
	count := int64(0)
	db.Model(&TestModel{}).Count(&count)
	suite.Equal(int64(5), count)

	database.RegisterModel(&TestModel{})
	suite.ClearDatabase()
	database.ClearRegisteredModels()

	db.Model(&TestModel{}).Count(&count)
	suite.Equal(int64(0), count)

	db.Migrator().DropTable(&TestModel{})
	config.Set("database.connection", "none")
}

func (suite *CustomTestSuite) TestClearDatabaseView() {
	config.Set("database.connection", "mysql")
	db := database.GetConnection()
	db.AutoMigrate(&TestViewModel{})

	for i := 0; i < 5; i++ {
		db.Create(&TestViewModel{Name: fmt.Sprintf("Test %d", i)})
	}
	defer func() {
		if err := db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&TestViewModel{}).Error; err != nil {
			panic(err)
		}
	}()
	count := int64(0)
	db.Model(&TestViewModel{}).Count(&count)
	suite.Equal(int64(5), count)

	database.RegisterModel(&TestViewModel{})
	suite.ClearDatabase()
	database.ClearRegisteredModels()

	db.Model(&TestViewModel{}).Count(&count)
	suite.Equal(int64(5), count)
}

func (suite *CustomTestSuite) TestClearDatabaseTables() {
	config.Set("database.connection", "mysql")
	db := database.GetConnection()
	db.AutoMigrate(&TestModel{})

	database.RegisterModel(&TestModel{})
	suite.ClearDatabaseTables()
	database.ClearRegisteredModels()

	found := false
	rows, err := db.Raw("SHOW TABLES;").Rows()
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		name := ""
		if err := rows.Scan(&name); err != nil {
			panic(err)
		}
		if name == "test_models" {
			found = true
		}
	}

	suite.False(found)

	config.Set("database.connection", "none")
}

func TestConcurrentSuiteExecution(t *testing.T) { // Suites should not execute in parallel
	// This test is only useful if the race detector is enabled
	res := 0
	suite1 := new(ConcurrentTestSuite)
	suite2 := new(ConcurrentTestSuite)
	suite1.res = &res
	suite2.res = &res

	c := make(chan bool, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		// Executing this ten times almost guarantees
		// there WILL be a race condition.
		wg.Add(2)
		go func() {
			defer wg.Done()
			RunTest(t, suite1)
		}()
		go func() {
			defer wg.Done()
			RunTest(t, suite2)
		}()
	}

	go func() {
		wg.Wait()
		c <- true
	}()

	select {
	case <-ctx.Done():
		assert.Fail(t, "Timeout exceeded in concurrent suites test")
	case val := <-c:
		assert.True(t, val)
	}

}

func (suite *ConcurrentTestSuite) TestExecutionOrder() {
	*suite.res++
}

func TestTestSuite(t *testing.T) {
	suite := new(CustomTestSuite)
	RunTest(t, suite)
	assert.Equal(t, 5*time.Second, suite.Timeout())
}

func (s *FailingTestSuite) TestRunServerTimeout() {
	s.RunServer(func(router *Router) {}, func() {
		time.Sleep(s.Timeout() + 1)
	})
}

func TestTestSuiteFail(t *testing.T) {
	if err := os.Rename("config.test.json", "config.test.json.bak"); err != nil {
		panic(err)
	}
	defer func() {
		if err := os.Rename("config.test.json.bak", "config.test.json"); err != nil {
			panic(err)
		}
	}()
	mockT := new(testing.T)
	RunTest(mockT, new(FailingTestSuite))
	assert.True(t, mockT.Failed())
}

func (suite *MigratingTestSuite) TearDownSuite() {
	suite.ClearDatabaseTables()
}

func TestMigrate(t *testing.T) {
	if err := config.LoadFrom("resources/config.migration-test.json"); err != nil {
		assert.Fail(t, "Failed to load config", err)
		return
	}
	suite := new(MigratingTestSuite)
	RunTest(t, suite)
}
