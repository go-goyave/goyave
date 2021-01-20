package websocket

import (
	"net/http"

	"github.com/System-Glitch/goyave/v3"
	"github.com/System-Glitch/goyave/v3/config"

	ws "github.com/gorilla/websocket"
)

// TODO test websocket.go

// HandlerFunc is a handler for websocket connections.
// The request parameter contains the original upgraded HTTP request.
//
// To keep connection alive, these handlers should run an infinite for loop
// and check for close errors.
//
// The following HandlerFunc is an example of an "echo" feature using websockets:
//
//  func(c *websocket.Conn, request *goyave.Request) error {
//  	for {
//  		mt, message, err := c.ReadMessage()
//  		if err != nil {
//  			if websocket.IsCloseError(err) {
//  				return nil
//  			}
//  			return fmt.Errorf("read: %v", err)
//  		}
//  		goyave.Logger.Printf("recv: %s", message)
//  		err = c.WriteMessage(mt, message)
//  		if err != nil {
//  			return fmt.Errorf("write: %v", err)
//  		}
//  	}
//  }
type HandlerFunc func(*Conn, *goyave.Request) error

// UpgradeErrorHandler is a specific Handler type for connection upgrade errors.
// These handlers are called when an error occurs while the protocol switching.
type UpgradeErrorHandler func(response *goyave.Response, request *goyave.Request, status int, reason error)

func defaultUpgradeErrorHandler(response *goyave.Response, request *goyave.Request, status int, reason error) {
	text := http.StatusText(status)
	if config.GetBool("app.debug") && reason != nil {
		text = reason.Error()
	}
	message := map[string]string{
		"error": text,
	}
	response.JSON(status, message)
}

// ErrorHandler is a specific Handler type for handling errors returned by HandlerFunc.
type ErrorHandler func(request *goyave.Request, err error)

// Conn represents a WebSocket connection.
type Conn struct {
	*ws.Conn
}

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

	// ErrorHandler specifies the function handling errors returned by HandlerFunc.
	// If nil, the error is written to "goyave.ErrLogger".
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
	// ignored: this the Goyave upgrader's "ErrorHandler" and "CheckOrigin".
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
// to the given HandlerFunc.
//
// HTTP response's status is set to "101 Switching Protocols". The connection is closed
// automatically after the HandlerFunc returns or panics. If an error is returned,
// it will be printed to "goyave.ErrLogger".
// Bear in mind that the recovery middleware doesn't work on websocket connections,
// as we are not in an HTTP context anymore.
func (u *Upgrader) Handler(handler HandlerFunc) goyave.Handler {
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
		defer c.Close()
		if err := handler(&Conn{c}, request); err != nil {
			if u.ErrorHandler != nil {
				u.ErrorHandler(request, err)
			} else {
				goyave.ErrLogger.Println(err)
			}
		}
	}
}

type adapter struct {
	upgradeErrorHandler UpgradeErrorHandler
	checkOrigin         func(r *goyave.Request) bool
	request             *goyave.Request
}

func (a *adapter) onError(w http.ResponseWriter, r *http.Request, status int, reason error) {
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
