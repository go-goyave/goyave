package goyave

import (
	"errors"
	"net/http"
	"os"
	"regexp"
	"strings"

	"goyave.dev/goyave/v4/cors"
	"goyave.dev/goyave/v4/util/fsutil"
)

type routeMatcher interface {
	match(req *http.Request, match *routeMatch) bool
}

// Router registers routes to be matched and executes a handler.
type Router struct {
	parent         *Router
	corsOptions    *cors.Options
	statusHandlers map[int]Handler
	namedRoutes    map[string]*Route
	regexCache     map[string]*regexp.Regexp

	parameterizable
	middlewareHolder
	globalMiddleware *middlewareHolder

	prefix            string
	routes            []*Route
	subrouters        []*Router
	hasCORSMiddleware bool
}

var _ http.Handler = (*Router)(nil) // implements http.Handler
var _ routeMatcher = (*Router)(nil) // implements routeMatcher

// Handler is a controller or middleware function
type Handler func(*Response, *Request)

type middlewareHolder struct {
	middleware []Middleware
}

type routeMatch struct {
	route       *Route
	corsOptions *cors.Options
	parameters  map[string]string
	err         error
	currentPath string
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

// PanicStatusHandler for the HTTP 500 error.
// If debugging is enabled, writes the error details to the response and
// print stacktrace in the console.
// If debugging is not enabled, writes `{"error": "Internal Server Error"}`
// to the response.
func PanicStatusHandler(response *Response, _ *Request) {
	response.error(response.GetError())
	if response.empty {
		message := map[string]string{
			"error": http.StatusText(response.GetStatus()),
		}
		response.JSON(response.GetStatus(), message)
	}
}

// ErrorStatusHandler a generic status handler for non-success codes.
// Writes the corresponding status message to the response.
func ErrorStatusHandler(response *Response, _ *Request) {
	message := map[string]string{
		"error": http.StatusText(response.GetStatus()),
	}
	response.JSON(response.GetStatus(), message)
}

// ValidationStatusHandler for HTTP 400 and HTTP 422 errors.
// Writes the validation errors to the response.
func ValidationStatusHandler(response *Response, _ *Request) {
	message := map[string]interface{}{"validationError": response.GetError()}
	response.JSON(response.GetStatus(), message)
}

// NewRouter create a new root-level Router that is pre-configured with core
// middleware (recovery, parse and language), as well as status handlers
// for all standard HTTP status codes.
//
// You don't need to manually build your router using this function
// if you are using `goyave.Start()`. This method can however be useful for external
// tooling that build routers without starting the HTTP server. Don't forget to call
// router.ClearRegexCache() when you are done registering routes.
func NewRouter() *Router {
	router := &Router{
		parent:            nil,
		prefix:            "",
		hasCORSMiddleware: false,
		statusHandlers:    make(map[int]Handler, 41),
		namedRoutes:       make(map[string]*Route, 5),
		middlewareHolder: middlewareHolder{
			middleware: nil,
		},
		globalMiddleware: &middlewareHolder{
			middleware: make([]Middleware, 0, 3),
		},
		regexCache: make(map[string]*regexp.Regexp, 5),
	}
	router.StatusHandler(PanicStatusHandler, http.StatusInternalServerError)
	router.StatusHandler(ValidationStatusHandler, http.StatusBadRequest, http.StatusUnprocessableEntity)
	for i := 401; i <= 418; i++ {
		router.StatusHandler(ErrorStatusHandler, i)
	}
	for i := 423; i <= 426; i++ {
		router.StatusHandler(ErrorStatusHandler, i)
	}
	router.StatusHandler(ErrorStatusHandler, 421, 428, 429, 431, 444, 451)
	router.StatusHandler(ErrorStatusHandler, 501, 502, 503, 504, 505, 506, 507, 508, 510, 511)
	router.GlobalMiddleware(recoveryMiddleware, parseRequestMiddleware, languageMiddleware)
	return router
}

// ClearRegexCache set internal router's regex cache used for route parameters optimisation to nil
// so it can be garbage collected.
// You don't need to call this function if you are using `goyave.Start()`.
// However, this method SHOULD be called by external tooling that build routers without starting the HTTP
// server when they are done registering routes and subrouters.
func (r *Router) ClearRegexCache() {
	r.regexCache = nil
	for _, subrouter := range r.subrouters {
		subrouter.ClearRegexCache()
	}
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

// ServeHTTP dispatches the handler registered in the matched route.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Scheme != "" && req.URL.Scheme != protocol {
		address := getAddress(protocol) + req.URL.Path
		query := req.URL.Query()
		if len(query) != 0 {
			address += "?" + query.Encode()
		}
		http.Redirect(w, req, address, http.StatusPermanentRedirect)
		return
	}

	match := routeMatch{currentPath: req.URL.Path}
	r.match(req, &match)
	r.requestHandler(&match, w, req)
}

func (r *Router) match(req *http.Request, match *routeMatch) bool {
	// Check if router itself matches
	var params []string
	if r.parameterizable.regex != nil {
		params = r.parameterizable.regex.FindStringSubmatch(match.currentPath)
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
			if router.match(req, match) {
				if router.prefix == "" && match.route == methodNotAllowedRoute {
					// This allows route groups with subrouters having empty prefix.
					break
				}
				return true
			}
		}

		// Check if any route matches
		for _, route := range r.routes {
			if route.match(req, match) {
				match.corsOptions = r.corsOptions
				return true
			}
		}
	}

	match.corsOptions = r.corsOptions
	if match.err == errMatchMethodNotAllowed {
		match.route = methodNotAllowedRoute
		return true
	}

	match.route = notFoundRoute
	return false
}

func (r *Router) makeParameters(match []string) map[string]string {
	return r.parameterizable.makeParameters(match, r.parameters)
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
		prefix:            prefix,
		corsOptions:       r.corsOptions,
		hasCORSMiddleware: r.hasCORSMiddleware,
		statusHandlers:    r.copyStatusHandlers(),
		namedRoutes:       r.namedRoutes,
		routes:            make([]*Route, 0, 5), // Typical CRUD has 5 routes
		middlewareHolder: middlewareHolder{
			middleware: nil,
		},
		globalMiddleware: r.globalMiddleware,
		regexCache:       r.regexCache,
	}
	router.compileParameters(router.prefix, false, r.regexCache)
	r.subrouters = append(r.subrouters, router)
	return router
}

// Group create a new sub-router with an empty prefix.
func (r *Router) Group() *Router {
	return r.Subrouter("")
}

// GlobalMiddleware apply one or more global middleware. Global middleware are
// executed for every request, including when the request doesn't match any route
// or if it results in "Method Not Allowed".
// These middleware are global to the main Router: they will also be executed for subrouters.
// Global Middleware are always executed first.
// Use global middleware for logging and rate limiting for example.
func (r *Router) GlobalMiddleware(middleware ...Middleware) {
	r.globalMiddleware.middleware = append(r.globalMiddleware.middleware, middleware...)
}

// Middleware apply one or more middleware to the route group.
func (r *Router) Middleware(middleware ...Middleware) {
	if r.middleware == nil {
		r.middleware = make([]Middleware, 0, 3)
	}
	r.middleware = append(r.middleware, middleware...)
}

// Route register a new route.
//
// Multiple methods can be passed using a pipe-separated string.
//
//	"PUT|PATCH"
//
// The validation rules set is optional. If you don't want your route
// to be validated, pass "nil".
//
// If the route matches the "GET" method, the "HEAD" method is automatically added
// to the matcher if it's missing.
//
// If the router has CORS options set, the "OPTIONS" method is automatically added
// to the matcher if it's missing, so it allows preflight requests.
//
// Returns the generated route.
func (r *Router) Route(methods string, uri string, handler Handler) *Route {
	return r.registerRoute(methods, uri, handler)
}

func (r *Router) registerRoute(methods string, uri string, handler Handler) *Route {
	if r.corsOptions != nil && !strings.Contains(methods, "OPTIONS") {
		methods += "|OPTIONS"
	}

	if strings.Contains(methods, "GET") && !strings.Contains(methods, "HEAD") {
		methods += "|HEAD"
	}

	if uri == "/" && r.parent != nil && !(r.prefix == "" && r.parent.parent == nil) {
		uri = ""
	}

	route := &Route{
		name:    "",
		uri:     uri,
		methods: strings.Split(methods, "|"),
		parent:  r,
		handler: handler,
	}
	route.compileParameters(route.uri, true, r.regexCache)
	r.routes = append(r.routes, route)
	return route
}

// Get registers a new route with the GET and HEAD methods.
func (r *Router) Get(uri string, handler Handler) *Route {
	return r.registerRoute(http.MethodGet, uri, handler)
}

// Post registers a new route with the POST method.
func (r *Router) Post(uri string, handler Handler) *Route {
	return r.registerRoute(http.MethodPost, uri, handler)
}

// Put registers a new route with the PUT method.
func (r *Router) Put(uri string, handler Handler) *Route {
	return r.registerRoute(http.MethodPut, uri, handler)
}

// Patch registers a new route with the PATCH method.
func (r *Router) Patch(uri string, handler Handler) *Route {
	return r.registerRoute(http.MethodPatch, uri, handler)
}

// Delete registers a new route with the DELETE method.
func (r *Router) Delete(uri string, handler Handler) *Route {
	return r.registerRoute(http.MethodDelete, uri, handler)
}

// Options registers a new route wit the OPTIONS method.
func (r *Router) Options(uri string, handler Handler) *Route {
	return r.registerRoute(http.MethodOptions, uri, handler)
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
func (r *Router) Static(uri string, directory string, download bool, middleware ...Middleware) {
	r.registerRoute(http.MethodGet, uri+"{resource:.*}", staticHandler(directory, download)).Middleware(middleware...)
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
// Codes in the 400 and 500 ranges have a default status handler.
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

		var err error
		if download {
			err = response.Download(path, file[strings.LastIndex(file, "/")+1:])
		} else {
			err = response.File(path)
		}

		if _, ok := err.(*os.PathError); err != nil && !ok {
			ErrLogger.Println(err)
		}
	}
}

func cleanStaticPath(directory string, file string) string {
	file = strings.TrimPrefix(file, "/")
	path := directory + "/" + file
	if fsutil.IsDirectory(path) {
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
		corsOptions: match.corsOptions,
		Rules:       match.route.validationRules,
		Params:      match.parameters,
		Extra:       map[string]interface{}{},
	}
	response := newResponse(w, rawRequest)
	handler := match.route.handler

	// Validate last.
	// Allows custom middleware to be executed after core
	// middleware and before validation.
	handler = validateRequestMiddleware(handler)

	// Route-specific middleware is executed after router middleware
	handler = match.route.applyMiddleware(handler)

	parent := match.route.parent
	for parent != nil {
		handler = parent.applyMiddleware(handler)
		parent = parent.parent
	}

	handler = r.globalMiddleware.applyMiddleware(handler)

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

	if !response.wroteHeader && !response.hijacked {
		response.WriteHeader(response.status)
	}

	response.close()
}

func (h *middlewareHolder) applyMiddleware(handler Handler) Handler {
	for i := len(h.middleware) - 1; i >= 0; i-- {
		handler = h.middleware[i](handler)
	}
	return handler
}

func (rm *routeMatch) mergeParams(params map[string]string) {
	if rm.parameters == nil {
		rm.parameters = params
	}
	for k, v := range params {
		rm.parameters[k] = v
	}
}

func (rm *routeMatch) trimCurrentPath(fullMatch string) {
	rm.currentPath = rm.currentPath[len(fullMatch):]
}
