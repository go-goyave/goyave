package goyave

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"syscall"
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
}

func (suite *GoyaveTestSuite) loadConfig() {
	if err := config.Load(); err != nil {
		suite.FailNow(err.Error())
	}
	config.Set("tlsKey", "resources/server.key")
	config.Set("tlsCert", "resources/server.crt")
}

func (suite *GoyaveTestSuite) runServer(routeRegistrer func(*Router), callback func()) {
	c := make(chan bool, 1)
	c2 := make(chan bool, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	RegisterStartupHook(func() {
		callback()
		Stop()
		ClearStartupHooks()
		c <- true
	})

	go func() {
		Start(routeRegistrer)
		c2 <- true
	}()

	select {
	case <-ctx.Done():
		suite.Fail("Timeout exceeded in runServer")
	case <-c:
		fmt.Println("Shutdown OK")
	}
	<-c2
}

func (suite *GoyaveTestSuite) TestGetAddress() {
	suite.loadConfig()
	suite.Equal("127.0.0.1:1235", getAddress("http"))
	suite.Equal("127.0.0.1:1236", getAddress("https"))
}

func (suite *GoyaveTestSuite) TestStartStopServer() {
	config.Clear()
	proc, err := os.FindProcess(os.Getpid())
	if err == nil {
		c := make(chan bool, 1)
		c2 := make(chan bool, 1)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		RegisterStartupHook(func() {
			suite.True(IsReady())
			if runtime.GOOS == "windows" {
				fmt.Println("Testing on a windows machine. Cannot test proc signals")
				Stop()
			} else {
				fmt.Println("send sig")
				proc.Signal(syscall.SIGTERM)
				time.Sleep(10 * time.Millisecond)
				for IsReady() { // TODO: Ready is still true (may be ok)
					time.Sleep(10 * time.Millisecond)
					proc.Signal(syscall.SIGTERM)
				}
			}
			c <- true
		})
		go func() {
			Start(func(router *Router) {})
			c2 <- true
		}()

		select {
		case <-ctx.Done():
			suite.Fail("Timeout exceeded in server start/stop test")
		case <-c:
			suite.False(IsReady())
			suite.Nil(server)
			suite.Nil(hookChannel)
			ClearStartupHooks()
		}
		<-c2
	} else {
		fmt.Println("WARNING: Couldn't get process PID, skipping SIGINT test")
	}
}

func (suite *GoyaveTestSuite) TestTLSServer() {
	suite.loadConfig()
	config.Set("protocol", "https")
	suite.runServer(func(router *Router) {
		router.Route("GET", "/hello", helloHandler, nil)
	}, func() {
		netClient := createHTTPClient()
		resp, err := netClient.Get("http://127.0.0.1:1235/hello")
		suite.Nil(err)
		if err != nil {
			fmt.Println(err)
		}

		suite.NotNil(resp)
		if resp != nil {
			suite.Equal(308, resp.StatusCode)

			body, err := ioutil.ReadAll(resp.Body)
			suite.Nil(err)
			suite.Equal("<a href=\"https://127.0.0.1:1236/hello\">Permanent Redirect</a>.\n\n", string(body))
		}

		resp, err = netClient.Get("https://127.0.0.1:1236/hello")
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

	config.Set("protocol", "http")
}

func (suite *GoyaveTestSuite) TestTLSRedirectServerError() {
	suite.loadConfig()
	c := make(chan bool)
	c2 := make(chan bool)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	blockingServer := &http.Server{
		Addr:    getAddress("http"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	}

	go func() {
		go func() {
			c2 <- true
			// Run a server using the same port.
			err := blockingServer.ListenAndServe()
			if err != http.ErrServerClosed {
				suite.Fail(err.Error())
			}
			c2 <- true
		}()
		<-c2
		config.Set("protocol", "https")
		suite.runServer(func(router *Router) {}, func() {})
		config.Set("protocol", "http")
		c <- true
	}()

	select {
	case <-ctx.Done():
		suite.Fail("Timeout exceeded in redirect server error test")
	case <-c:
		suite.False(IsReady())
		suite.Nil(redirectServer)
		suite.Nil(stopChannel)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	blockingServer.Shutdown(ctx)
	<-c2
}

func (suite *GoyaveTestSuite) TestStaticServing() {
	suite.runServer(func(router *Router) {
		router.Static("/resources", "resources", true)
	}, func() {
		netClient := createHTTPClient()
		resp, err := netClient.Get("http://127.0.0.1:1235/resources/nothing")
		suite.Nil(err)
		if err != nil {
			fmt.Println(err)
		}
		suite.NotNil(resp)
		if resp != nil {
			suite.Equal(404, resp.StatusCode)
		}

		resp, err = netClient.Get("http://127.0.0.1:1235/resources/lang/en-US/locale.json")
		suite.Nil(err)
		if err != nil {
			fmt.Println(err)
		}
		suite.NotNil(resp)
		if resp != nil {
			suite.Equal(200, resp.StatusCode)

			body, err := ioutil.ReadAll(resp.Body)
			suite.Nil(err)
			suite.Equal("{\n    \"disallow-non-validated-fields\": \"Non-validated fields are forbidden.\"\n}", string(body))
		}
	})
}

func (suite *GoyaveTestSuite) TestServerError() {
	suite.loadConfig()
	suite.testServerError("http")
	suite.testServerError("https")
}

func (suite *GoyaveTestSuite) testServerError(protocol string) {
	c := make(chan bool)
	c2 := make(chan bool)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var ln net.Listener

	go func() {
		go func() {

			// Run a server using the same port as Goyave, so Goyave fails to bind.
			if protocol != "https" {
				var err error
				ln, err = net.Listen("tcp", getAddress(protocol))
				if err != nil {
					suite.Fail(err.Error())
				}
				c2 <- true
			}
			c2 <- true
			c2 <- true
		}()
		<-c2
		config.Set("protocol", protocol)
		if protocol == "https" {
			// Invalid certificates
			config.Set("tlsKey", "doesntexist")
			config.Set("tlsCert", "doesntexist")
		}

		fmt.Println("test server error " + protocol)
		Start(func(router *Router) {}) // TODO: server already running
		config.Set("protocol", "http")
		c <- true
	}()

	select {
	case <-ctx.Done():
		suite.Fail("Timeout exceeded in server error test")
	case <-c:
		suite.False(IsReady())
		suite.Nil(server)
	}

	if protocol != "https" {
		ln.Close()
	}
	<-c2
}

func (suite *GoyaveTestSuite) TestServerAlreadyRunning() {
	suite.loadConfig()
	suite.runServer(func(router *Router) {}, func() {
		suite.Panics(func() {
			Start(func(router *Router) {})
		})
	})
}

func TestGoyaveTestSuite(t *testing.T) {
	suite.Run(t, new(GoyaveTestSuite))
}
