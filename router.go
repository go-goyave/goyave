package goyave

import (
	"net/http"
	"strings"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/cors"
	"github.com/System-Glitch/goyave/v2/helper/filesystem"
	"github.com/System-Glitch/goyave/v2/validation"
	"github.com/gorilla/mux"
)

// Router registers routes to be matched and executes a handler.
type Router struct {
	muxRouter         *mux.Router
	corsOptions       *cors.Options
	parent            *Router
	hasCORSMiddleware bool
	middleware        []Middleware
	statusHandlers    map[int]Handler
}

// Handler is a controller or middleware function
type Handler func(*Response, *Request)

func panicStatusHandler(response *Response, request *Request) {
	response.Error(response.GetError())
	if response.empty {
		message := map[string]string{
			"error": http.StatusText(response.GetStatus()),
		}
		response.JSON(response.GetStatus(), message)
	}
}

func errorStatusHandler(response *Response, request *Request) {
	message := map[string]string{
		"error": http.StatusText(response.GetStatus()),
	}
	response.JSON(response.GetStatus(), message)
}

func (r *Router) muxStatusHandler(status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, rawRequest *http.Request) {
		r.requestHandler(w, rawRequest, func(response *Response, r *Request) {
			response.Status(status)
		}, nil)
	})
}

func newRouter() *Router {
	muxRouter := mux.NewRouter()
	muxRouter.Schemes(config.GetString("protocol"))
	router := &Router{
		muxRouter:      muxRouter,
		statusHandlers: make(map[int]Handler, 13),
		middleware:     make([]Middleware, 0, 3),
		parent:         nil,
	}
	muxRouter.NotFoundHandler = router.muxStatusHandler(http.StatusNotFound)
	muxRouter.MethodNotAllowedHandler = router.muxStatusHandler(http.StatusMethodNotAllowed)
	router.StatusHandler(panicStatusHandler, http.StatusInternalServerError)
	router.StatusHandler(errorStatusHandler, 404, 405, 501, 502, 503, 504, 505, 506, 507, 508, 510, 511)
	router.Middleware(recoveryMiddleware, parseRequestMiddleware, languageMiddleware)
	return router
}

func (r *Router) copyStatusHandlers() map[int]Handler {
	cpy := make(map[int]Handler, len(r.statusHandlers))
	for key, value := range r.statusHandlers {
		cpy[key] = value
	}
	return cpy
}

// Subrouter create a new sub-router from this router.
// Use subrouters to create route groups and to apply middleware to multiple routes.
// CORS options are also inherited.
func (r *Router) Subrouter(prefix string) *Router {
	router := &Router{
		muxRouter:         r.muxRouter.PathPrefix(prefix).Subrouter(),
		parent:            r,
		corsOptions:       r.corsOptions,
		hasCORSMiddleware: r.hasCORSMiddleware,
		statusHandlers:    r.copyStatusHandlers(),
		middleware:        make([]Middleware, 0, 3),
	}
	router.muxRouter.NotFoundHandler = r.muxRouter.NotFoundHandler
	router.muxRouter.MethodNotAllowedHandler = r.muxRouter.MethodNotAllowedHandler
	return router
}

// Middleware apply one or more middleware to the route group.
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
//
// If the router has CORS options set, the "OPTIONS" method is automatically added
// to the matcher if it's missing, so it allows preflight requests.
func (r *Router) Route(methods string, uri string, handler Handler, validationRules validation.RuleSet) {
	r.route(methods, uri, handler, validationRules)
}

func (r *Router) route(methods string, uri string, handler Handler, validationRules validation.RuleSet) *mux.Route {
	if r.corsOptions != nil && !strings.Contains(methods, "OPTIONS") {
		methods += "|OPTIONS"
	}

	return r.muxRouter.HandleFunc(uri, func(w http.ResponseWriter, rawRequest *http.Request) {
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

// CORS set the CORS options for this route group.
// If the options are not nil, the CORS middleware is automatically added.
func (r *Router) CORS(options *cors.Options) {
	r.corsOptions = options
	if options != nil && !r.hasCORSMiddleware {
		r.Middleware(corsMiddleware)
		r.hasCORSMiddleware = true
	}
}

// StatusHandler set a handler for responses with an empty body.
// The handler will be automatically executed if the request's life-cycle reaches its end
// and nothing has been written in the response body.
//
// Multiple status codes can be given. The handler will be executed if one of them matches.
//
// This method can be used to define custom error handlers for example.
//
// Status handlers are inherited as a copy in sub-routers. Modifying a child's status handler
// will not modify its parent's.
//
// Codes in the 500 range and codes 404 and 405 have a default status handler.
func (r *Router) StatusHandler(handler Handler, status int, additionalStatuses ...int) {
	r.statusHandlers[status] = handler
	for _, s := range additionalStatuses {
		r.statusHandlers[s] = handler
	}
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
		corsOptions: r.corsOptions,
		Rules:       rules,
		Params:      mux.Vars(rawRequest),
	}
	response := &Response{
		httpRequest:    rawRequest,
		ResponseWriter: w,
		empty:          true,
		status:         0,
	}

	// Validate last.
	// Allows custom middleware to be executed after core
	// middleware and before validation.
	handler = validateRequestMiddleware(handler)
	handler = r.applyMiddleware(handler)

	parent := r.parent
	for parent != nil {
		handler = parent.applyMiddleware(handler)
		parent = parent.parent
	}

	handler(response, request)

	r.finalize(response, request)
}

func (r *Router) applyMiddleware(handler Handler) Handler {
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handler = r.middleware[i](handler)
	}
	return handler
}

// finalize the request's life-cycle.
func (r *Router) finalize(response *Response, request *Request) {
	if response.empty {
		if response.status == 0 {
			// If the response is empty, return status 204 to
			// comply with RFC 7231, 6.3.5
			response.Status(http.StatusNoContent)
		} else if statusHandler, ok := r.statusHandlers[response.status]; ok {
			// Status has been set but body is empty.
			// Execute status handler if exists.
			statusHandler(response, request)
		}
	}

	if !response.wroteHeader {
		response.WriteHeader(response.status)
	}
}
