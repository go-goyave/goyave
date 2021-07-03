package goyave

import (
	"net/http"
)

// NativeMiddlewareFunc is a function which receives an http.Handler and returns another http.Handler.
type NativeMiddlewareFunc func(http.Handler) http.Handler

// NativeHandler is an adapter function for "http.Handler".
// With this adapter, you can plug non-Goyave handlers to your application.
//
// Just remember that the body contains the raw data, which haven't been validated
// nor converted. This means that native handlers are not guaranteed to work and
// cannot modify the request data. Request properties, such as headers, can still
// be modified.
// Prefer implementing a Goyave handler.
//
// This feature is a compatibility layer with the rest of the Golang web ecosystem.
// Prefer using Goyave handlers if possible.
func NativeHandler(handler http.Handler) Handler {
	return func(response *Response, request *Request) {
		handler.ServeHTTP(response, request.httpRequest)
	}
}

// NativeMiddleware is an adapter function for standard library middleware.
//
// Native middleware work like native handlers. See "NativeHandler" for more details.
func NativeMiddleware(middleware NativeMiddlewareFunc) Middleware {
	return func(next Handler) Handler {
		return func(response *Response, request *Request) {
			middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if request.httpRequest != r {
					request.httpRequest = r
				}
				// FIXME if a native middleware replaces the http.ResponseWriter, it
				// may not work as expected.
				next(response, request)
			})).ServeHTTP(response, request.httpRequest)
		}
	}
}
