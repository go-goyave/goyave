package goyave

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/lang"
	"github.com/stretchr/testify/assert"
)

type CustomTestSuite struct {
	TestSuite
}

type FailingTestSuite struct {
	TestSuite
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
		resp, err := suite.getHTTPClient().Get("http://127.0.0.1:1235/hello") // TODO will be replace with helpers Get/Post/...
		suite.Nil(err)
		if err != nil {
			fmt.Println(err)
		}

		suite.NotNil(resp)
		if resp != nil {
			suite.Equal(200, resp.StatusCode)

			body, err := ioutil.ReadAll(resp.Body)
			suite.Nil(err)
			suite.Equal("Hi!", string(body))
		}
	})
}

func (suite *CustomTestSuite) testRunServerTimeout() { // TOTO re-enable this test
	suite.SetTimeout(time.Second)
	suite.RunServer(func(router *Router) {}, func() {
		time.Sleep(suite.Timeout() + 1*time.Second)
		suite.True(suite.T().Failed()) // TODO How to un-fail this test? Or run in fake testing.T
	})
	suite.SetTimeout(5 * time.Second)
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
