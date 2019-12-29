package cors

import (
	"net/http"

	"github.com/System-Glitch/goyave/v2"
	"github.com/System-Glitch/goyave/v2/helper"
)

// Middleware is the middleware handling CORS, using the given options.
// This middleware is automatically inserted first to the router's list of middleware
// if it has defined CORS Options.
func (o *Options) Middleware() goyave.Middleware {
	return func(next goyave.Handler) goyave.Handler {
		return func(response *goyave.Response, request *goyave.Request) {
			if request.Method() == http.MethodOptions && request.Header().Get("Access-Control-Request-Method") != "" {
				// TODO preflight
				if o.OptionsPassthrough {
					next(response, request)
				} else {
					response.WriteHeader(http.StatusNoContent)
				}
			} else {
				// TODO actual request
				next(response, request)
			}
		}
	}
}

func (o *Options) handlePreflight(response *goyave.Response, request *goyave.Request) {

}

func (o *Options) handleRequest(response *goyave.Response, request *goyave.Request) bool {
	headers := response.Header()
	origin := request.Header().Get("Origin")

	headers.Add("Vary", "Origin")

	if origin == "" {
		return true
	}

	if !o.validateOrigin(request) {
		response.Status(http.StatusForbidden)
		return false
	}

	return true
}

func (o *Options) validateOrigin(request *goyave.Request) bool {
	return len(o.AllowedOrigins) == 0 ||
		o.AllowedOrigins[0] == "*" ||
		helper.Contains(o.AllowedOrigins, request.Header().Get("Origin"))
}

func (o *Options) validateMethod(request *goyave.Request) bool {
	return len(o.AllowedMethods) == 0 || helper.Contains(o.AllowedMethods, request.Method())
}

func (o *Options) validateHeaders(request *goyave.Request) bool {
	if len(o.AllowedHeaders) == 0 {
		return true
	}

	for _, h := range request.Header() {
		if !helper.Contains(o.AllowedHeaders, h) {
			return false
		}
	}

	return true
}
