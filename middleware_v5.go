package goyave

import (
	"net/http"
	"runtime/debug"
)

// Middleware function generating middleware handler function.
//
// Request data is available to middleware, but bear in mind that
// it had not been validated yet. That means that you can modify or
// filter data.
type MiddlewareV5 func(HandlerV5) HandlerV5

type middlewareHolderV5 struct {
	middleware []MiddlewareV5
}

func (h *middlewareHolderV5) applyMiddleware(handler HandlerV5) HandlerV5 {
	for i := len(h.middleware) - 1; i >= 0; i-- {
		handler = h.middleware[i](handler)
	}
	return handler
}

// recoveryMiddleware is a middleware that recovers from panic and sends a 500 error code.
// If debugging is enabled in the config and the default status handler for the 500 status code
// had not been changed, the error is also written in the response.
func recoveryMiddlewareV5(next HandlerV5) HandlerV5 {
	return func(c *Context) {
		panicked := true
		defer func() {
			if err := recover(); err != nil || panicked {
				c.Server().ErrLogger.Println(err)
				c.Extra[ExtraError] = err
				if c.Config().GetBool("app.debug") {
					c.Extra[ExtraStacktrace] = string(debug.Stack())
				}
				c.Status(http.StatusInternalServerError)
			}
		}()

		next(c)
		panicked = false
	}
}

// languageMiddleware is a middleware that sets the language of a request.
//
// Uses the "Accept-Language" header to determine which language to use. If
// the header is not set or the language is not available, uses the default
// language as fallback.
//
// If "*" is provided, the default language will be used.
// If multiple languages are given, the first available language will be used,
// and if none are available, the default language will be used.
// If no variant is given (for example "en"), the first available variant will be used.
// For example, if "en-US" and "en-UK" are available and the request accepts "en",
// "en-US" will be used.
func languageMiddlewareV5(next HandlerV5) HandlerV5 {
	return func(c *Context) {
		if header := c.RequestHeader().Get("Accept-Language"); len(header) > 0 {
			c.Lang = c.Server().Lang.DetectLanguage(header)
		} else {
			c.Lang = c.Server().Lang.Default
		}
		next(c)
	}
}
