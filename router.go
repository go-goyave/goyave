package goyave

import (
	"errors"
	"io/fs"
	"net/http"
	"regexp"
	"strings"

	"maps"
	"slices"

	"github.com/samber/lo"
	"goyave.dev/goyave/v5/cors"
	errorutil "goyave.dev/goyave/v5/util/errors"
)

// Common route meta keys.
const (
	MetaCORS = "goyave.cors"
)

// Special route names.
const (
	RouteMethodNotAllowed = "goyave.method-not-allowed"
	RouteNotFound         = "goyave.not-found"
)

var (
	errMatchMethodNotAllowed = errors.New("Method not allowed for this route")
	errMatchNotFound         = errors.New("No match for this URI")

	methodNotAllowedRoute = newRoute(func(response *Response, _ *Request) {
		response.Status(http.StatusMethodNotAllowed)
	}, RouteMethodNotAllowed)
	notFoundRoute = newRoute(func(response *Response, _ *Request) {
		response.Status(http.StatusNotFound)
	}, RouteNotFound)
)

// Handler responds to an HTTP request.
//
// The given `Response` and `Request` value should not
// be used outside of the context of an HTTP request. e.g.: passed to
// a goroutine or used after the finalization step in the request lifecycle.
type Handler func(response *Response, request *Request)

type routeMatcher interface {
	match(method string, match *routeMatch) bool
}

type routeMatch struct {
	route       *Route
	parameters  map[string]string
	err         error
	currentPath string
}

func (rm *routeMatch) mergeParams(params map[string]string) {
	if rm.parameters == nil {
		rm.parameters = params
		return
	}
	for k, v := range params {
		rm.parameters[k] = v
	}
}

func (rm *routeMatch) trimCurrentPath(fullMatch string) {
	length := len(fullMatch)
	rm.currentPath = rm.currentPath[length:]
}

// Router registers routes to be matched and executes a handler.
type Router struct {
	server         *Server
	parent         *Router
	statusHandlers map[int]StatusHandler
	namedRoutes    map[string]*Route
	regexCache     map[string]*regexp.Regexp
	Meta           map[string]any

	parameterizable
	middlewareHolder
	globalMiddleware *middlewareHolder

	prefix     string
	routes     []*Route
	subrouters []*Router

	slashCount int
}

var _ http.Handler = (*Router)(nil) // implements http.Handler
var _ routeMatcher = (*Router)(nil) // implements routeMatcher

// NewRouter create a new root-level Router that is pre-configured with core
// middleware (recovery and language), as well as status handlers
// for all standard HTTP status codes.
//
// You don't need to manually build your router using this function.
// This method can however be useful for external tooling that build
// routers without starting the HTTP server. Don't forget to call
// `router.ClearRegexCache()` when you are done registering routes.
func NewRouter(server *Server) *Router {
	router := &Router{
		server:         server,
		parent:         nil,
		prefix:         "",
		statusHandlers: make(map[int]StatusHandler, 41),
		namedRoutes:    make(map[string]*Route, 5),
		middlewareHolder: middlewareHolder{
			middleware: nil,
		},
		globalMiddleware: &middlewareHolder{
			middleware: make([]Middleware, 0, 2),
		},
		regexCache: make(map[string]*regexp.Regexp, 5),
		Meta:       make(map[string]any),
	}
	router.StatusHandler(&PanicStatusHandler{}, http.StatusInternalServerError)
	for i := 400; i <= 418; i++ {
		router.StatusHandler(&ErrorStatusHandler{}, i)
	}
	router.StatusHandler(&ValidationStatusHandler{}, http.StatusUnprocessableEntity)
	for i := 423; i <= 426; i++ {
		router.StatusHandler(&ErrorStatusHandler{}, i)
	}
	router.StatusHandler(&ErrorStatusHandler{}, 421, 428, 429, 431, 444, 451)
	router.StatusHandler(&ErrorStatusHandler{}, 501, 502, 503, 504, 505, 506, 507, 508, 510, 511)
	router.GlobalMiddleware(&recoveryMiddleware{}, &languageMiddleware{})
	return router
}

// ClearRegexCache set internal router's regex cache used for route parameters optimisation to nil
// so it can be garbage collected.
// You don't need to call this function if you are using `goyave.Server`.
// However, this method SHOULD be called by external tooling that build routers without starting the HTTP
// server when they are done registering routes and subrouters.
func (r *Router) ClearRegexCache() {
	r.regexCache = nil
	for _, subrouter := range r.subrouters {
		subrouter.ClearRegexCache()
	}
}

// GetParent returns the parent Router of this router (can be `nil`).
func (r *Router) GetParent() *Router {
	return r.parent
}

// GetRoutes returns the list of routes belonging to this router.
func (r *Router) GetRoutes() []*Route {
	cpy := make([]*Route, len(r.routes))
	copy(cpy, r.routes)
	return cpy
}

// GetSubrouters returns the list of subrouters belonging to this router.
func (r *Router) GetSubrouters() []*Router {
	cpy := make([]*Router, len(r.subrouters))
	copy(cpy, r.subrouters)
	return cpy
}

// GetRoute get a named route.
// Returns nil if the route doesn't exist.
func (r *Router) GetRoute(name string) *Route {
	return r.namedRoutes[name]
}

// SetMeta attach a value to this router identified by the given key.
//
// This value is inherited by all subrouters and routes, unless they override
// it at their level.
func (r *Router) SetMeta(key string, value any) *Router {
	r.Meta[key] = value
	return r
}

// RemoveMeta detach the meta value identified by the given key from this router.
// This doesn't remove meta using the same key from the parent routers.
func (r *Router) RemoveMeta(key string) *Router {
	delete(r.Meta, key)
	return r
}

// LookupMeta value identified by the given key. If not found in this router,
// the value is recursively fetched in the parent routers.
//
// Returns the value and `true` if found in the current router or one of the
// parent routers, `nil` and `false` otherwise.
func (r *Router) LookupMeta(key string) (any, bool) {
	val, ok := r.Meta[key]
	if ok {
		return val, ok
	}
	if r.parent != nil {
		return r.parent.LookupMeta(key)
	}
	return nil, false
}

// GlobalMiddleware apply one or more global middleware. Global middleware are
// executed for every request, including when the request doesn't match any route
// or if it results in "Method Not Allowed".
// These middleware are global to the main Router: they will also be executed for subrouters.
// Global Middleware are always executed first.
// Use global middleware for logging and rate limiting for example.
func (r *Router) GlobalMiddleware(middleware ...Middleware) *Router {
	for _, m := range middleware {
		m.Init(r.server)
	}
	r.globalMiddleware.middleware = append(r.globalMiddleware.middleware, middleware...)
	return r
}

// Middleware apply one or more middleware to the route group.
func (r *Router) Middleware(middleware ...Middleware) *Router {
	if r.middleware == nil {
		r.middleware = make([]Middleware, 0, 3)
	}
	for _, m := range middleware {
		m.Init(r.server)
	}
	r.middleware = append(r.middleware, middleware...)
	return r
}

// CORS set the CORS options for this route group.
// If the options are not `nil`, the CORS middleware is automatically added globally.
// To disable CORS for this router, subrouters and routes, give `nil` options.
// CORS can be re-enabled for subrouters and routes on a case-by-case basis
// using non-nil options.
func (r *Router) CORS(options *cors.Options) *Router {
	r.Meta[MetaCORS] = options
	if options == nil {
		return r
	}
	if !hasMiddleware[*corsMiddleware](r.globalMiddleware.middleware) {
		r.GlobalMiddleware(&corsMiddleware{})
	}
	return r
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
// Codes in the 400 and 500 ranges have a default status handler.
func (r *Router) StatusHandler(handler StatusHandler, status int, additionalStatuses ...int) {
	handler.Init(r.server)
	r.statusHandlers[status] = handler
	for _, s := range additionalStatuses {
		r.statusHandlers[s] = handler
	}
}

// ServeHTTP dispatches the handler registered in the matched route.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Scheme != "" && req.URL.Scheme != "http" {
		address := r.server.getProxyAddress(r.server.config) + req.URL.Path
		query := req.URL.Query()
		if len(query) != 0 {
			address += "?" + query.Encode()
		}
		http.Redirect(w, req, address, http.StatusPermanentRedirect)
		return
	}

	match := routeMatch{currentPath: req.URL.Path}
	r.match(req.Method, &match)
	r.requestHandler(&match, w, req)
}

// TODO export RouteMatch and add Match with string param function

func (r *Router) match(method string, match *routeMatch) bool {
	// Check if router itself matches
	var params []string
	if r.parameterizable.regex != nil {
		i := -1
		if len(match.currentPath) > 0 {
			// Ignore slashes in router prefix
			i = nthIndex(match.currentPath[1:], "/", r.slashCount) + 1
		}
		if i <= 0 {
			i = len(match.currentPath)
		}
		currentPath := match.currentPath[:i]
		params = r.parameterizable.regex.FindStringSubmatch(currentPath)
	} else {
		params = []string{""}
	}

	if params != nil {
		match.trimCurrentPath(params[0])
		if len(params) > 1 {
			match.mergeParams(r.makeParameters(params))
		}

		// Check in subrouters first
		for _, router := range r.subrouters {
			if router.match(method, match) {
				if router.prefix == "" && match.route == methodNotAllowedRoute {
					// This allows route groups with subrouters having empty prefix.
					continue
				}
				return true
			}
		}

		// Check if any route matches
		for _, route := range r.routes {
			if route.match(method, match) {
				return true
			}
		}
	}

	if match.err == errMatchMethodNotAllowed {
		match.route = methodNotAllowedRoute
		return true
	}

	match.route = notFoundRoute
	// Return true if the subrouter matched so we don't turn back and check other subrouters
	return params != nil && len(params[0]) > 0
}

func nthIndex(str, substr string, n int) int {
	index := -1
	for nth := 0; nth < n; nth++ {
		i := strings.Index(str, substr)
		if i == -1 || i == len(str) {
			return -1
		}
		index += i + 1
		str = str[i+1:]
	}
	return index
}

func (r *Router) makeParameters(match []string) map[string]string {
	return r.parameterizable.makeParameters(match, r.parameters)
}

// Subrouter create a new sub-router from this router.
// Use subrouters to create route groups and to apply middleware to multiple routes.
// CORS options are also inherited.
//
// Subrouters are matched before routes. For example, if you have a subrouter with a
// prefix "/{name}" and a route "/route", the "/route" will never match.
func (r *Router) Subrouter(prefix string) *Router {
	if prefix == "/" {
		prefix = ""
	}

	router := &Router{
		server:         r.server,
		parent:         r,
		prefix:         prefix,
		statusHandlers: maps.Clone(r.statusHandlers),
		Meta:           make(map[string]any),
		namedRoutes:    r.namedRoutes,
		routes:         make([]*Route, 0, 5), // Typical CRUD has 5 routes
		middlewareHolder: middlewareHolder{
			middleware: nil,
		},
		globalMiddleware: r.globalMiddleware,
		regexCache:       r.regexCache,
	}
	if prefix != "" {
		router.compileParameters(router.prefix, false, r.regexCache)
		router.slashCount = strings.Count(prefix, "/")
	}
	r.subrouters = append(r.subrouters, router)
	return router
}

// Group create a new sub-router with an empty prefix.
func (r *Router) Group() *Router {
	return r.Subrouter("")
}

// Route register a new route.
//
// Multiple methods can be passed.
//
// If the route matches the "GET" method, the "HEAD" method is automatically added
// to the matcher if it's missing.
//
// If the router has the CORS middleware, the "OPTIONS" method is automatically added
// to the matcher if it's missing, so it allows preflight requests.
//
// Returns the generated route.
func (r *Router) Route(methods []string, uri string, handler Handler) *Route {
	return r.registerRoute(methods, uri, handler)
}

// Get registers a new route with the GET and HEAD methods.
func (r *Router) Get(uri string, handler Handler) *Route {
	return r.registerRoute([]string{http.MethodGet}, uri, handler)
}

// Post registers a new route with the POST method.
func (r *Router) Post(uri string, handler Handler) *Route {
	return r.registerRoute([]string{http.MethodPost}, uri, handler)
}

// Put registers a new route with the PUT method.
func (r *Router) Put(uri string, handler Handler) *Route {
	return r.registerRoute([]string{http.MethodPut}, uri, handler)
}

// Patch registers a new route with the PATCH method.
func (r *Router) Patch(uri string, handler Handler) *Route {
	return r.registerRoute([]string{http.MethodPatch}, uri, handler)
}

// Delete registers a new route with the DELETE method.
func (r *Router) Delete(uri string, handler Handler) *Route {
	return r.registerRoute([]string{http.MethodDelete}, uri, handler)
}

// Options registers a new route wit the OPTIONS method.
func (r *Router) Options(uri string, handler Handler) *Route {
	return r.registerRoute([]string{http.MethodOptions}, uri, handler)
}

// Static serve a directory and its subdirectories of static resources.
// Set the "download" parameter to true if you want the files to be sent as an attachment
// instead of an inline element.
//
// If no file is given in the url, or if the given file is a directory, the handler will
// send the "index.html" file if it exists.
func (r *Router) Static(fs fs.StatFS, uri string, download bool) *Route {
	return r.registerRoute([]string{http.MethodGet}, uri+"{resource:.*}", staticHandler(fs, download))
}

func (r *Router) registerRoute(methods []string, uri string, handler Handler) *Route {
	methodsSlice := slices.Clone(methods)

	corsOptions, hasCORSOptions := r.LookupMeta(MetaCORS)
	if hasCORSOptions && corsOptions != (*cors.Options)(nil) && !lo.Contains(methodsSlice, http.MethodOptions) {
		methodsSlice = append(methodsSlice, http.MethodOptions)
	}

	if lo.Contains(methodsSlice, http.MethodGet) && !lo.Contains(methodsSlice, http.MethodHead) {
		methodsSlice = append(methodsSlice, http.MethodHead)
	}

	if uri == "/" && r.parent != nil && !(r.prefix == "" && r.parent.parent == nil) {
		uri = ""
	}

	route := &Route{
		name:    "",
		uri:     uri,
		methods: methodsSlice,
		parent:  r,
		handler: handler,
		Meta:    make(map[string]any),
	}
	route.compileParameters(route.uri, true, r.regexCache)
	r.routes = append(r.routes, route)
	return route
}

// Controller register all routes for a controller implementing the `Registrer` interface.
// Automatically calls `Init()` and `RegisterRoutes()` on the given controller.
func (r *Router) Controller(controller Registrer) *Router {
	controller.Init(r.server)
	controller.RegisterRoutes(r)
	return r
}

func (r *Router) requestHandler(match *routeMatch, w http.ResponseWriter, rawRequest *http.Request) {
	request := NewRequest(rawRequest)
	request.Route = match.route
	if match.parameters == nil {
		request.RouteParams = map[string]string{}
	} else {
		request.RouteParams = match.parameters
	}
	response := NewResponse(r.server, request, w)
	handler := match.route.handler

	// Route-specific middleware is executed after router middleware
	handler = match.route.applyMiddleware(handler)

	parent := match.route.parent
	for parent != nil {
		handler = parent.applyMiddleware(handler)
		parent = parent.parent
	}

	handler = r.globalMiddleware.applyMiddleware(handler)

	handler(response, request)

	if err := r.finalize(response, request); err != nil {
		r.server.Logger.Error(err)
	}

	requestPool.Put(request)
	responsePool.Put(response)
}

// finalize the request's life-cycle.
func (r *Router) finalize(response *Response, request *Request) error {
	if response.empty {
		if response.status == 0 {
			// If the response is empty, return status 204 to
			// comply with RFC 7231, 6.3.5
			response.Status(http.StatusNoContent)
		} else if statusHandler, ok := r.statusHandlers[response.status]; ok {
			// Status has been set but body is empty.
			// Execute status handler if exists.
			statusHandler.Handle(response, request)
		}
	}

	if !response.wroteHeader && !response.hijacked {
		response.WriteHeader(response.status)
	}

	return errorutil.New(response.close())
}
