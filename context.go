package goyave

import (
	"net/http"
	"runtime/debug"

	"gorm.io/gorm"
	"goyave.dev/goyave/v4/config"
)

const (
	// ExtraError the key used in Context.Extra to store an error
	// reported with the Error function or via the recovery middleware.
	ExtraError = "goyave.error"

	// ExtraStacktrace the key used in Context.Extra to store the
	// stacktrace if debug is enabled and an error is reported.
	ExtraStacktrace = "goyave.stacktrace"
)

type Context[T any] struct {
	*RequestV5[T]
	*ResponseV5
	server      *Server
	route       *Route
	RouteParams map[string]string
	Extra       map[string]any
}

func (c *Context[T]) Server() *Server {
	return c.server
}

func (c *Context[T]) DB() *gorm.DB {
	return c.server.DB()
}

func (c *Context[T]) Config() *config.Config {
	return c.server.Config()
}

// Route returns the current route.
func (c *Context[T]) Route() *Route {
	return c.route
}

// Error print the error in the console and return it with an error code 500 (or previously defined
// status code using `response.Status()`).
// If debugging is enabled in the config, the error is also written in the response
// and the stacktrace is printed in the console.
// If debugging is not enabled, only the status code is set, which means you can still
// write to the response, or use your error status handler.
func (c *Context[T]) Error(err any) {
	c.server.ErrLogger.Println(err)
	c.error(err)
}

func (c *Context[T]) error(err any) {
	c.Extra[ExtraError] = err
	if c.server.Config().GetBool("app.debug") {
		stacktrace := c.Extra[ExtraStacktrace]
		if stacktrace == "" {
			stacktrace = string(debug.Stack())
		}
		c.server.ErrLogger.Print(stacktrace)
		if !c.Hijacked() {
			var message interface{}
			if e, ok := err.(error); ok {
				message = e.Error()
			} else {
				message = err
			}
			status := http.StatusInternalServerError
			if c.status != 0 {
				status = c.status
			}
			c.JSON(status, map[string]interface{}{"error": message})
			return
		}
	}

	// Don't set r.empty to false to let error status handler process the error
	c.Status(http.StatusInternalServerError)
}

func (c *Context[T]) Clone() *Context[T] {
	return &Context[T]{
		RequestV5:   c.RequestV5,
		ResponseV5:  c.ResponseV5,
		server:      c.server,
		route:       c.route,
		RouteParams: c.RouteParams,
		Extra:       c.Extra,
	}
}
