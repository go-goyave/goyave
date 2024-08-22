package websocket

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/testutil"

	ws "github.com/gorilla/websocket"

	stdslog "log/slog"
)

func prepareTestConfig() goyave.Options {
	cfg := config.LoadDefault()
	cfg.Set("server.port", 0)
	cfg.Set("server.websocketCloseTimeout", 1)
	cfg.Set("app.debug", false)
	return goyave.Options{Config: cfg}
}

type testController struct {
	goyave.Component
	t  *testing.T
	wg *sync.WaitGroup

	serve          func(conn *Conn, r *goyave.Request) error
	checkOrigin    func(r *goyave.Request) bool
	upgradeHeaders func(r *goyave.Request) http.Header
}

func (c *testController) CheckOrigin(r *goyave.Request) bool {
	if c.checkOrigin != nil {
		return c.checkOrigin(r)
	}
	return false
}

func (c *testController) UpgradeHeaders(r *goyave.Request) http.Header {
	if c.upgradeHeaders != nil {
		return c.upgradeHeaders(r)
	}
	return http.Header{}
}

func (c *testController) Serve(conn *Conn, r *goyave.Request) error {
	c.wg.Add(1)
	defer c.wg.Done()
	if c.serve != nil {
		return c.serve(conn, r)
	}
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			if IsCloseError(err) {
				return err
			}
			err = fmt.Errorf("read: %w", err)
			assert.Error(c.t, err)
			return err
		}
		err = conn.WriteMessage(mt, message)
		if err != nil {
			err = fmt.Errorf("write: %w", err)
			assert.Error(c.t, err)
			return err
		}
	}
}

type testControllerWithErrorHandler struct {
	testController

	onUpgradeError func(response *goyave.Response, request *goyave.Request, status int, reason error)
	onError        func(c *testControllerWithErrorHandler, request *goyave.Request, err error)
}

func (c *testControllerWithErrorHandler) OnUpgradeError(response *goyave.Response, request *goyave.Request, status int, reason error) {
	if c.onUpgradeError != nil {
		c.onUpgradeError(response, request, status, reason)
		return
	}
}

func (c *testControllerWithErrorHandler) OnError(request *goyave.Request, err error) {
	if c.onError != nil {
		c.onError(c, request, err)
		return
	}
}

type testControllerRegistrer struct {
	testController

	registerRoute func(*goyave.Router, goyave.Handler)
}

func (c *testControllerRegistrer) RegisterRoute(router *goyave.Router, handler goyave.Handler) {
	if c.registerRoute != nil {
		c.registerRoute(router, handler)
	}
}

func TestIsCloseError(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{err: &ws.CloseError{Code: ws.CloseNormalClosure}, want: true},
		{err: &ws.CloseError{Code: ws.CloseGoingAway}, want: true},
		{err: &ws.CloseError{Code: ws.CloseNoStatusReceived}, want: true},
		{err: fmt.Errorf("wrap: %w", &ws.CloseError{Code: ws.CloseNoStatusReceived}), want: true},
		{err: &ws.CloseError{Code: ws.CloseAbnormalClosure}, want: false},
		{err: &ws.CloseError{Code: ws.CloseProtocolError}, want: false},
		{err: fmt.Errorf("wrap: %w", &ws.CloseError{Code: ws.CloseProtocolError}), want: false},
		{err: errors.New(&ws.CloseError{Code: ws.CloseNormalClosure}), want: true},
		{err: errors.New(&ws.CloseError{Code: ws.CloseProtocolError}), want: false},
	}

	for _, c := range cases {
		t.Run(c.err.Error(), func(t *testing.T) {
			assert.Equal(t, c.want, IsCloseError(c.err))
		})
	}
}

func TestAdapterOnError(t *testing.T) {
	req := testutil.NewTestRequest(http.MethodGet, "/websocket", nil)
	resp, _ := testutil.NewTestResponse(req)
	reasonErr := fmt.Errorf("test adapter error")
	executed := false
	a := adapter{
		upgradeErrorHandler: func(response *goyave.Response, request *goyave.Request, status int, reason error) {
			assert.Equal(t, req, request)
			assert.Equal(t, resp, response)
			assert.Equal(t, http.StatusBadRequest, status)
			assert.Equal(t, reasonErr, reason)
			executed = true
		},
		request: req,
	}

	a.onError(resp, req.Request(), http.StatusBadRequest, reasonErr)
	assert.True(t, executed)
	assert.Equal(t, "13", resp.Header().Get("Sec-Websocket-Version"))

	assert.Panics(t, func() {
		a.onError(resp, req.Request(), http.StatusInternalServerError, reasonErr)
	})
}

func TestGetCheckOriginFunction(t *testing.T) {
	req := testutil.NewTestRequest(http.MethodGet, "/websocket", nil)
	a := adapter{
		request: req,
	}
	assert.Nil(t, a.getCheckOriginFunc())

	executed := false
	a.checkOrigin = func(r *goyave.Request) bool {
		assert.Equal(t, req, r)
		executed = true
		return true
	}

	f := a.getCheckOriginFunc()
	assert.NotNil(t, f)
	assert.True(t, f(req.Request()))
	assert.True(t, executed)
}

func TestDefaultUpgradeErrorHandler(t *testing.T) {
	cases := []struct {
		config    func() goyave.Options
		expect    func(*testing.T, map[string]string)
		reasonErr error
		desc      string
	}{
		{
			desc:      "debug_on",
			config:    func() goyave.Options { return goyave.Options{Config: config.LoadDefault()} },
			reasonErr: fmt.Errorf("test upgrade error handler"),
			expect: func(t *testing.T, body map[string]string) {
				assert.Equal(t, map[string]string{"error": "test upgrade error handler"}, body)
			},
		},
		{
			desc:      "debug_off",
			config:    prepareTestConfig,
			reasonErr: fmt.Errorf("test upgrade error handler"),
			expect: func(t *testing.T, body map[string]string) {
				assert.Equal(t, map[string]string{"error": http.StatusText(http.StatusBadRequest)}, body)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			server := testutil.NewTestServerWithOptions(t, c.config())
			req := server.NewTestRequest(http.MethodGet, "/websocket", nil)
			resp, recorder := server.NewTestResponse(req)

			upgrader := &Upgrader{}
			upgrader.Init(server.Server)
			upgrader.defaultUpgradeErrorHandler(resp, req, http.StatusBadRequest, c.reasonErr)

			result := recorder.Result()
			assert.Equal(t, "application/json; charset=utf-8", result.Header.Get("Content-Type"))
			assert.Equal(t, http.StatusBadRequest, result.StatusCode)
			body, err := testutil.ReadJSONBody[map[string]string](result.Body)
			assert.NoError(t, result.Body.Close())
			assert.NoError(t, err)
			c.expect(t, body)
		})
	}
}

func TestMakeUpgrader(t *testing.T) {
	upgrader := Upgrader{}

	req := testutil.NewTestRequest(http.MethodGet, "/websocket", nil)
	u := upgrader.makeUpgrader(req)

	assert.Equal(t, upgrader.Settings.HandshakeTimeout, u.HandshakeTimeout)
	assert.Equal(t, upgrader.Settings.ReadBufferSize, u.ReadBufferSize)
	assert.Equal(t, upgrader.Settings.WriteBufferSize, u.WriteBufferSize)
	assert.Equal(t, upgrader.Settings.WriteBufferPool, u.WriteBufferPool)
	assert.Equal(t, upgrader.Settings.Subprotocols, u.Subprotocols)
	assert.Equal(t, upgrader.Settings.EnableCompression, u.EnableCompression)
	assert.NotNil(t, u.Error)
	assert.Nil(t, u.CheckOrigin)

	upgrader.Settings.EnableCompression = true
	u = upgrader.makeUpgrader(req)
	assert.Equal(t, upgrader.Settings.EnableCompression, u.EnableCompression)

	upgradeErrorExecuted := false
	checkOriginExecuted := false
	upgrader.Controller = &testControllerWithErrorHandler{
		onUpgradeError: func(_ *goyave.Response, _ *goyave.Request, _ int, _ error) {
			upgradeErrorExecuted = true
		},
		testController: testController{
			checkOrigin: func(_ *goyave.Request) bool {
				checkOriginExecuted = true
				return true
			},
		},
	}

	u = upgrader.makeUpgrader(req)
	assert.True(t, u.CheckOrigin(req.Request()))
	assert.True(t, checkOriginExecuted)

	resp, _ := testutil.NewTestResponse(req)
	u.Error(resp, nil, 0, nil)
	assert.True(t, upgradeErrorExecuted)
}

func TestUpgrade(t *testing.T) {
	// Server shutdown doesn't wait for Hijacked connections to
	// terminate before returning.
	wg := sync.WaitGroup{}
	wg.Add(2)

	var routeURL string
	server := testutil.NewTestServerWithOptions(t, prepareTestConfig())
	server.RegisterRoutes(func(_ *goyave.Server, r *goyave.Router) {
		upgrader := New(&testController{
			t:  t,
			wg: &wg,
			checkOrigin: func(_ *goyave.Request) bool {
				return true
			},
			upgradeHeaders: func(_ *goyave.Request) http.Header {
				headers := http.Header{}
				headers.Add("X-Test", "Value")
				return headers
			},
		})
		r.Subrouter("/websocket").Controller(upgrader)
	})

	server.RegisterStartupHook(func(s *goyave.Server) {
		defer func() {
			server.Stop()
			wg.Done()
		}()
		route := s.Router().GetSubrouters()[0].GetRoutes()[0]
		routeURL = "ws" + strings.TrimPrefix(route.BuildURL(), "http")

		conn, resp, err := ws.DefaultDialer.Dial(routeURL, nil)
		assert.Equal(t, "Value", resp.Header.Get("X-Test"))
		assert.NoError(t, err, fmt.Sprintf("RESPONSE STATUS: %d, RESPONSE HEADERS: %v", resp.StatusCode, resp.Header))
		assert.NoError(t, resp.Body.Close())
		defer func() {
			assert.NoError(t, conn.Close())
		}()

		message := []byte("hello world")
		assert.NoError(t, conn.WriteMessage(ws.TextMessage, message))

		messageType, data, err := conn.ReadMessage()
		assert.NoError(t, err)
		assert.Equal(t, ws.TextMessage, messageType)
		assert.Equal(t, message, data)

		m := ws.FormatCloseMessage(ws.CloseNormalClosure, "Connection closed by client")
		assert.NoError(t, conn.WriteControl(ws.CloseMessage, m, time.Now().Add(time.Second)))
	})

	go func() {
		assert.NoError(t, server.Start())
		wg.Done()
	}()
	wg.Wait()
}

func TestUpgradeError(t *testing.T) {
	server := testutil.NewTestServerWithOptions(t, prepareTestConfig())
	server.RegisterRoutes(func(_ *goyave.Server, r *goyave.Router) {
		upgrader := New(&testController{
			t: t,
			checkOrigin: func(_ *goyave.Request) bool {
				return true
			},
		})
		r.Subrouter("/websocket").Controller(upgrader)
	})

	resp := server.TestRequest(httptest.NewRequest(http.MethodGet, "/websocket", nil))

	assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
	assert.NoError(t, resp.Body.Close())
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"error": http.StatusText(http.StatusBadRequest)}, body)
}

func TestRegistrer(t *testing.T) {
	server := testutil.NewTestServerWithOptions(t, prepareTestConfig())
	upgrader := New(&testControllerRegistrer{
		registerRoute: func(router *goyave.Router, handler goyave.Handler) {
			router.Get("", handler).SetMeta("key", "value").Name("websocket")
		},
	})
	router := server.Router()
	router.Subrouter("/websocket").Controller(upgrader)

	route := router.GetRoute("websocket")
	if !assert.NotNil(t, route) {
		return
	}

	assert.Equal(t, "value", route.Meta["key"])
}

func TestConnCloseHandshakeTimeout(t *testing.T) {
	c := newConn(&ws.Conn{}, 0)

	c.SetCloseHandshakeTimeout(time.Second * 2)
	assert.Equal(t, time.Second*2, c.closeTimeout)
	assert.Equal(t, time.Second*2, c.GetCloseHandshakeTimeout())
}

func TestCloseHandshakeTimeout(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	var routeURL string
	server := testutil.NewTestServerWithOptions(t, prepareTestConfig())
	server.RegisterRoutes(func(_ *goyave.Server, r *goyave.Router) {
		upgrader := New(&testController{
			t:  t,
			wg: &wg,
			serve: func(_ *Conn, _ *goyave.Request) error {
				return nil // Immediately return to trigger the close handshake
			},
			checkOrigin: func(_ *goyave.Request) bool {
				return true
			},
		})
		r.Subrouter("/websocket").Controller(upgrader)
	})

	server.RegisterStartupHook(func(s *goyave.Server) {
		defer func() {
			server.Stop()
			wg.Done()
		}()
		route := s.Router().GetSubrouters()[0].GetRoutes()[0]
		routeURL = "ws" + strings.TrimPrefix(route.BuildURL(), "http")

		conn, resp, err := ws.DefaultDialer.Dial(routeURL, nil)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err, fmt.Sprintf("RESPONSE STATUS: %d, RESPONSE HEADERS: %v", resp.StatusCode, resp.Header))
		defer func() {
			assert.NoError(t, conn.Close())
		}()
		time.Sleep(1500 * time.Millisecond)

		messageType, _, err := conn.ReadMessage()
		assert.Error(t, err)

		// The server has sent the close handshake payload with NormalClosureMessage
		assert.Equal(t, &ws.CloseError{Code: ws.CloseNormalClosure, Text: NormalClosureMessage}, err)
		assert.Equal(t, -1, messageType)
	})

	go func() {
		assert.NoError(t, server.Start())
		wg.Done()
	}()
	wg.Wait()
}

func TestCloseHandler(t *testing.T) {
	c := newConn(&ws.Conn{}, 1*time.Second)

	assert.NoError(t, c.closeHandler(ws.CloseNormalClosure, ""))
	select {
	case <-c.waitClose:
	default:
		assert.Fail(t, "Expected waitClose to not be empty")
	}
}

func TestGracefulClose(t *testing.T) {
	cases := []struct {
		expectedError *ws.CloseError
		serve         func(conn *Conn, r *goyave.Request) error
		errorHandler  func(c *testControllerWithErrorHandler, request *goyave.Request, err error)
		expectedLogs  *regexp.Regexp
		desc          string
	}{
		{
			desc: "recovery",
			serve: func(_ *Conn, _ *goyave.Request) error {
				panic("websocket handler panic")
			},
			expectedError: &ws.CloseError{Code: ws.CloseInternalServerErr, Text: http.StatusText(http.StatusInternalServerError)},
			expectedLogs:  regexp.MustCompile(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","msg":"websocket handler panic","trace":".+"}\n`),
		},
		{
			desc: "normal_error",
			serve: func(_ *Conn, _ *goyave.Request) error {
				return errors.New("websocket handler error")
			},
			expectedError: &ws.CloseError{Code: ws.CloseInternalServerErr, Text: http.StatusText(http.StatusInternalServerError)},
			expectedLogs:  regexp.MustCompile(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","msg":"websocket handler error","trace":".+"}\n`),
		},
		{
			desc: "erro_handler",
			serve: func(_ *Conn, _ *goyave.Request) error {
				return errors.New("websocket handler error")
			},
			errorHandler: func(c *testControllerWithErrorHandler, _ *goyave.Request, _ error) {
				c.Logger().Info("message override")
			},
			expectedError: &ws.CloseError{Code: ws.CloseInternalServerErr, Text: http.StatusText(http.StatusInternalServerError)},
			expectedLogs:  regexp.MustCompile(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"INFO","msg":"message override"}\n`),
		},
		{
			desc: "normal_serve_closure",
			serve: func(_ *Conn, _ *goyave.Request) error {
				return nil
			},
			expectedError: &ws.CloseError{Code: ws.CloseNormalClosure, Text: NormalClosureMessage},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			wg := sync.WaitGroup{}
			wg.Add(2)

			var routeURL string
			server := testutil.NewTestServerWithOptions(t, prepareTestConfig())
			server.RegisterRoutes(func(_ *goyave.Server, r *goyave.Router) {
				var ctrl Controller = &testController{
					t:     t,
					wg:    &wg,
					serve: c.serve,

					checkOrigin: func(_ *goyave.Request) bool {
						return true
					},
				}
				if c.errorHandler != nil {
					ctrl = &testControllerWithErrorHandler{
						testController: *ctrl.(*testController),
						onError:        c.errorHandler,
					}
				}
				upgrader := New(ctrl)
				r.Subrouter("/websocket").Controller(upgrader)
			})
			buf := &bytes.Buffer{}
			server.Logger = slog.New(stdslog.NewJSONHandler(buf, &stdslog.HandlerOptions{Level: stdslog.LevelInfo}))

			server.RegisterStartupHook(func(s *goyave.Server) {
				defer func() {
					server.Stop()
					wg.Done()
				}()
				route := s.Router().GetSubrouters()[0].GetRoutes()[0]
				routeURL = "ws" + strings.TrimPrefix(route.BuildURL(), "http")

				testGracefulClose(t, routeURL, c.expectedError)
			})

			go func() {
				assert.NoError(t, server.Start())
				wg.Done()
			}()
			wg.Wait()
			if c.expectedLogs != nil {
				assert.Regexp(t, c.expectedLogs, buf.String())
			}
		})
	}
}

func testGracefulClose(t *testing.T, routeURL string, expectedError *ws.CloseError) {
	conn, resp, err := ws.DefaultDialer.Dial(routeURL, nil)
	assert.NoError(t, resp.Body.Close())
	assert.NoError(t, err, fmt.Sprintf("RESPONSE STATUS: %d, RESPONSE HEADERS: %v", resp.StatusCode, resp.Header))
	defer func() {
		assert.NoError(t, conn.Close())
	}()

	messageType, _, err := conn.ReadMessage()
	assert.Equal(t, expectedError, err)

	// advanceFrame returns noFrame (-1) when a close frame is received
	assert.Equal(t, -1, messageType)
}

func TestCloseConnectionClosed(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(3)

	var routeURL string
	server := testutil.NewTestServerWithOptions(t, prepareTestConfig())
	server.RegisterRoutes(func(s *goyave.Server, r *goyave.Router) {
		upgrader := New(&testController{
			t: t,
			checkOrigin: func(_ *goyave.Request) bool {
				return true
			},
		})
		upgrader.Init(s)
		upgrader.Controller.Init(s)
		r.Subrouter("/websocket").Get("", func(response *goyave.Response, request *goyave.Request) {
			defer wg.Done()
			c, err := upgrader.makeUpgrader(request).Upgrade(response, request.Request(), nil)
			assert.NoError(t, err)
			response.Status(http.StatusSwitchingProtocols)

			conn := newConn(c, time.Second)

			assert.NoError(t, conn.Conn.Close()) // Connection closed right away, server wont be able to write anymore
			err = conn.CloseNormal()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "use of closed network connection")
		})
	})

	server.RegisterStartupHook(func(s *goyave.Server) {
		defer func() {
			server.Stop()
			wg.Done()
		}()
		route := s.Router().GetSubrouters()[0].GetRoutes()[0]
		routeURL = "ws" + strings.TrimPrefix(route.BuildURL(), "http")

		conn, resp, err := ws.DefaultDialer.Dial(routeURL, nil)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err, fmt.Sprintf("RESPONSE STATUS: %d, RESPONSE HEADERS: %v", resp.StatusCode, resp.Header))
		defer func() {
			assert.NoError(t, conn.Close())
		}()

		_, _, err = conn.ReadMessage()
		assert.Error(t, err)
		assert.Equal(t, &ws.CloseError{Code: ws.CloseAbnormalClosure, Text: "unexpected EOF"}, err)
	})

	go func() {
		assert.NoError(t, server.Start())
		wg.Done()
	}()
	wg.Wait()
}

func TestCloseWriteTimeout(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(3)

	var routeURL string
	server := testutil.NewTestServerWithOptions(t, prepareTestConfig())
	server.RegisterRoutes(func(s *goyave.Server, r *goyave.Router) {
		upgrader := New(&testController{
			t: t,
			checkOrigin: func(_ *goyave.Request) bool {
				return true
			},
		})
		upgrader.Init(s)
		upgrader.Controller.Init(s)
		r.Subrouter("/websocket").Get("", func(response *goyave.Response, request *goyave.Request) {
			defer wg.Done()
			c, err := upgrader.makeUpgrader(request).Upgrade(response, request.Request(), nil)
			assert.NoError(t, err)
			response.Status(http.StatusSwitchingProtocols)

			conn := newConn(c, time.Second)
			conn.closeTimeout = -1 * time.Second

			// No error expected, the connection should close as normal without waiting
			assert.NoError(t, conn.CloseNormal())
		})
	})

	server.RegisterStartupHook(func(s *goyave.Server) {
		defer func() {
			server.Stop()
			wg.Done()
		}()
		route := s.Router().GetSubrouters()[0].GetRoutes()[0]
		routeURL = "ws" + strings.TrimPrefix(route.BuildURL(), "http")

		conn, resp, err := ws.DefaultDialer.Dial(routeURL, nil)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err, fmt.Sprintf("RESPONSE STATUS: %d, RESPONSE HEADERS: %v", resp.StatusCode, resp.Header))
		assert.NoError(t, conn.Close())
	})

	go func() {
		assert.NoError(t, server.Start())
		wg.Done()
	}()
	wg.Wait()
}
