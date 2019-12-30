package cors

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/System-Glitch/goyave/v2"
	"github.com/System-Glitch/goyave/v2/helper"
)

// Middleware is the middleware handling CORS, using the given options.
// This middleware is automatically inserted first to the router's list of middleware
// if it has defined CORS Options.
func (o *Options) Middleware() goyave.Middleware {
	return func(next goyave.Handler) goyave.Handler {
		return func(response *goyave.Response, request *goyave.Request) {
			o.configureCommon(response, request)

			if request.Method() == http.MethodOptions && request.Header().Get("Access-Control-Request-Method") != "" {
				o.handlePreflight(response, request)
				if o.OptionsPassthrough {
					next(response, request)
				} else {
					response.WriteHeader(http.StatusNoContent)
				}
			} else {
				next(response, request)
			}
		}
	}
}

func (o *Options) configureCommon(response *goyave.Response, request *goyave.Request) {
	o.configureOrigin(response, request)
	o.configureCredentials(response)
	o.configureExposedHeaders(response, request)
}

func (o *Options) configureOrigin(response *goyave.Response, request *goyave.Request) {
	if len(o.AllowedOrigins) == 0 || o.AllowedOrigins[0] == "*" {
		response.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		headers := response.Header()
		if o.validateOrigin(request) {
			headers.Set("Access-Control-Allow-Origin", request.Header().Get("Origin"))
		}
		headers.Add("Vary", "Origin")
	}
}

func (o *Options) configureCredentials(response *goyave.Response) {
	if o.AllowCredentials {
		response.Header().Set("Access-Control-Allow-Credentials", "true")
	}
}

func (o *Options) configureExposedHeaders(response *goyave.Response, request *goyave.Request) {
	request.Header().Set("Access-Control-Expose-Headers", strings.Join(o.ExposedHeaders, ", "))
}

func (o *Options) configureAllowedMethods(response *goyave.Response) {
	headers := response.Header()
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Set("Access-Control-Allow-Methods", strings.Join(o.AllowedMethods, ", "))
}

func (o *Options) configureAllowedHeaders(response *goyave.Response, request *goyave.Request) {
	headers := response.Header()
	if len(o.AllowedHeaders) == 0 {
		headers.Add("Vary", "Access-Control-Request-Headers")
		headers.Set("Access-Control-Allow-Headers", request.Header().Get("Access-Control-Request-Headers"))
	} else {
		headers.Set("Access-Control-Allow-Headers", strings.Join(o.AllowedHeaders, ", "))
	}

}

func (o *Options) configureMaxAge(response *goyave.Response, request *goyave.Request) {
	response.Header().Set("Access-Control-Max-Age", strconv.FormatUint(uint64(o.MaxAge.Seconds()), 10))
}

func (o *Options) handlePreflight(response *goyave.Response, request *goyave.Request) {
	o.configureAllowedMethods(response)
	o.configureAllowedHeaders(response, request)
	o.configureMaxAge(response, request)
}

func (o *Options) validateOrigin(request *goyave.Request) bool {
	return len(o.AllowedOrigins) == 0 ||
		o.AllowedOrigins[0] == "*" ||
		helper.Contains(o.AllowedOrigins, request.Header().Get("Origin"))
}
