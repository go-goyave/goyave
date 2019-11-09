package goyave

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/System-Glitch/goyave/config"
	"github.com/stretchr/testify/suite"
)

type GoyaveTestSuite struct {
	suite.Suite
}

func helloHandler(response *Response, request *Request) {
	response.String(http.StatusOK, "Hi!")
}

func createHTTPClient() *http.Client {
	config := &tls.Config{
		InsecureSkipVerify: true, // TODO add test self-signed certificate to rootCA pool
	}

	return &http.Client{
		Timeout:   time.Second * 5,
		Transport: &http.Transport{TLSClientConfig: config},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func (suite *GoyaveTestSuite) SetupSuite() {
	os.Setenv("GOYAVE_ENV", "test")
	config.LoadConfig()
}

func (suite *GoyaveTestSuite) runServer(routeRegistrer func(*Router), callback func()) {
	c := make(chan bool, 1)

	RegisterStartupHook(func() {
		callback()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		Stop(ctx)
		ClearStartupHooks()
		c <- true
	})
	Start(routeRegistrer)
	<-c
}

func (suite *GoyaveTestSuite) TestGetAddress() {
	suite.Equal("127.0.0.1:1235", getAddress("http"))
	suite.Equal("127.0.0.1:1236", getAddress("https"))
}

func (suite *GoyaveTestSuite) TestStartStopServer() {
	proc, err := os.FindProcess(os.Getpid())
	if err == nil {
		c := make(chan bool, 1)

		RegisterStartupHook(func() {
			suite.True(IsReady())
			proc.Signal(os.Interrupt)
			time.Sleep(500 * time.Millisecond)
			suite.False(IsReady())
			suite.Nil(server)
			ClearStartupHooks()
			c <- true
		})
		Start(func(router *Router) {})
		<-c
	} else {
		fmt.Println("WARNING: Couldn't get process PID, skipping SIGINT test")
	}
}

func (suite *GoyaveTestSuite) TestTLSServer() {
	config.Set("protocol", "https")
	config.Set("tlsKey", "resources/server.key")
	config.Set("tlsCert", "resources/server.crt")
	suite.runServer(func(router *Router) {
		router.Route("GET", "/hello", helloHandler, nil)
	}, func() {
		netClient := createHTTPClient()
		resp, err := netClient.Get("http://127.0.0.1:1235/hello")
		suite.Nil(err)
		suite.NotNil(resp)
		if resp != nil {
			suite.Equal(308, resp.StatusCode)

			body, err := ioutil.ReadAll(resp.Body)
			suite.Nil(err)
			suite.Equal("<a href=\"https://127.0.0.1:1236/hello\">Permanent Redirect</a>.\n\n", string(body))
		}
	})

	suite.runServer(func(router *Router) {
		router.Route("GET", "/hello", helloHandler, nil)
	}, func() {
		netClient := createHTTPClient()
		resp, err := netClient.Get("https://127.0.0.1:1236/hello")
		suite.Nil(err)
		suite.NotNil(resp)
		if resp != nil {
			suite.Equal(200, resp.StatusCode)

			body, err := ioutil.ReadAll(resp.Body)
			suite.Nil(err)
			suite.Equal("Hi!", string(body))
		}
	})

	config.Set("protocol", "http")
}

func (suite *GoyaveTestSuite) TestStaticServing() {
	suite.runServer(func(router *Router) {
		router.Static("/resources", "resources", true)
	}, func() {
		netClient := createHTTPClient()
		resp, err := netClient.Get("http://127.0.0.1:1235/resources/nothing")
		suite.Nil(err)
		suite.NotNil(resp)
		if resp != nil {
			suite.Equal(404, resp.StatusCode)
		}

		resp, err = netClient.Get("http://127.0.0.1:1235/resources/lang/en-US/locale.json")
		suite.Nil(err)
		suite.NotNil(resp)
		if resp != nil {
			suite.Equal(200, resp.StatusCode)

			body, err := ioutil.ReadAll(resp.Body)
			suite.Nil(err)
			suite.Equal("{\n    \"disallow-non-validated-fields\": \"Non-validated fields are forbidden.\"\n}", string(body))
		}
	})
}

func TestGoyaveTestSuite(t *testing.T) {
	suite.Run(t, new(GoyaveTestSuite))
}
