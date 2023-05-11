package websocket

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
	"unicode/utf8"

	"goyave.dev/goyave/v4"
	"goyave.dev/goyave/v4/config"

	ws "github.com/gorilla/websocket"
)

const (
	// NormalClosureMessage the message sent with the close frame
	// during the close handshake.
	NormalClosureMessage = "Server closed connection"

	maxCloseMessageLength = 123
)

var timeout time.Duration

func setTimeout() {
	if timeout == 0 {
		timeout = time.Duration(config.GetInt("server.timeout")) * time.Second
	}
}

// PanicError error sent to the upgrader's ErrorHandler if a panic occurred.
// If debugging is disabled, the "Stacktrace" field will be empty.
type PanicError struct {
	Reason     error
	Stacktrace string
}

func (e *PanicError) Error() string {
	return e.Reason.Error()
}

// Unwrap return the panic reason.
func (e *PanicError) Unwrap() error {
	return e.Reason
}

// Handler is a handler for websocket connections.
// The request parameter contains the original upgraded HTTP request.
//
// To keep the connection alive, these handlers should run an infinite for loop that
// can return on error or exit in a predictable manner.
//
// They also can start goroutines for reads and writes, but shouldn't return before
// both of them do. The Handler is responsible of synchronizing the goroutines it started,
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
// The following websocket Handler is an example of an "echo" feature using websockets:
//
//	func Echo(c *websocket.Conn, request *goyave.Request) error {
//		for {
//			mt, message, err := c.ReadMessage()
//			if err != nil {
//				return err
//			}
//			goyave.Logger.Printf("recv: %s", message)
//			err = c.WriteMessage(mt, message)
//			if err != nil {
//				return fmt.Errorf("write: %w", err)
//			}
//		}
//	}
type Handler func(*Conn, *goyave.Request) error

// UpgradeErrorHandler is a specific Handler type for connection upgrade errors.
// These handlers are called when an error occurs while the protocol switching.
type UpgradeErrorHandler func(response *goyave.Response, request *goyave.Request, status int, reason error)

func defaultUpgradeErrorHandler(response *goyave.Response, _ *goyave.Request, status int, reason error) {
	text := http.StatusText(status)
	if config.GetBool("app.debug") && reason != nil {
		text = reason.Error()
	}
	message := map[string]string{
		"error": text,
	}
	response.JSON(status, message)
}

// ErrorHandler is a specific Handler type for handling errors returned by websocket Handler.
type ErrorHandler func(request *goyave.Request, err error)

// Upgrader is responsible for the upgrade of HTTP connections to
// websocket connections.
type Upgrader struct {
	// UpgradeErrorHandler specifies the function for generating HTTP error responses.
	//
	// The default UpgradeErrorHandler returns a JSON response containing the status text
	// corresponding to the status code returned. If debugging is enabled, the reason error
	// message is returned instead.
	//
	//  {"error": "message"}
	UpgradeErrorHandler UpgradeErrorHandler

	// ErrorHandler specifies the function handling errors returned by websocket Handler.
	// If nil, the error is written to "goyave.ErrLogger". If the error is caused by
	// a panic and debugging is enabled, the default ErrorHandler also writes the stacktrace.
	ErrorHandler ErrorHandler

	// CheckOrigin returns true if the request Origin header is acceptable. If
	// CheckOrigin is nil, then a safe default is used: return false if the
	// Origin request header is present and the origin host is not equal to
	// request Host header.
	//
	// A CheckOrigin function should carefully validate the request origin to
	// prevent cross-site request forgery.
	CheckOrigin func(r *goyave.Request) bool

	// Headers function generating headers to be sent with the protocol switching response.
	Headers func(request *goyave.Request) http.Header

	// Settings the parameters for upgrading the connection. "Error" and "CheckOrigin" are
	// ignored: use the Goyave upgrader's "UpgradeErrorHandler" and "CheckOrigin".
	Settings ws.Upgrader
}

func (u *Upgrader) makeUpgrader(request *goyave.Request) *ws.Upgrader {
	upgradeErrorHandler := u.UpgradeErrorHandler
	if upgradeErrorHandler == nil {
		upgradeErrorHandler = defaultUpgradeErrorHandler
	}
	a := adapter{
		upgradeErrorHandler: upgradeErrorHandler,
		checkOrigin:         u.CheckOrigin,
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
func (u *Upgrader) Handler(handler Handler) goyave.Handler {
	setTimeout()
	return func(response *goyave.Response, request *goyave.Request) {
		var headers http.Header
		if u.Headers != nil {
			headers = u.Headers(request)
		}

		c, err := u.makeUpgrader(request).Upgrade(response, request.Request(), headers)
		if err != nil {
			return
		}
		response.Status(http.StatusSwitchingProtocols)

		go u.serve(c, request, handler)
	}
}

func (u *Upgrader) serve(c *ws.Conn, request *goyave.Request, handler Handler) {
	conn := newConn(c)
	panicked := true
	var err error
	defer func() { // Panic recovery
		if panicReason := recover(); panicReason != nil || panicked {
			stack := ""
			if config.GetBool("app.debug") {
				stack = string(debug.Stack())
			}

			if e, ok := panicReason.(error); ok {
				err = fmt.Errorf("%w", e)
			} else {
				err = fmt.Errorf("%v", panicReason)
			}

			err = &PanicError{
				Reason:     err,
				Stacktrace: stack,
			}
		}

		if IsCloseError(err) {
			conn.CloseNormal()
			return
		}
		if err != nil {
			if u.ErrorHandler != nil {
				u.ErrorHandler(request, err)
			} else {
				goyave.ErrLogger.Println(err)
				if e, ok := err.(*PanicError); ok && e.Stacktrace != "" {
					goyave.ErrLogger.Println(e.Stacktrace)
				}
			}
			conn.CloseWithError(err)
		} else {
			conn.internalClose(ws.CloseNormalClosure, NormalClosureMessage)
		}
	}()

	err = handler(conn, request)
	panicked = false
}

type adapter struct {
	upgradeErrorHandler UpgradeErrorHandler
	checkOrigin         func(r *goyave.Request) bool
	request             *goyave.Request
}

func (a *adapter) onError(w http.ResponseWriter, _ *http.Request, status int, reason error) {
	if status == http.StatusInternalServerError {
		panic(reason)
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

func truncateMessage(message string, maxLength int) string {
	if len([]byte(message)) <= maxLength {
		return message
	}

	res := make([]rune, 0, maxLength)
	count := 0
	for _, r := range message {
		l := utf8.RuneLen(r)
		if count+l > maxLength {
			break
		}

		count += l
		res = append(res, r)
	}

	return string(res)
}
