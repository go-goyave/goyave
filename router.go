package goyave

import (
	"net/http"
	"strings"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/helper/filesystem"
	"github.com/System-Glitch/goyave/v2/validation"
	"github.com/gorilla/mux"
)

// Router registers routes to be matched and dispatches a handler.
type Router struct {
	muxRouter  *mux.Router
	middleware []Middleware
}

// Handler is a controller or middleware function
type Handler func(*Response, *Request)

func newRouter() *Router {
	muxRouter := mux.NewRouter()
	muxRouter.Schemes(config.GetString("protocol"))
	router := &Router{muxRouter: muxRouter}
	router.Middleware(recoveryMiddleware, parseRequestMiddleware, languageMiddleware)
	return router
}

// Subrouter create a new sub-router from this router.
// Use subrouters to create route groups and to apply middleware to multiple routes.
func (r *Router) Subrouter(prefix string) *Router {
	router := &Router{muxRouter: r.muxRouter.PathPrefix(prefix).Subrouter()}

	// Apply parent middleware to subrouter
	router.Middleware(r.middleware...)
	return router
}

// Middleware apply one or more middleware(s) to the route group.
func (r *Router) Middleware(middleware ...Middleware) {
	r.middleware = append(r.middleware, middleware...)
}

// Route register a new route.
//
// Multiple methods can be passed using a pipe-separated string.
//  "PUT|PATCH"
//
// The validation rules set is optional. If you don't want your route
// to be validated, pass "nil".
func (r *Router) Route(methods string, uri string, handler Handler, validationRules validation.RuleSet) {
	r.muxRouter.HandleFunc(uri, func(w http.ResponseWriter, rawRequest *http.Request) {
		r.requestHandler(w, rawRequest, handler, validationRules)
	}).Methods(strings.Split(methods, "|")...)
}

// Static serve a directory and its subdirectories of static resources.
// Set the "download" parameter to true if you want the files to be sent as an attachment
// instead of an inline element.
//
// If no file is given in the url, or if the given file is a directory, the handler will
// send the "index.html" file if it exists.
func (r *Router) Static(uri string, directory string, download bool) {
	r.Route("GET", uri+"{resource:.*}", staticHandler(directory, download), nil)
}

func staticHandler(directory string, download bool) Handler {
	return func(response *Response, r *Request) {
		file := r.Params["resource"]
		path := cleanStaticPath(directory, file)

		if filesystem.FileExists(path) {
			if download {
				response.Download(path, file[strings.LastIndex(file, "/")+1:])
			} else {
				response.File(path)
			}
		} else {
			response.Status(http.StatusNotFound)
		}
	}
}

func cleanStaticPath(directory string, file string) string {
	if strings.HasPrefix(file, "/") {
		file = file[1:]
	}
	path := directory + "/" + file
	if filesystem.IsDirectory(path) {
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		path += "index.html"
	}
	return path
}

func (r *Router) requestHandler(w http.ResponseWriter, rawRequest *http.Request, handler Handler, rules validation.RuleSet) {
	request := &Request{
		httpRequest: rawRequest,
		Rules:       rules,
		Params:      mux.Vars(rawRequest),
	}
	response := &Response{
		httpRequest:    rawRequest,
		ResponseWriter: w,
		empty:          true,
	}

	// Validate last.
	// Allows custom middleware to be executed after core
	// middleware and before validation.
	handler = validateRequestMiddleware(handler)
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handler = r.middleware[i](handler)
	}

	handler(response, request)

	// If the response is empty, return status 204 to
	// comply with RFC 7231, 6.3.5
	if response.empty {
		response.Status(http.StatusNoContent)
	}
}
