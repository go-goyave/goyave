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
	"github.com/gorilla/websocket"
	ws "github.com/gorilla/websocket"
)

type WebsocketTestSuite struct {
	goyave.TestSuite
}

func echoWSHandler(c *Conn, request *goyave.Request) error {
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			if IsCloseError(err) {
				return nil
			}
			return fmt.Errorf("read: %v", err)
		}
		goyave.Logger.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			return fmt.Errorf("write: %v", err)
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
		route := r.Get("/websocket", upgrader.Handler(echoWSHandler))
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
		suite.Nil(conn.WriteMessage(websocket.TextMessage, message))

		messageType, data, err := conn.ReadMessage()
		suite.Nil(err)
		suite.Equal(websocket.TextMessage, messageType)
		suite.Equal(message, data)
	})
}

func (suite *WebsocketTestSuite) TestUpgradeError() {
	suite.RunServer(func(r *goyave.Router) {
		upgrader := Upgrader{}
		r.Get("/websocket", upgrader.Handler(echoWSHandler))
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
	executed := make(chan struct{})
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
		conn, _, err := ws.DefaultDialer.Dial(routeURL, nil)
		if err != nil {
			suite.Error(err)
			return
		}
		defer conn.Close()
	})

	select {
	case <-time.After(suite.Timeout()):
		suite.Fail("Timeout waiting for upgrader error handler")
	case <-executed:
	}
}

// func (suite *WebsocketTestSuite) TestGracefulClose() {
// 	routeURL := ""
// 	suite.RunServer(func(r *goyave.Router) {
// 		upgrader := Upgrader{}
// 		route := r.Get("/websocket", upgrader.Handler(func(c *Conn, request *goyave.Request) error {
// 			return fmt.Errorf("test error")
// 		}))
// 		routeURL = "ws" + strings.TrimPrefix(route.BuildURL(), config.GetString("server.protocol"))
// 	}, func() {
// 		conn, _, err := ws.DefaultDialer.Dial(routeURL, nil)
// 		if err != nil {
// 			suite.Error(err)
// 			return
// 		}
// 		defer conn.Close()

// 		messageType, _, err := conn.ReadMessage()
// 		suite.Nil(err)
// 		suite.Equal(websocket.CloseMessage, messageType)
// 	})
// }

func TestWebsocketSuite(t *testing.T) {
	goyave.RunTest(t, new(WebsocketTestSuite))
}
