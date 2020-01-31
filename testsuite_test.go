package goyave

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/database"
	"github.com/System-Glitch/goyave/v2/helper/filesystem"
	"github.com/System-Glitch/goyave/v2/lang"
	"github.com/stretchr/testify/assert"
)

type CustomTestSuite struct {
	TestSuite
}

type FailingTestSuite struct {
	TestSuite
}

type TestModel struct {
	ID   uint   `gorm:"primary_key"`
	Name string `gorm:"type:varchar(100)"`
}

func genericHandler(message string) func(response *Response, request *Request) {
	return func(response *Response, request *Request) {
		response.String(http.StatusOK, message)
	}
}

func (suite *CustomTestSuite) TestEnv() {
	suite.Equal("test", os.Getenv("GOYAVE_ENV"))
	suite.Equal("test", config.GetString("environment"))
	suite.Equal("Malformed JSON", lang.Get("en-US", "malformed-json"))
}

func (suite *CustomTestSuite) TestRunServer() {
	suite.RunServer(func(router *Router) {
		router.Route("GET", "/hello", func(response *Response, request *Request) {
			response.String(http.StatusOK, "Hi!")
		}, nil)
	}, func() {
		resp, err := suite.Get("/hello", nil)
		suite.Nil(err)
		if err != nil {
			fmt.Println(err)
		}

		suite.NotNil(resp)
		if resp != nil {
			suite.Equal(200, resp.StatusCode)
			suite.Equal("Hi!", string(suite.GetBody(resp)))
		}
	})
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

	suite.Equal(418, result.StatusCode)
}

func (suite *CustomTestSuite) TestRequests() {
	suite.RunServer(func(router *Router) {
		router.Route("GET", "/get", genericHandler("get"), nil)
		router.Route("POST", "/post", genericHandler("post"), nil)
		router.Route("PUT", "/put", genericHandler("put"), nil)
		router.Route("PATCH", "/patch", genericHandler("patch"), nil)
		router.Route("DELETE", "/delete", genericHandler("delete"), nil)
		router.Route("GET", "/headers", func(response *Response, request *Request) {
			response.String(http.StatusOK, request.Header().Get("Accept-Language"))
		}, nil)
	}, func() {
		resp, err := suite.Get("/get", nil)
		suite.Nil(err)
		if err == nil {
			suite.Equal("get", string(suite.GetBody(resp)))
		}
		resp, err = suite.Get("/post", nil)
		suite.Nil(err)
		if err == nil {
			suite.Equal(http.StatusMethodNotAllowed, resp.StatusCode)
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
		}

		// Errors
		resp, err = suite.Get("invalid", nil)
		suite.NotNil(err)
		suite.Nil(resp)

	})
}

func (suite *CustomTestSuite) TestJSON() {
	suite.RunServer(func(router *Router) {
		router.Route("GET", "/invalid", genericHandler("get"), nil)
		router.Route("GET", "/get", func(response *Response, request *Request) {
			response.JSON(http.StatusOK, map[string]interface{}{
				"field":  "value",
				"number": 42,
			})
		}, nil)
	}, func() {
		resp, err := suite.Get("/get", nil)
		suite.Nil(err)
		if err == nil {
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
		}, nil)
	}, func() {
		resp, err := suite.Get("/get", nil)
		suite.Nil(err)
		if err == nil {
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
	err := ioutil.WriteFile("test-file.txt", []byte("test-content"), 0644)
	if err != nil {
		panic(err)
	}
	defer filesystem.Delete("test-file.txt")
	files := suite.CreateTestFiles("test-file.txt")
	suite.Equal(1, len(files))
	suite.Equal("test-file.txt", files[0].Header.Filename)
	body, err := ioutil.ReadAll(files[0].Data)
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
	err := ioutil.WriteFile(path, []byte("test-content"), 0644)
	if err != nil {
		panic(err)
	}
	defer filesystem.Delete(path)

	suite.RunServer(func(router *Router) {
		router.Route("POST", "/post", func(response *Response, request *Request) {
			content, err := ioutil.ReadAll(request.File("file")[0].Data)
			if err != nil {
				panic(err)
			}
			response.JSON(http.StatusOK, map[string]interface{}{
				"file":  string(content),
				"field": request.String("field"),
			})
		}, nil)
	}, func() {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		suite.WriteFile(writer, path, "file", filepath.Base(path))
		suite.WriteField(writer, "field", "hello world")
		err := writer.Close()
		if err != nil {
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
	config.Set("dbConnection", "mysql")
	db := database.GetConnection()
	db.AutoMigrate(&TestModel{})

	for i := 0; i < 5; i++ {
		db.Create(&TestModel{Name: fmt.Sprintf("Test %d", i)})
	}
	count := 0
	db.Model(&TestModel{}).Count(&count)
	suite.Equal(5, count)

	database.RegisterModel(&TestModel{})
	suite.ClearDatabase()
	database.ClearRegisteredModels()

	db.Model(&TestModel{}).Count(&count)
	suite.Equal(0, count)

	db.Exec("DROP TABLE test_models;")
	config.Set("dbConnection", "none")
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
	os.Rename("config.test.json", "config.test.json.bak")
	mockT := new(testing.T)
	RunTest(mockT, new(FailingTestSuite))
	assert.True(t, mockT.Failed())
	os.Rename("config.test.json.bak", "config.test.json")
}
