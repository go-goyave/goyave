package goyave

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/helper/filesystem"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type GoyaveTestSuite struct {
	TestSuite
}

func helloHandler(response *Response, request *Request) {
	response.String(http.StatusOK, "Hi!")
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

func (suite *GoyaveTestSuite) TestGetHost() {
	suite.loadConfig()
	suite.Equal("127.0.0.1:1235", getHost("http"))
	suite.Equal("127.0.0.1:1236", getHost("https"))
}

func (suite *GoyaveTestSuite) TestGetAddress() {
	suite.loadConfig()
	suite.Equal("http://127.0.0.1:1235", getAddress("http"))
	suite.Equal("https://127.0.0.1:1236", getAddress("https"))

	config.Set("domain", "test.system-glitch.me")
	suite.Equal("http://test.system-glitch.me:1235", getAddress("http"))
	suite.Equal("https://test.system-glitch.me:1236", getAddress("https"))

	config.Set("port", 80.0)
	config.Set("httpsPort", 443.0)
	suite.Equal("http://test.system-glitch.me", getAddress("http"))
	suite.Equal("https://test.system-glitch.me", getAddress("https"))
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
				for IsReady() {
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
	suite.RunServer(func(router *Router) {
		router.Route("GET", "/hello", helloHandler, nil)
	}, func() {
		netClient := suite.getHTTPClient()
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

	go func() {
		go func() {
			// Run a server using the same port.
			ln, err := net.Listen("tcp", getHost("http"))
			if err != nil {
				suite.Fail(err.Error())
				return
			}
			c2 <- true
			<-c2
			ln.Close()
			c2 <- true
		}()
		<-c2
		config.Set("protocol", "https")
		suite.RunServer(func(router *Router) {}, func() {})
		config.Set("protocol", "http")
		c2 <- true
		<-c2
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
}

func (suite *GoyaveTestSuite) TestStaticServing() {
	suite.RunServer(func(router *Router) {
		router.Static("/resources", "resources", true)
	}, func() {
		netClient := suite.getHTTPClient()
		resp, err := netClient.Get("http://127.0.0.1:1235/resources/nothing")
		suite.Nil(err)
		if err != nil {
			fmt.Println(err)
		}
		suite.NotNil(resp)
		if resp != nil {
			suite.Equal(404, resp.StatusCode)
		}

		err = ioutil.WriteFile("resources/lang/en-US/test-file.txt", []byte("test-content"), 0644)
		if err != nil {
			panic(err)
		}
		defer filesystem.Delete("resources/lang/en-US/test-file.txt")
		resp, err = netClient.Get("http://127.0.0.1:1235/resources/lang/en-US/test-file.txt")
		suite.Nil(err)
		if err != nil {
			fmt.Println(err)
		}
		suite.NotNil(resp)
		if resp != nil {
			suite.Equal(200, resp.StatusCode)

			body, err := ioutil.ReadAll(resp.Body)
			suite.Nil(err)
			suite.Equal("test-content", string(body))
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
				ln, err = net.Listen("tcp", getHost(protocol))
				if err != nil {
					suite.Fail(err.Error())
				}
				c2 <- true
			} else {
				c2 <- true
			}
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
		Start(func(router *Router) {})
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
	suite.RunServer(func(router *Router) {}, func() {
		suite.Panics(func() {
			Start(func(router *Router) {})
		})
	})
}

func (suite *GoyaveTestSuite) TestMaintenanceMode() {
	suite.loadConfig()
	suite.RunServer(func(router *Router) {
		router.Route("GET", "/hello", helloHandler, nil)
	}, func() {
		EnableMaintenance()
		suite.True(IsMaintenanceEnabled())

		netClient := suite.getHTTPClient()
		resp, err := netClient.Get("http://127.0.0.1:1235/hello")
		suite.Nil(err)
		if err != nil {
			fmt.Println(err)
		}

		suite.NotNil(resp)
		if resp != nil {
			suite.Equal(503, resp.StatusCode)
		}

		DisableMaintenance()
		suite.False(IsMaintenanceEnabled())

		resp, err = netClient.Get("http://127.0.0.1:1235/hello")
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

	config.Set("maintenance", true)
	suite.RunServer(func(router *Router) {
		router.Route("GET", "/hello", helloHandler, nil)
	}, func() {
		suite.True(IsMaintenanceEnabled())

		netClient := suite.getHTTPClient()
		resp, err := netClient.Get("http://127.0.0.1:1235/hello")
		suite.Nil(err)
		if err != nil {
			fmt.Println(err)
		}

		suite.NotNil(resp)
		if resp != nil {
			suite.Equal(503, resp.StatusCode)
		}

		DisableMaintenance()

		suite.False(IsMaintenanceEnabled())

		resp, err = netClient.Get("http://127.0.0.1:1235/hello")
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
	config.Set("maintenance", false)
}

func (suite *GoyaveTestSuite) TestAutoMigrate() {
	suite.loadConfig()
	config.Set("dbConnection", "mysql")
	config.Set("dbAutoMigrate", true)
	suite.RunServer(func(router *Router) {}, func() {})
	config.Set("dbAutoMigrate", false)
	config.Set("dbConnection", "none")
}

func TestGoyaveTestSuite(t *testing.T) {
	RunTest(t, new(GoyaveTestSuite))
}
