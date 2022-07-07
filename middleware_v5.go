package goyave

import (
	"net/http"
	"runtime/debug"
)

type MiddlewareV5 interface {
	IController
	Handle(HandlerV5) HandlerV5
}

type middlewareHolderV5 struct {
	middleware []MiddlewareV5
}

func (h *middlewareHolderV5) applyMiddleware(handler HandlerV5) HandlerV5 {
	for i := len(h.middleware) - 1; i >= 0; i-- {
		handler = h.middleware[i].Handle(handler)
	}
	return handler
}

// recoveryMiddleware is a middleware that recovers from panic and sends a 500 error code.
// If debugging is enabled in the config and the default status handler for the 500 status code
// had not been changed, the error is also written in the response.
type recoveryMiddlewareV5 struct {
	Controller
}

func (m *recoveryMiddlewareV5) Handle(next HandlerV5) HandlerV5 {
	return func(response *ResponseV5, request *RequestV5) {
		panicked := true
		defer func() {
			if err := recover(); err != nil || panicked {
				m.ErrLogger().Println(err)
				request.Extra[ExtraError] = err
				if m.Config().GetBool("app.debug") {
					request.Extra[ExtraStacktrace] = string(debug.Stack())
				}
				response.Status(http.StatusInternalServerError)
			}
		}()

		next(response, request)
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
type languageMiddlewareV5 struct {
	Controller
}

func (m *languageMiddlewareV5) Handle(next HandlerV5) HandlerV5 {
	return func(response *ResponseV5, request *RequestV5) {
		if header := request.RequestHeader().Get("Accept-Language"); len(header) > 0 {
			request.Lang = m.Lang().DetectLanguage(header)
		} else {
			request.Lang = m.Lang().Default
		}
		next(response, request)
	}
}
