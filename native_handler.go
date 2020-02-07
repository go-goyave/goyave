package goyave

import (
	"net/http"

	"github.com/gorilla/mux"
)

// NativeHandler is an adapter function for "http.Handler".
// With this adapter, you can plug non-Goyave handlers to your application.
//
// If the request is a JSON request, the native handler will not be able to
// read the body, as it has already been parsed by the framework and is stored in
// the "goyave.Request" object. However, form data can be accessed as usual.
// Just remember that it contains the raw data, which haven't been validated
// nor converted. This means that native handlers are not guaranteed to work and
// cannot modify the request data.
// Prefer implementing a Goyave handler.
//
// This feature is a compatibility layer with the rest of the Golang web ecosystem.
// Prefer using Goyave handlers if possible.
func NativeHandler(handler http.Handler) Handler {
	return func(response *Response, request *Request) {
		handler.ServeHTTP(response, request.httpRequest)
	}
}

// NativeMiddleware is an adapter function "mux.MiddlewareFunc".
//
// Deprecated: Goyave doesn't use gorilla/mux anymore. This function will be removed
// in a future major release.
//
// With this adapter, you can plug Gorilla Mux middleware to your application.
//
// Native middleware work like native handlers. See "NativeHandler" for more details.
func NativeMiddleware(middleware mux.MiddlewareFunc) Middleware {
	return func(next Handler) Handler {
		return func(response *Response, request *Request) {
			middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next(response, request)
			})).ServeHTTP(response, request.httpRequest)
		}
	}
}
