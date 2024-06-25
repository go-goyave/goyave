package goyave

import (
	"net/http"
	"strings"

	"gorm.io/gorm"
	"goyave.dev/goyave/v5/cors"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/validation"
)

// Middleware are special handlers executed in a stack above the controller handler.
// They allow to inspect and filter requests, transform responses or provide additional
// information to the next handlers in the stack.
// Example uses are authentication, authorization, logging, panic recovery, CORS,
// validation, gzip compression.
type Middleware interface {
	Composable
	Handle(next Handler) Handler
}

type middlewareHolder struct {
	middleware []Middleware
}

func (h *middlewareHolder) applyMiddleware(handler Handler) Handler {
	for i := len(h.middleware) - 1; i >= 0; i-- {
		handler = h.middleware[i].Handle(handler)
	}
	return handler
}

// GetMiddleware returns a copy of the middleware applied on this holder.
func (h *middlewareHolder) GetMiddleware() []Middleware {
	return append(make([]Middleware, 0, len(h.middleware)), h.middleware...)
}

func findMiddleware[T Middleware](m []Middleware) T {
	for _, middleware := range m {
		if m, ok := middleware.(T); ok {
			return m
		}
	}
	var zero T
	return zero
}

func hasMiddleware[T Middleware](m []Middleware) bool {
	for _, middleware := range m {
		if _, ok := middleware.(T); ok {
			return true
		}
	}
	return false
}

// routeHasMiddleware returns true if the given route or any of its
// parents has a middleware of the T type.
func routeHasMiddleware[T Middleware](route *Route) bool {
	return hasMiddleware[T](route.middleware)
}

// routerHasMiddleware returns true if the given route or any of its
// parents has a middleware of the T type. Also returns true if the middleware
// is present as global middleware.
func routerHasMiddleware[T Middleware](router *Router) bool {
	return hasMiddleware[T](router.globalMiddleware.middleware) || hasMiddleware[T](router.middleware) || (router.parent != nil && routerHasMiddleware[T](router.parent))
}

// recoveryMiddleware is a middleware that recovers from panic and sends a 500 error code.
// If debugging is enabled in the config and the default status handler for the 500 status code
// had not been changed, the error is also written in the response.
type recoveryMiddleware struct {
	Component
}

func (m *recoveryMiddleware) Handle(next Handler) Handler {
	return func(response *Response, request *Request) {
		panicked := true
		defer func() {
			if err := recover(); err != nil || panicked {
				e := errors.NewSkip(err, 4).(*errors.Error) // Skipped: runtime.Callers, NewSkip, this func, runtime.panic
				m.Logger().Error(e)
				response.err = e
				response.status = http.StatusInternalServerError // Force status override
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
type languageMiddleware struct {
	Component
}

func (m *languageMiddleware) Handle(next Handler) Handler {
	return func(response *Response, request *Request) {
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
// This middleware requires the parse middleware.
type validateRequestMiddleware struct {
	Component
	BodyRules  RuleSetFunc
	QueryRules RuleSetFunc
}

func (m *validateRequestMiddleware) Handle(next Handler) Handler {
	return func(response *Response, r *Request) {
		extra := map[any]any{
			validation.ExtraRequest{}: r,
		}
		contentType := r.Header().Get("Content-Type")

		var db *gorm.DB
		if m.Config().GetString("database.connection") != "none" {
			db = m.DB().WithContext(r.Context())
		}
		var errsBag *validation.Errors
		var queryErrsBag *validation.Errors
		var errors []error
		if m.QueryRules != nil {
			opt := &validation.Options{
				Data:                     r.Query,
				Rules:                    m.QueryRules(r).AsRules(),
				ConvertSingleValueArrays: true,
				Language:                 r.Lang,
				DB:                       db,
				Config:                   m.Config(),
				Logger:                   m.Logger(),
				Extra:                    extra,
			}
			r.Extra[ExtraQueryValidationRules{}] = opt.Rules
			var err []error
			queryErrsBag, err = validation.Validate(opt)
			if queryErrsBag != nil {
				r.Extra[ExtraQueryValidationError{}] = queryErrsBag
			}
			if err != nil {
				errors = append(errors, err...)
			}
		}
		if m.BodyRules != nil {
			opt := &validation.Options{
				Data:                     r.Data,
				Rules:                    m.BodyRules(r).AsRules(),
				ConvertSingleValueArrays: !strings.HasPrefix(contentType, "application/json"),
				Language:                 r.Lang,
				DB:                       db,
				Config:                   m.Config(),
				Logger:                   m.Logger(),
				Extra:                    extra,
			}
			r.Extra[ExtraBodyValidationRules{}] = opt.Rules
			var err []error
			errsBag, err = validation.Validate(opt)
			if errsBag != nil {
				r.Extra[ExtraValidationError{}] = errsBag
			}
			if err != nil {
				errors = append(errors, err...)
			}
			r.Data = opt.Data
		}

		if len(errors) != 0 {
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

type corsMiddleware struct {
	Component
}

func (m *corsMiddleware) Handle(next Handler) Handler {
	return func(response *Response, request *Request) {
		o, ok := request.Route.LookupMeta(MetaCORS)
		if !ok || o == nil || o == (*cors.Options)(nil) {
			next(response, request)
			return
		}

		options := o.(*cors.Options)
		headers := response.Header()
		requestHeaders := request.Header()

		if request.Method() == http.MethodOptions && requestHeaders.Get("Access-Control-Request-Method") == "" {
			response.Status(http.StatusBadRequest)
			return
		}

		options.ConfigureCommon(headers, requestHeaders)

		if request.Method() == http.MethodOptions {
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
