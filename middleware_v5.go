package goyave

import (
	"net/http"
	"runtime/debug"
	"strings"

	"goyave.dev/goyave/v4/cors"
	"goyave.dev/goyave/v4/validation"
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

// GetMiddleware returns a copy of the middleware applied on this holder.
func (h *middlewareHolderV5) GetMiddleware() []MiddlewareV5 {
	return append(make([]MiddlewareV5, 0, len(h.middleware)), h.middleware...)
}

func findMiddleware[T MiddlewareV5](m []MiddlewareV5) T {
	for _, middleware := range m {
		if m, ok := middleware.(T); ok {
			return m
		}
	}
	var zero T
	return zero
}

func hasMiddleware[T MiddlewareV5](m []MiddlewareV5) bool {
	for _, middleware := range m {
		if _, ok := middleware.(T); ok {
			return true
		}
	}
	return false
}

// routeHasMiddleware returns true if the given route or any of its
// parents has a middleware of the T type.
func routeHasMiddleware[T MiddlewareV5](route *RouteV5) bool {
	return hasMiddleware[T](route.middleware)
}

// routerHasMiddleware returns true if the given route or any of its
// parents has a middleware of the T type.
func routerHasMiddleware[T MiddlewareV5](router *RouterV5) bool {
	return hasMiddleware[T](router.middleware) || (router.parent != nil && routerHasMiddleware[T](router.parent))
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
		if header := request.Header().Get("Accept-Language"); len(header) > 0 {
			request.Lang = m.Lang().DetectLanguage(header)
		} else {
			request.Lang = m.Lang().GetDefault()
		}
		next(response, request)
	}
}

// validateRequestMiddleware is a middleware that validates the request.
// If validation is not rules are not met, sets the response status to 422 Unprocessable Entity
// or 400 Bad Request and the response error (which can be retrieved with `GetError()`) to the
// `validation.Errors` returned by the validator.
// This data can then be used in a status handler.
type validateRequestMiddlewareV5 struct {
	Controller
	BodyRules  RuleSetFunc
	QueryRules RuleSetFunc
}

func (m *validateRequestMiddlewareV5) Handle(next HandlerV5) HandlerV5 {
	return func(response *ResponseV5, r *RequestV5) {
		extra := map[string]any{
			validation.ExtraRequest: r,
		}
		contentType := r.Header().Get("Content-Type")

		var errsBag *validation.ErrorsV5
		var queryErrsBag *validation.ErrorsV5
		var errors []error
		if m.QueryRules != nil {
			opt := &validation.Options{
				Data:                     r.Query,
				Rules:                    m.QueryRules(r),
				ConvertSingleValueArrays: true,
				Language:                 r.Lang,
				Extra:                    extra,
			}
			var err []error
			queryErrsBag, err = validation.ValidateV5(opt)
			if queryErrsBag != nil {
				r.Extra[ExtraQueryValidationError] = queryErrsBag
			}
			if err != nil {
				errors = append(errors, err...)
			}
		}
		if m.BodyRules != nil {
			opt := &validation.Options{
				Data:                     r.Data,
				Rules:                    m.BodyRules(r),
				ConvertSingleValueArrays: !strings.HasPrefix(contentType, "application/json"),
				Language:                 r.Lang,
				Extra:                    extra,
			}
			var err []error
			errsBag, err = validation.ValidateV5(opt)
			if errsBag != nil {
				r.Extra[ExtraValidationError] = errsBag
			}
			if err != nil {
				errors = append(errors, err...)
			}
			r.Data = opt.Data
		}

		if errors != nil && len(errors) != 0 {
			response.Error(errors)
			return
		}

		if errsBag != nil || queryErrsBag != nil {
			response.Status(http.StatusUnprocessableEntity)
			return
		}

		next(response, r)
	}
}

type corsMiddlewareV5 struct {
	Controller
}

func (m *corsMiddlewareV5) Handle(next HandlerV5) HandlerV5 {
	return func(response *ResponseV5, request *RequestV5) {
		o, ok := request.Route().LookupMeta(MetaCORS)
		if !ok || o == nil {
			next(response, request)
			return
		}

		options := o.(*cors.Options)
		headers := response.Header()
		requestHeaders := request.Header()

		options.ConfigureCommon(headers, requestHeaders)

		if request.Method() == http.MethodOptions && requestHeaders.Get("Access-Control-Request-Method") != "" {
			options.HandlePreflight(headers, requestHeaders)
			if options.OptionsPassthrough {
				next(response, request)
			} else {
				response.WriteHeader(http.StatusNoContent)
			}
		} else {
			next(response, request)
		}
	}
}
