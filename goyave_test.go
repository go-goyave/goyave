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
	suite.SetTimeout(5 * time.Second)
}

func (suite *GoyaveTestSuite) loadConfig() {
	if err := config.Load(); err != nil {
		suite.FailNow(err.Error())
	}
	config.Set("server.tls.key", "resources/server.key")
	config.Set("server.tls.cert", "resources/server.crt")
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

	config.Set("server.domain", "test.system-glitch.me")
	suite.Equal("http://test.system-glitch.me:1235", getAddress("http"))
	suite.Equal("https://test.system-glitch.me:1236", getAddress("https"))

	config.Set("server.port", 80.0)
	config.Set("server.httpsPort", 443.0)
	suite.Equal("http://test.system-glitch.me", getAddress("http"))
	suite.Equal("https://test.system-glitch.me", getAddress("https"))
}

func (suite *GoyaveTestSuite) TestStartStopServer() {
	config.Clear()
	proc, err := os.FindProcess(os.Getpid())
	if err == nil {
		c := make(chan bool, 1)
		c2 := make(chan bool, 1)
		ctx, cancel := context.WithTimeout(context.Background(), suite.Timeout())
		defer cancel()

		RegisterStartupHook(func() {
			suite.True(IsReady())
			if runtime.GOOS == "windows" {
				fmt.Println("Testing on a windows machine. Cannot test proc signals")
				Stop()
			} else {
				if err := proc.Signal(syscall.SIGTERM); err != nil {
					suite.Fail(err.Error())
				}
				time.Sleep(10 * time.Millisecond)
				for IsReady() {
					time.Sleep(10 * time.Millisecond)
					if err := proc.Signal(syscall.SIGTERM); err != nil {
						suite.Fail(err.Error())
					}
				}
			}
			c <- true
		})
		go func() {
			if err := Start(func(router *Router) {}); err != nil {
				suite.Fail(err.Error())
			}
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
	config.Set("server.protocol", "https")
	suite.RunServer(func(router *Router) {
		router.Route("GET", "/hello", helloHandler)
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
			resp.Body.Close()
			suite.Nil(err)
			suite.Equal("<a href=\"https://127.0.0.1:1236/hello\">Permanent Redirect</a>.\n\n", string(body))
		}

		resp, err = netClient.Get("http://127.0.0.1:1235/hello?param=1")
		suite.Nil(err)
		if err != nil {
			fmt.Println(err)
		}

		suite.NotNil(resp)
		if resp != nil {
			suite.Equal(308, resp.StatusCode)

			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			suite.Nil(err)
			suite.Equal("<a href=\"https://127.0.0.1:1236/hello?param=1\">Permanent Redirect</a>.\n\n", string(body))
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
			resp.Body.Close()
			suite.Nil(err)
			suite.Equal("Hi!", string(body))
		}
	})

	config.Set("server.protocol", "http")
}

func (suite *GoyaveTestSuite) TestTLSRedirectServerError() {
	suite.loadConfig()
	c := make(chan bool)
	c2 := make(chan bool)
	ctx, cancel := context.WithTimeout(context.Background(), suite.Timeout())
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
		config.Set("server.protocol", "https")
		suite.RunServer(func(router *Router) {}, func() {})
		config.Set("server.protocol", "http")
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

		err = ioutil.WriteFile("resources/template/test-static-serve.txt", []byte("test-content"), 0644)
		if err != nil {
			panic(err)
		}
		defer filesystem.Delete("resources/template/test-static-serve.txt")
		resp, err = netClient.Get("http://127.0.0.1:1235/resources/template/test-static-serve.txt")
		suite.Nil(err)
		if err != nil {
			fmt.Println(err)
		}
		suite.NotNil(resp)
		if resp != nil {
			suite.Equal(200, resp.StatusCode)

			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
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
	c := make(chan error)
	c2 := make(chan bool)
	ctx, cancel := context.WithTimeout(context.Background(), suite.Timeout())
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
		config.Set("server.protocol", protocol)
		if protocol == "https" {
			// Invalid certificates
			config.Set("server.tls.key", "doesntexist")
			config.Set("server.tls.cert", "doesntexist")
		}

		fmt.Println("test server error " + protocol)
		err := Start(func(router *Router) {})
		config.Set("server.protocol", "http")
		c <- err
	}()

	select {
	case <-ctx.Done():
		suite.Fail("Timeout exceeded in server error test")
	case err := <-c:
		suite.False(IsReady())
		suite.Nil(server)
		suite.NotNil(err)
		if protocol == "https" {
			suite.Equal(ExitHTTPError, err.(*Error).ExitCode)
		} else {
			suite.Equal(ExitNetworkError, err.(*Error).ExitCode)
		}
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
			if err := Start(func(router *Router) {}); err != nil {
				suite.Fail(err.Error())
			}
		})
	})
}

func (suite *GoyaveTestSuite) TestMaintenanceMode() {
	suite.loadConfig()
	suite.RunServer(func(router *Router) {
		router.Route("GET", "/hello", helloHandler)
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
			resp.Body.Close()
			suite.Nil(err)
			suite.Equal("Hi!", string(body))
		}
	})

	config.Set("server.maintenance", true)
	suite.RunServer(func(router *Router) {
		router.Route("GET", "/hello", helloHandler)
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
			resp.Body.Close()
			suite.Nil(err)
			suite.Equal("Hi!", string(body))
		}
	})
	config.Set("server.maintenance", false)
}

func (suite *GoyaveTestSuite) TestAutoMigrate() {
	suite.loadConfig()
	config.Set("database.connection", "mysql")
	config.Set("database.autoMigrate", true)
	suite.RunServer(func(router *Router) {}, func() {})
	config.Set("database.autoMigrate", false)
	config.Set("database.Connection", "none")
}

func (suite *GoyaveTestSuite) TestError() {
	err := &Error{ExitHTTPError, fmt.Errorf("test error")}
	suite.Equal("test error", err.Error())
}

func (suite *GoyaveTestSuite) TestConfigError() {
	config.Clear()
	if err := os.Chdir("config"); err != nil {
		panic(err)
	}
	defer os.Chdir("..")

	os.Setenv("GOYAVE_ENV", "test_invalid")
	defer os.Setenv("GOYAVE_ENV", "test")

	c := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), suite.Timeout())
	defer cancel()

	go func() {
		c <- Start(func(r *Router) {})
	}()

	select {
	case <-ctx.Done():
		suite.Fail("Timeout exceeded in Goyave test suite TestConfigError")
	case err := <-c:
		suite.NotNil(err)
		if err != nil {
			e := err.(*Error)
			suite.Equal(ExitInvalidConfig, e.ExitCode)
			suite.Equal("Invalid config:\n\t- \"app.environment\" type must be string", e.Error())
		}
	}
}

func TestGoyaveTestSuite(t *testing.T) {
	RunTest(t, new(GoyaveTestSuite))
}
