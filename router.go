package goyave

import (
	"errors"
	"net/http"
	"strings"

	"github.com/System-Glitch/goyave/v2/cors"
	"github.com/System-Glitch/goyave/v2/helper/filesystem"
	"github.com/System-Glitch/goyave/v2/validation"
)

type routeMatcher interface {
	match(req *http.Request, match *routeMatch) bool
}

// Router registers routes to be matched and executes a handler.
type Router struct {
	parent            *Router
	prefix            string
	corsOptions       *cors.Options
	hasCORSMiddleware bool

	routes         []*Route
	subrouters     []*Router // not sure needed, maybe consider subrouter as a route
	statusHandlers map[int]Handler
	namedRoutes    map[string]*Route
	middlewareHolder
	parametrizeable
}

var _ http.Handler = (*Router)(nil) // implements http.Handler
var _ routeMatcher = (*Router)(nil) // implements routeMatcher

// Handler is a controller or middleware function
type Handler func(*Response, *Request)

type middlewareHolder struct {
	middleware []Middleware
}

type routeMatch struct {
	route      *Route
	err        error
	parameters map[string]string
}

var (
	errMatchMethodNotAllowed = errors.New("Method not allowed for this route")
	errMatchNotFound         = errors.New("No match for this URI")

	methodNotAllowedRoute = newRoute(func(response *Response, request *Request) {
		response.Status(http.StatusMethodNotAllowed)
	})
	notFoundRoute = newRoute(func(response *Response, request *Request) {
		response.Status(http.StatusNotFound)
	})
)

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

func newRouter() *Router {
	// TODO match scheme (protocol)
	router := &Router{
		parent:            nil,
		prefix:            "",
		hasCORSMiddleware: false,
		statusHandlers:    make(map[int]Handler, 15),
		namedRoutes:       make(map[string]*Route, 5),
		middlewareHolder: middlewareHolder{
			middleware: make([]Middleware, 0, 3),
		},
	}
	router.StatusHandler(panicStatusHandler, http.StatusInternalServerError)
	router.StatusHandler(errorStatusHandler, 401, 403, 404, 405, 501, 502, 503, 504, 505, 506, 507, 508, 510, 511)
	router.Middleware(recoveryMiddleware, parseRequestMiddleware, languageMiddleware)
	return router
}

// ServeHTTP dispatches the handler registered in the matched route.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var match routeMatch
	r.match(req, &match)
	r.requestHandler(&match, w, req)
}

func (r *Router) match(req *http.Request, match *routeMatch) bool { // TODO test match
	for _, route := range r.routes {
		if route.match(req, match) {
			return true
		}
	}

	// No route found, check in subrouters
	for _, router := range r.subrouters {
		if router.match(req, match) {
			return true
		}
	}

	if match.err == errMatchMethodNotAllowed {
		match.route = methodNotAllowedRoute
		return true
	}

	match.route = notFoundRoute
	return false
}

// Subrouter create a new sub-router from this router.
// Use subrouters to create route groups and to apply middleware to multiple routes.
// CORS options are also inherited.
func (r *Router) Subrouter(prefix string) *Router {
	if prefix == "/" {
		prefix = ""
	}

	router := &Router{
		parent:            r,
		prefix:            r.prefix + prefix,
		corsOptions:       r.corsOptions,
		hasCORSMiddleware: r.hasCORSMiddleware,
		statusHandlers:    r.copyStatusHandlers(),
		namedRoutes:       r.namedRoutes,
		middlewareHolder: middlewareHolder{
			middleware: make([]Middleware, 0, 3),
		},
	}
	router.compileParameters(router.prefix)
	r.subrouters = append(r.subrouters, router)
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
//
// Returns the generated route.
func (r *Router) Route(methods string, uri string, handler Handler, validationRules validation.RuleSet, middleware ...Middleware) *Route {
	return r.registerRoute(methods, uri, handler, validationRules, middleware...)
}

func (r *Router) registerRoute(methods string, uri string, handler Handler, validationRules validation.RuleSet, middleware ...Middleware) *Route {
	if r.corsOptions != nil && !strings.Contains(methods, "OPTIONS") {
		methods += "|OPTIONS"
	}

	if uri == "/" {
		uri = ""
	}

	route := &Route{
		name:            "",
		uri:             r.prefix + uri, // TODO use partial route only for optimization
		methods:         strings.Split(methods, "|"),
		parent:          r,
		handler:         handler,
		validationRules: validationRules,
		middlewareHolder: middlewareHolder{
			middleware: middleware,
		},
	}
	route.compileParameters(route.uri)
	r.routes = append(r.routes, route)
	return route
}

// GetRoute get a named route.
// Returns nil if the route doesn't exist.
func (r *Router) GetRoute(name string) *Route {
	return r.namedRoutes[name]
}

// Static serve a directory and its subdirectories of static resources.
// Set the "download" parameter to true if you want the files to be sent as an attachment
// instead of an inline element.
//
// If no file is given in the url, or if the given file is a directory, the handler will
// send the "index.html" file if it exists.
func (r *Router) Static(uri string, directory string, download bool) {
	r.registerRoute("GET", uri+"{resource:.*}", staticHandler(directory, download), nil)
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
// Codes in the 500 range and codes 401, 403, 404 and 405 have a default status handler.
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

func (r *Router) copyStatusHandlers() map[int]Handler {
	cpy := make(map[int]Handler, len(r.statusHandlers))
	for key, value := range r.statusHandlers {
		cpy[key] = value
	}
	return cpy
}

func (r *Router) requestHandler(match *routeMatch, w http.ResponseWriter, rawRequest *http.Request) {
	request := &Request{
		httpRequest: rawRequest,
		route:       match.route,
		corsOptions: r.corsOptions,
		Rules:       match.route.validationRules,
		Params:      match.parameters,
	}
	response := &Response{
		httpRequest:    rawRequest,
		ResponseWriter: w,
		empty:          true,
		status:         0,
	}

	handler := match.route.handler

	// Validate last.
	// Allows custom middleware to be executed after core
	// middleware and before validation.
	handler = validateRequestMiddleware(handler)

	// Route-specific middleware is executed after router middleware
	handler = match.route.applyMiddleware(handler)
	handler = r.applyMiddleware(handler)

	parent := r.parent
	for parent != nil {
		handler = parent.applyMiddleware(handler)
		parent = parent.parent
	}

	handler(response, request)

	r.finalize(response, request)
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

func (h *middlewareHolder) applyMiddleware(handler Handler) Handler {
	for i := len(h.middleware) - 1; i >= 0; i-- {
		handler = h.middleware[i](handler)
	}
	return handler
}
