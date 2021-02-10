package websocket

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/System-Glitch/goyave/v3"
	"github.com/System-Glitch/goyave/v3/config"
	ws "github.com/gorilla/websocket"
)

type WebsocketTestSuite struct {
	goyave.TestSuite
	previousTimeout int
}

func (suite *WebsocketTestSuite) SetupSuite() {
	suite.previousTimeout = config.GetInt("server.timeout")
	config.Set("server.timeout", 1)
}

func (suite *WebsocketTestSuite) TearDownSuite() {
	config.Set("server.timeout", suite.previousTimeout)
}

func (suite *WebsocketTestSuite) echoWSHandler(c *Conn, request *goyave.Request) error {
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			if IsCloseError(err) {
				return nil
			}
			err := fmt.Errorf("read: %v", err)
			suite.Error(err)
			return err
		}
		goyave.Logger.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			err := fmt.Errorf("write: %v", err)
			suite.Error(err)
			return err
		}
	}
}

func (suite *WebsocketTestSuite) TestIsCloseError() {
	suite.True(IsCloseError(&ws.CloseError{Code: ws.CloseNormalClosure}))
	suite.True(IsCloseError(&ws.CloseError{Code: ws.CloseGoingAway}))
	suite.True(IsCloseError(&ws.CloseError{Code: ws.CloseNoStatusReceived}))
	suite.False(IsCloseError(&ws.CloseError{Code: ws.CloseAbnormalClosure}))
	suite.False(IsCloseError(&ws.CloseError{Code: ws.CloseProtocolError}))
}

func (suite *WebsocketTestSuite) TestAdapterOnError() {
	req := suite.CreateTestRequest(httptest.NewRequest("GET", "/websocket", nil))
	recorder := httptest.NewRecorder()
	resp := suite.CreateTestResponse(recorder)
	reasonErr := fmt.Errorf("test adapter error")
	executed := false
	a := adapter{
		upgradeErrorHandler: func(response *goyave.Response, request *goyave.Request, status int, reason error) {
			suite.Equal(req, request)
			suite.Equal(resp, response)
			suite.Equal(http.StatusBadRequest, status)
			suite.Equal(reasonErr, reason)
			executed = true
		},
		request: req,
	}

	a.onError(resp, req.Request(), http.StatusBadRequest, reasonErr)

	suite.True(executed)

	suite.Panics(func() {
		a.onError(resp, req.Request(), http.StatusInternalServerError, reasonErr)
	})

}

func (suite *WebsocketTestSuite) TestDefaultUpgradeErrorHandler() {
	req := suite.CreateTestRequest(httptest.NewRequest("GET", "/websocket", nil))
	recorder := httptest.NewRecorder()
	resp := suite.CreateTestResponse(recorder)
	reasonErr := fmt.Errorf("test upgrade error handler")

	previousDebug := config.GetBool("app.debug")
	config.Set("app.debug", false)
	defer config.Set("app.debug", previousDebug)
	defaultUpgradeErrorHandler(resp, req, http.StatusBadRequest, reasonErr)

	result := recorder.Result()

	suite.Equal("application/json; charset=utf-8", result.Header.Get("Content-Type"))
	suite.Equal(http.StatusBadRequest, result.StatusCode)

	json := map[string]string{}
	err := suite.GetJSONBody(result, &json)
	result.Body.Close()
	suite.Nil(err)
	if err != nil {
		suite.Equal(http.StatusText(http.StatusBadRequest), json["error"])
	}

	recorder = httptest.NewRecorder()
	resp = suite.CreateTestResponse(recorder)

	config.Set("app.debug", false)
	defaultUpgradeErrorHandler(resp, req, http.StatusBadRequest, reasonErr)

	result = recorder.Result()

	suite.Equal("application/json; charset=utf-8", result.Header.Get("Content-Type"))
	suite.Equal(http.StatusBadRequest, result.StatusCode)

	json = map[string]string{}
	err = suite.GetJSONBody(result, &json)
	result.Body.Close()
	suite.Nil(err)
	if err != nil {
		suite.Equal(reasonErr.Error(), json["error"])
	}
	config.Set("app.debug", true)
	defaultUpgradeErrorHandler(resp, req, http.StatusBadRequest, reasonErr)
}

func (suite *WebsocketTestSuite) TestAdapterCheckOrigin() {
	req := suite.CreateTestRequest(httptest.NewRequest("GET", "/websocket", nil))
	a := adapter{
		request: req,
	}

	suite.Nil(a.getCheckOriginFunc())

	executed := false
	a.checkOrigin = func(r *goyave.Request) bool {
		suite.Equal(req, r)
		executed = true
		return true
	}

	f := a.getCheckOriginFunc()
	suite.NotNil(f)
	suite.True(f(req.Request()))
	suite.True(executed)
}

func (suite *WebsocketTestSuite) TestMakeUpgrader() {
	upgrader := Upgrader{}

	req := suite.CreateTestRequest(httptest.NewRequest("GET", "/websocket", nil))
	u := upgrader.makeUpgrader(req)

	suite.Equal(upgrader.Settings.HandshakeTimeout, u.HandshakeTimeout)
	suite.Equal(upgrader.Settings.ReadBufferSize, u.ReadBufferSize)
	suite.Equal(upgrader.Settings.WriteBufferSize, u.WriteBufferSize)
	suite.Equal(upgrader.Settings.WriteBufferPool, u.WriteBufferPool)
	suite.Equal(upgrader.Settings.Subprotocols, u.Subprotocols)
	suite.Equal(upgrader.Settings.EnableCompression, u.EnableCompression)
	suite.NotNil(u.Error)
	suite.Nil(u.CheckOrigin)

	upgrader.Settings.EnableCompression = true
	u = upgrader.makeUpgrader(req)
	suite.Equal(upgrader.Settings.EnableCompression, u.EnableCompression)
}

func (suite *WebsocketTestSuite) TestUpgrade() {
	routeURL := ""
	suite.RunServer(func(r *goyave.Router) {
		upgrader := Upgrader{}
		route := r.Get("/websocket", upgrader.Handler(suite.echoWSHandler))
		routeURL = "ws" + strings.TrimPrefix(route.BuildURL(), config.GetString("server.protocol"))
	}, func() {
		// TODO document testing websocket
		conn, _, err := ws.DefaultDialer.Dial(routeURL, nil)
		if err != nil {
			suite.Error(err)
			return
		}
		defer conn.Close()

		message := []byte("hello world")
		suite.Nil(conn.WriteMessage(ws.TextMessage, message))

		messageType, data, err := conn.ReadMessage()
		suite.Nil(err)
		suite.Equal(ws.TextMessage, messageType)
		suite.Equal(message, data)

		m := ws.FormatCloseMessage(ws.CloseNormalClosure, "Connection closed by client")
		suite.Nil(conn.WriteControl(ws.CloseMessage, m, time.Now().Add(time.Second)))
	})
}

func (suite *WebsocketTestSuite) TestUpgradeError() {
	suite.RunServer(func(r *goyave.Router) {
		upgrader := Upgrader{}
		r.Get("/websocket", upgrader.Handler(suite.echoWSHandler))
	}, func() {

		resp, err := suite.Get("/websocket", nil)
		suite.Nil(err)
		suite.Equal("application/json; charset=utf-8", resp.Header.Get("Content-Type"))
		suite.Equal(http.StatusBadRequest, resp.StatusCode)

		json := map[string]string{}
		err = suite.GetJSONBody(resp, &json)
		resp.Body.Close()
		suite.Nil(err)
		if err != nil {
			suite.Equal(http.StatusText(http.StatusBadRequest), json["error"])
		}
	})
}

func (suite *WebsocketTestSuite) TestErrorHandler() {
	routeURL := ""
	executed := make(chan struct{}, 1)
	suite.RunServer(func(r *goyave.Router) {
		upgrader := Upgrader{
			ErrorHandler: func(request *goyave.Request, err error) {
				executed <- struct{}{}
			},
		}
		route := r.Get("/websocket", upgrader.Handler(func(c *Conn, request *goyave.Request) error {
			return fmt.Errorf("test error")
		}))
		routeURL = "ws" + strings.TrimPrefix(route.BuildURL(), config.GetString("server.protocol"))
	}, func() {
		suite.checkGracefulCloseResponse(routeURL, "Internal server error")
	})

	select {
	case <-time.After(suite.Timeout()):
		suite.Fail("Timeout waiting for upgrader error handler")
	case <-executed:
	}
}

func (suite *WebsocketTestSuite) TestGracefulClose() {
	routeURL := ""
	suite.RunServer(func(r *goyave.Router) {
		upgrader := Upgrader{}
		route := r.Get("/websocket", upgrader.Handler(func(c *Conn, request *goyave.Request) error {
			return nil // Immediately close connection
		}))
		routeURL = "ws" + strings.TrimPrefix(route.BuildURL(), config.GetString("server.protocol"))
	}, func() {
		conn, _, err := ws.DefaultDialer.Dial(routeURL, nil)
		if err != nil {
			suite.Error(err)
			return
		}
		defer conn.Close()

		messageType, _, err := conn.ReadMessage()
		suite.NotNil(err)

		closeErr, ok := err.(*ws.CloseError)
		suite.True(ok)
		if ok {
			suite.Equal(ws.CloseNormalClosure, closeErr.Code)
			suite.Equal(NormalClosureMessage, closeErr.Text)
		}

		// advanceFrame returns noFrame (-1) when a close frame is received
		suite.Equal(-1, messageType)
	})
}

func (suite *WebsocketTestSuite) TestGracefulCloseOnError() {
	routeURL := ""
	suite.RunServer(func(r *goyave.Router) {
		upgrader := Upgrader{}
		route := r.Get("/websocket", upgrader.Handler(func(c *Conn, request *goyave.Request) error {
			return fmt.Errorf("test error") // Immediately close connection with an error
		}))
		routeURL = "ws" + strings.TrimPrefix(route.BuildURL(), config.GetString("server.protocol"))
	}, func() {
		suite.checkGracefulCloseResponse(routeURL, "Internal server error")
	})
}

func (suite *WebsocketTestSuite) TestGracefulCloseOnErrorDebug() {
	routeURL := ""
	suite.RunServer(func(r *goyave.Router) {
		upgrader := Upgrader{}
		route := r.Get("/websocket", upgrader.Handler(func(c *Conn, request *goyave.Request) error {
			return fmt.Errorf("test error") // Immediately close connection with an error
		}))
		routeURL = "ws" + strings.TrimPrefix(route.BuildURL(), config.GetString("server.protocol"))
	}, func() {
		previousDebug := config.Get("app.debug")
		config.Set("app.debug", true)
		defer config.Set("app.debug", previousDebug)
		suite.checkGracefulCloseResponse(routeURL, "test error")
	})
}

func (suite *WebsocketTestSuite) checkGracefulCloseResponse(routeURL, expectedMessage string) {
	conn, _, err := ws.DefaultDialer.Dial(routeURL, nil)
	if err != nil {
		suite.Error(err)
		return
	}
	defer conn.Close()

	messageType, _, err := conn.ReadMessage()
	suite.NotNil(err)

	closeErr, ok := err.(*ws.CloseError)
	suite.True(ok)
	if ok {
		suite.Equal(ws.CloseInternalServerErr, closeErr.Code)
		suite.Equal(expectedMessage, closeErr.Text)
	}

	// advanceFrame returns noFrame (-1) when a close frame is received
	suite.Equal(-1, messageType)
}

func (suite *WebsocketTestSuite) TestUpgradeHeaders() {
	routeURL := ""
	suite.RunServer(func(r *goyave.Router) {
		upgrader := Upgrader{
			Headers: func(request *goyave.Request) http.Header {
				headers := http.Header{}
				headers.Set("X-Test-Header", "value")
				return headers
			},
		}
		route := r.Get("/websocket", upgrader.Handler(func(c *Conn, request *goyave.Request) error {
			return nil
		}))
		routeURL = "ws" + strings.TrimPrefix(route.BuildURL(), config.GetString("server.protocol"))

	}, func() {
		conn, resp, err := ws.DefaultDialer.Dial(routeURL, nil)
		if err != nil {
			suite.Error(err)
			return
		}
		defer conn.Close()

		suite.Equal(http.StatusSwitchingProtocols, resp.StatusCode)
		suite.Equal("value", resp.Header.Get("X-Test-Header"))

		_, _, err = conn.ReadMessage()
		suite.NotNil(err)
		_, ok := err.(*ws.CloseError)
		suite.True(ok)
	})
}

func (suite *WebsocketTestSuite) TestCloseHandshakeTimeout() {
	routeURL := ""
	suite.RunServer(func(r *goyave.Router) {
		upgrader := Upgrader{}
		route := r.Get("/websocket", upgrader.Handler(func(c *Conn, request *goyave.Request) error {
			return nil
		}))
		routeURL = "ws" + strings.TrimPrefix(route.BuildURL(), config.GetString("server.protocol"))
	}, func() {
		conn, _, err := ws.DefaultDialer.Dial(routeURL, nil)
		if err != nil {
			suite.Error(err)
			return
		}
		defer conn.Close()

		time.Sleep(1500 * time.Millisecond)

		mt, _, err := conn.ReadMessage()
		suite.NotNil(err)
		suite.Equal(-1, mt)
	})
}

func (suite *WebsocketTestSuite) TestCloseHandler() {
	c := newConn(&ws.Conn{})
	c.receivedClose = true

	suite.Nil(c.closeHandler(ws.CloseNormalClosure, ""))
	select {
	case <-c.waitClose:
		suite.Fail("Expected waitClose to be empty, but read something")
	default:
	}

	c = newConn(&ws.Conn{})

	suite.Nil(c.closeHandler(ws.CloseNormalClosure, ""))
	select {
	case <-c.waitClose:
	default:
		suite.Fail("Expected waitClose to not be empty")
	}
}

func (suite *WebsocketTestSuite) TestRecovery() {
	routeURL := ""
	suite.RunServer(func(r *goyave.Router) {
		upgrader := Upgrader{}
		route := r.Get("/websocket", upgrader.Handler(func(c *Conn, request *goyave.Request) error {
			panic(fmt.Errorf("test error"))
		}))
		routeURL = "ws" + strings.TrimPrefix(route.BuildURL(), config.GetString("server.protocol"))
	}, func() {
		previousDebug := config.Get("app.debug")
		config.Set("app.debug", true)
		defer config.Set("app.debug", previousDebug)
		suite.checkGracefulCloseResponse(routeURL, "test error")
	})
}

func (suite *WebsocketTestSuite) TestRecoveryNil() {
	routeURL := ""
	suite.RunServer(func(r *goyave.Router) {
		upgrader := Upgrader{}
		route := r.Get("/websocket", upgrader.Handler(func(c *Conn, request *goyave.Request) error {
			panic(nil)
		}))
		routeURL = "ws" + strings.TrimPrefix(route.BuildURL(), config.GetString("server.protocol"))
	}, func() {
		previousDebug := config.Get("app.debug")
		config.Set("app.debug", true)
		defer config.Set("app.debug", previousDebug)
		suite.checkGracefulCloseResponse(routeURL, "<nil>")
	})
}

func (suite *WebsocketTestSuite) TestRecoveryString() {
	routeURL := ""
	suite.RunServer(func(r *goyave.Router) {
		upgrader := Upgrader{}
		route := r.Get("/websocket", upgrader.Handler(func(c *Conn, request *goyave.Request) error {
			panic("panic reason")
		}))
		routeURL = "ws" + strings.TrimPrefix(route.BuildURL(), config.GetString("server.protocol"))
	}, func() {
		previousDebug := config.Get("app.debug")
		config.Set("app.debug", true)
		defer config.Set("app.debug", previousDebug)
		suite.checkGracefulCloseResponse(routeURL, "panic reason")
	})
}

func (suite *WebsocketTestSuite) TestCloseTruncateErrorMessage() {
	routeURL := ""
	suite.RunServer(func(r *goyave.Router) {
		upgrader := Upgrader{}
		route := r.Get("/websocket", upgrader.Handler(func(c *Conn, request *goyave.Request) error {
			return fmt.Errorf("This error has a rather very long error message that is longer than one hundrer and twenty five characters and is therefore an invalid control frame")
		}))
		routeURL = "ws" + strings.TrimPrefix(route.BuildURL(), config.GetString("server.protocol"))
	}, func() {
		previousDebug := config.Get("app.debug")
		config.Set("app.debug", true)
		defer config.Set("app.debug", previousDebug)
		suite.checkGracefulCloseResponse(routeURL, "This error has a rather very long error message that is longer than one hundrer and twenty five characters and is therefore")
	})
}

func (suite *WebsocketTestSuite) TestTruncateMessage() {
	message := "This error has a rather very long error message that is longer than one hundrer and twenty five characters and is therefore an invalid control frame"
	suite.Equal("This error has a rather very long error message that is longer than one hundrer and twenty five characters and is therefore", truncateMessage(message, maxCloseMessageLength))
}

func TestWebsocketSuite(t *testing.T) {
	goyave.RunTest(t, new(WebsocketTestSuite))
}
