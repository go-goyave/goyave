package websocket

import (
	"net/http"
	"time"

	stderrors "errors"

	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/util/errors"

	ws "github.com/gorilla/websocket"
)

const (
	// NormalClosureMessage the message sent with the close frame
	// during the close handshake.
	NormalClosureMessage = "Server closed connection"
)

// Controller component for websockets.
type Controller interface {
	goyave.Composable

	// Serve is a handler for websocket connections.
	// The request parameter contains the original upgraded HTTP request.
	//
	// To keep the connection alive, these handlers should run an infinite for loop that
	// can return on error or exit in a predictable manner.
	//
	// They also can start goroutines for reads and writes, but shouldn't return before
	// both of them do. The handler is responsible of synchronizing the goroutines it started,
	// and ensure no reader nor writer are still active when it returns.
	//
	// When the websocket handler returns, the closing handshake is performed (if not already done
	// using "conn.Close()") and the connection is closed.
	//
	// If the websocket handler returns nil, it means that everything went fine and the
	// connection can be closed normally. On the other hand, the websocket handler
	// can return an error, such as a write error, to indicate that the connection should not
	// be closed normally. The behavior used when this happens depend on the implementation
	// of the HTTP handler that upgraded the connection.
	//
	// By default, the server shutdown doesn't wait for hijacked connections to be closed gracefully.
	// It is advised to register a shutdown hook blocking until all the connections are gracefully
	// closed using `*websocket.Conn.CloseNormal()`.
	//
	// The following websocket Handler is a simple example of an "echo" feature using websockets:
	//
	//	func (c *EchoController) Serve(c *websocket.Conn, request *goyave.Request) error {
	//		for {
	//			mt, message, err := c.ReadMessage()
	//			if err != nil {
	//				return errors.New(err)
	//			}
	//			c.Logger().Debug("recv", "message", string(message))
	//			err = c.WriteMessage(mt, message)
	//			if err != nil {
	//				return errors.Errof("write: %w", err)
	//			}
	//		}
	//	}
	Serve(conn *Conn, request *goyave.Request) error
}

// Registrer qualifies a `websocket.Controller` that registers its route itself, allowing
// to define validation rules, middleware, route meta, etc.
//
// If the `websocket.Controller` doesn't implement this interface, the route is registered
// for the GET method and an empty path.
type Registrer interface {
	// RegisterRoute registers the route for the websocket upgrade. The route must only match the
	// GET HTTP method and use the `goyave.Handler` received as a parameter.
	RegisterRoute(router *goyave.Router, handler goyave.Handler)
}

type upgradeErrorHandlerFunc func(response *goyave.Response, request *goyave.Request, status int, reason error)

// UpgradeErrorHandler allows a `websocket.Controller` to define a custom behavior when
// the protocol switching process fails.
//
// If the `websocket.Controller` doesn't implement this interface, the default
// error handler returns a JSON response containing the status text
// corresponding to the status code returned. If debugging is enabled, the reason error
// message is returned instead.
//
//	{"error": "message"}
type UpgradeErrorHandler interface {
	// OnUpgradeError specifies the function for generating HTTP error responses if the
	// protocol switching process fails. The error can be a user error or server error.
	OnUpgradeError(response *goyave.Response, request *goyave.Request, status int, reason error)
}

// ErrorHandler allows a `websocket.Controller` to define a custom behavior in case of error
// occurring in the controller's `Serve` function. This custom error handler is called for both
// handled and unhandled errors (panics).
//
// If the `websocket.Controller` doesn't implement this interface, the error is logged at the error level.
type ErrorHandler interface {
	// ErrorHandler specifies the function handling errors returned by the controller's `Serve` function
	// or if this function panics.
	OnError(request *goyave.Request, err error)
}

// OriginChecker allows a `websocket.Controller` to define custom origin header checking behavior.
//
// If the `websocket.Controller` doesn't implement this interface, a safe default is used:
// return false if the Origin request header is present and the origin host is not equal to
// request Host header.
type OriginChecker interface {
	// CheckOrigin returns true if the request Origin header is acceptable.
	//
	// A CheckOrigin function should carefully validate the request origin to
	// prevent cross-site request forgery.
	CheckOrigin(r *goyave.Request) bool
}

// HeaderUpgrader allows a `websocket.Controller` to define custom HTTP headers in the
// protocol switching response.
type HeaderUpgrader interface {
	// UpgradeHeaders function generating headers to be sent with the protocol switching response.
	UpgradeHeaders(r *goyave.Request) http.Header
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

// RegisterRoutes implementation of `goyave.Registrer`.
//
// If the `websocket.Controller` implements `websocket.Registrer`, uses its implementation
// to register the route. Otherwise registers the route for the GET method and an empty path.
func (u *Upgrader) RegisterRoutes(router *goyave.Router) {
	if registrer, ok := u.Controller.(Registrer); ok {
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
// If debug is enabled, the message will be the error message returned by the
// websocket handler. Otherwise the close frame will have status code 1000 (normal closure)
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
	conn := newConn(c, time.Duration(u.Config().GetInt("server.websocketCloseTimeout"))*time.Second)
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
		return func(_ *http.Request) bool {
			return a.checkOrigin(a.request)
		}
	}

	return nil
}

// IsCloseError returns true if the error is one of the following close errors:
// CloseNormalClosure (1000), CloseGoingAway (1001) or CloseNoStatusReceived (1005)
func IsCloseError(err error) bool {
	var closeError *ws.CloseError
	if stderrors.As(err, &closeError) {
		err = closeError
	}
	return ws.IsCloseError(err,
		ws.CloseNormalClosure,
		ws.CloseGoingAway,
		ws.CloseNoStatusReceived,
	)
}
