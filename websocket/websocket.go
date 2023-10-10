package websocket

import (
	"net/http"
	"time"

	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/util/errors"

	ws "github.com/gorilla/websocket"
)

// TODO document websocket
// TODO test websocket

const (
	// NormalClosureMessage the message sent with the close frame
	// during the close handshake.
	NormalClosureMessage = "Server closed connection"
)

type Controller interface {
	goyave.Composable
	Serve(*Conn, *goyave.Request) error
}

type Regsitrer interface {
	RegisterRoute(*goyave.Router, goyave.Handler)
}

type upgradeErrorHandlerFunc func(response *goyave.Response, request *goyave.Request, status int, reason error)

type UpgradeErrorHandler interface {
	OnUpgradeError(response *goyave.Response, request *goyave.Request, status int, reason error)
}

type ErrorHandler interface {
	OnError(request *goyave.Request, err error)
}

type OriginChecker interface {
	CheckOrigin(r *goyave.Request) bool
}

type HeaderUpgrader interface {
	UpgradeHeaders(request *goyave.Request) http.Header
}

// Upgrader is responsible for the upgrade of HTTP connections to
// websocket connections.
type Upgrader struct {
	goyave.Component

	Controller Controller

	// Settings the parameters for upgrading the connection. "Error" and "CheckOrigin" are
	// ignored: use implementations of the interfaces `UpgradeErrorHandler` and `ErrorHandler`.
	Settings ws.Upgrader
}

// New create a new Upgrader with default settings.
func New(controller Controller) *Upgrader {
	return &Upgrader{
		Controller: controller,
	}
}

func (u *Upgrader) RegisterRoutes(router *goyave.Router) {
	if registrer, ok := u.Controller.(Regsitrer); ok {
		registrer.RegisterRoute(router, u.Handler())
		return
	}
	router.Get("", u.Handler())
}

func (u *Upgrader) defaultUpgradeErrorHandler(response *goyave.Response, _ *goyave.Request, status int, reason error) {
	text := http.StatusText(status)
	if u.Config().GetBool("app.debug") && reason != nil {
		text = reason.Error()
	}
	message := map[string]string{
		"error": text,
	}
	response.JSON(status, message)
}

func (u *Upgrader) makeUpgrader(request *goyave.Request) *ws.Upgrader {
	upgradeErrorHandlerFunc := u.defaultUpgradeErrorHandler
	if upgradeErrorHandler, ok := u.Controller.(UpgradeErrorHandler); ok {
		upgradeErrorHandlerFunc = upgradeErrorHandler.OnUpgradeError
	}

	var checkOrigin func(r *goyave.Request) bool
	if originChecker, ok := u.Controller.(OriginChecker); ok {
		checkOrigin = originChecker.CheckOrigin
	}

	a := adapter{
		upgradeErrorHandler: upgradeErrorHandlerFunc,
		checkOrigin:         checkOrigin,
		request:             request,
	}

	upgrader := u.Settings
	upgrader.Error = a.onError
	upgrader.CheckOrigin = a.getCheckOriginFunc()
	return &upgrader
}

// Handler create an HTTP handler upgrading the HTTP connection before passing it
// to the given websocket Handler.
//
// HTTP response's status is set to "101 Switching Protocols".
//
// The connection is closed automatically after the websocket Handler returns, using the
// closing handshake defined by RFC 6455 Section 1.4 if possible and if not already
// performed using "conn.Close()".
//
// If the websocket Handler returns an error that is not a CloseError, the Upgrader's error
// handler will be executed and the close frame sent to the client will have status code
// 1011 (internal server error) and "Internal server error" as message.
// If debug is enabled, the message will be set to the one of the error returned by the
// websocket Handler. Otherwise the close frame will have status code 1000 (normal closure)
// and "Server closed connection" as a message.
//
// This HTTP handler features a recovery mechanism. If the websocket Handler panics,
// the connection will be gracefully closed just like if the websocket Handler returned
// an error without panicking.
//
// This HTTP Handler returns once the connection has been successfully upgraded. That means
// that, for example, logging middleware will log the request right away instead of waiting
// for the websocket connection to be closed.
func (u *Upgrader) Handler() goyave.Handler {
	u.Controller.Init(u.Server())
	return func(response *goyave.Response, request *goyave.Request) {
		var headers http.Header
		if headerUpgrader, ok := u.Controller.(HeaderUpgrader); ok {
			headers = headerUpgrader.UpgradeHeaders(request)
		}

		c, err := u.makeUpgrader(request).Upgrade(response, request.Request(), headers)
		if err != nil {
			return
		}
		response.Status(http.StatusSwitchingProtocols)

		go u.serve(c, request, u.Controller.Serve)
	}
}

func (u *Upgrader) serve(c *ws.Conn, request *goyave.Request, handler func(*Conn, *goyave.Request) error) {
	conn := newConn(c, time.Duration(u.Config().GetInt("server.websocketTimeout"))*time.Second)
	panicked := true
	var err error
	defer func() { // Panic recovery
		if panicReason := recover(); panicReason != nil || panicked {
			err = errors.NewSkip(panicReason, 4) // Skipped: runtime.Callers, NewSkip, this func, runtime.panic
		}

		if IsCloseError(err) {
			_ = conn.CloseNormal()
			return
		}
		if err != nil {
			if errorHandler, ok := u.Controller.(ErrorHandler); ok {
				errorHandler.OnError(request, err)
			} else {
				u.Logger().Error(err)
			}
			_ = conn.CloseWithError(err)
		} else {
			_ = conn.internalClose(ws.CloseNormalClosure, NormalClosureMessage)
		}
	}()

	err = handler(conn, request)
	if err != nil {
		err = errors.New(err)
	}
	panicked = false
}

type adapter struct {
	upgradeErrorHandler upgradeErrorHandlerFunc
	checkOrigin         func(r *goyave.Request) bool
	request             *goyave.Request
}

func (a *adapter) onError(w http.ResponseWriter, _ *http.Request, status int, reason error) {
	if status == http.StatusInternalServerError {
		panic(errors.New(reason))
	}
	w.Header().Set("Sec-Websocket-Version", "13")
	a.upgradeErrorHandler(w.(*goyave.Response), a.request, status, reason)
}

func (a *adapter) getCheckOriginFunc() func(r *http.Request) bool {
	if a.checkOrigin != nil {
		return func(r *http.Request) bool {
			return a.checkOrigin(a.request)
		}
	}

	return nil
}

// IsCloseError returns true if the error is one of the following close errors:
// CloseNormalClosure (1000), CloseGoingAway (1001) or CloseNoStatusReceived (1005)
func IsCloseError(err error) bool {
	return ws.IsCloseError(err,
		ws.CloseNormalClosure,
		ws.CloseGoingAway,
		ws.CloseNoStatusReceived,
	)
}
