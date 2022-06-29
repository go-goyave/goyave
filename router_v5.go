package goyave

import (
	"net/http"
	"regexp"
	"strings"

	"goyave.dev/goyave/v4/util/maputil"
)

const (
	MetaValidationRules = "goyave.validationRules"
	MetaCORS            = "goyave.cors"
)

var (
	methodNotAllowedRouteV5 = newRouteV5(func(c *Context) {
		c.Status(http.StatusMethodNotAllowed)
	})
	notFoundRouteV5 = newRouteV5(func(c *Context) {
		c.Status(http.StatusNotFound)
	})
)

type HandlerV5 func(*Context)

type routeMatcherV5 interface {
	match(req *http.Request, match *routeMatchV5) bool
}

type routeMatchV5 struct {
	route       *RouteV5
	parameters  map[string]string
	err         error
	currentPath string
}

func (rm *routeMatchV5) mergeParams(params map[string]string) {
	if rm.parameters == nil {
		rm.parameters = params
	}
	for k, v := range params {
		rm.parameters[k] = v
	}
}

func (rm *routeMatchV5) trimCurrentPath(fullMatch string) {
	rm.currentPath = rm.currentPath[len(fullMatch):]
}

// Router registers routes to be matched and executes a handler.
type RouterV5 struct {
	server         *Server
	parent         *RouterV5
	statusHandlers map[int]HandlerV5
	namedRoutes    map[string]*RouteV5
	regexCache     map[string]*regexp.Regexp
	Meta           map[string]any

	parameterizable
	middlewareHolderV5
	globalMiddleware *middlewareHolderV5

	prefix     string
	routes     []*RouteV5
	subrouters []*RouterV5
}

var _ http.Handler = (*RouterV5)(nil)   // implements http.Handler
var _ routeMatcherV5 = (*RouterV5)(nil) // implements routeMatcher

// NewRouter create a new root-level Router that is pre-configured with core
// middleware (recovery and language), as well as status handlers
// for all standard HTTP status codes.
//
// You don't need to manually build your router using this function.
// This method can however be useful for external tooling that build
// routers without starting the HTTP server. Don't forget to call
// `router.ClearRegexCache()` when you are done registering routes.
func NewRouterV5(server *Server) *RouterV5 {
	router := &RouterV5{
		server:         server,
		parent:         nil,
		prefix:         "",
		statusHandlers: make(map[int]HandlerV5, 41),
		namedRoutes:    make(map[string]*RouteV5, 5),
		middlewareHolderV5: middlewareHolderV5{
			middleware: nil,
		},
		globalMiddleware: &middlewareHolderV5{
			middleware: make([]MiddlewareV5, 0, 2),
		},
		regexCache: make(map[string]*regexp.Regexp, 5),
		Meta:       make(map[string]any),
	}
	router.StatusHandler(PanicStatusHandlerV5, http.StatusInternalServerError)
	router.StatusHandler(ValidationStatusHandlerV5, http.StatusBadRequest, http.StatusUnprocessableEntity)
	for i := 401; i <= 418; i++ {
		router.StatusHandler(ErrorStatusHandlerV5, i)
	}
	for i := 423; i <= 426; i++ {
		router.StatusHandler(ErrorStatusHandlerV5, i)
	}
	router.StatusHandler(ErrorStatusHandlerV5, 421, 428, 429, 431, 444, 451)
	router.StatusHandler(ErrorStatusHandlerV5, 501, 502, 503, 504, 505, 506, 507, 508, 510, 511)
	router.GlobalMiddleware(recoveryMiddlewareV5, languageMiddlewareV5)
	return router
}

// ClearRegexCache set internal router's regex cache used for route parameters optimisation to nil
// so it can be garbage collected.
// You don't need to call this function if you are using `goyave.Server`.
// However, this method SHOULD be called by external tooling that build routers without starting the HTTP
// server when they are done registering routes and subrouters.
func (r *RouterV5) ClearRegexCache() {
	r.regexCache = nil
	for _, subrouter := range r.subrouters {
		subrouter.ClearRegexCache()
	}
}

// GetRoutes returns the list of routes belonging to this router.
func (r *RouterV5) GetRoutes() []*RouteV5 {
	cpy := make([]*RouteV5, len(r.routes))
	copy(cpy, r.routes)
	return cpy
}

// GetSubrouters returns the list of subrouters belonging to this router.
func (r *RouterV5) GetSubrouters() []*RouterV5 {
	cpy := make([]*RouterV5, len(r.subrouters))
	copy(cpy, r.subrouters)
	return cpy
}

// GetRoute get a named route.
// Returns nil if the route doesn't exist.
func (r *RouterV5) GetRoute(name string) *RouteV5 {
	return r.namedRoutes[name]
}

func (r *RouterV5) SetMeta(key string, value any) *RouterV5 {
	r.Meta[key] = value
	return r
}

func (r *RouterV5) RemoveMeta(key string) *RouterV5 {
	delete(r.Meta, key)
	return r
}

// GlobalMiddleware apply one or more global middleware. Global middleware are
// executed for every request, including when the request doesn't match any route
// or if it results in "Method Not Allowed".
// These middleware are global to the main Router: they will also be executed for subrouters.
// Global Middleware are always executed first.
// Use global middleware for logging and rate limiting for example.
func (r *RouterV5) GlobalMiddleware(middleware ...MiddlewareV5) *RouterV5 { // TODO middleware signature will be changed automatically when Handler signature is changed
	r.globalMiddleware.middleware = append(r.globalMiddleware.middleware, middleware...)
	return r
}

// Middleware apply one or more middleware to the route group.
func (r *RouterV5) Middleware(middleware ...MiddlewareV5) *RouterV5 {
	if r.middleware == nil {
		r.middleware = make([]MiddlewareV5, 0, 3)
	}
	r.middleware = append(r.middleware, middleware...)
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
func (r *RouterV5) StatusHandler(handler HandlerV5, status int, additionalStatuses ...int) {
	r.statusHandlers[status] = handler
	for _, s := range additionalStatuses {
		r.statusHandlers[s] = handler
	}
}

// ServeHTTP dispatches the handler registered in the matched route.
func (r *RouterV5) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Scheme != "" && req.URL.Scheme != protocol {
		address := getAddress(protocol) + req.URL.Path
		query := req.URL.Query()
		if len(query) != 0 {
			address += "?" + query.Encode()
		}
		http.Redirect(w, req, address, http.StatusPermanentRedirect)
		return
	}

	match := routeMatchV5{currentPath: req.URL.Path}
	r.match(req, &match)
	r.requestHandler(&match, w, req)
}

func (r *RouterV5) match(req *http.Request, match *routeMatchV5) bool {
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
				if router.prefix == "" && match.route == methodNotAllowedRouteV5 {
					// This allows route groups with subrouters having empty prefix.
					break
				}
				return true
			}
		}

		// Check if any route matches
		for _, route := range r.routes {
			if route.match(req, match) {
				return true
			}
		}
	}

	if match.err == errMatchMethodNotAllowed {
		match.route = methodNotAllowedRouteV5
		return true
	}

	match.route = notFoundRouteV5
	return false
}

func (r *RouterV5) makeParameters(match []string) map[string]string {
	return r.parameterizable.makeParameters(match, r.parameters)
}

// Subrouter create a new sub-router from this router.
// Use subrouters to create route groups and to apply middleware to multiple routes.
// CORS options are also inherited.
func (r *RouterV5) Subrouter(prefix string) *RouterV5 {
	if prefix == "/" {
		prefix = ""
	}

	router := &RouterV5{
		server:         r.server,
		parent:         r,
		prefix:         prefix,
		statusHandlers: maputil.Clone(r.statusHandlers),
		Meta:           maputil.Clone(r.Meta),
		namedRoutes:    r.namedRoutes,
		routes:         make([]*RouteV5, 0, 5), // Typical CRUD has 5 routes
		middlewareHolderV5: middlewareHolderV5{
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
func (r *RouterV5) Group() *RouterV5 {
	return r.Subrouter("")
}

// Route register a new route.
//
// Multiple methods can be passed using a pipe-separated string.
//  "PUT|PATCH"
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
func (r *RouterV5) Route(methods string, uri string, handler HandlerV5) *RouteV5 {
	return r.registerRoute(methods, uri, handler)
}

// Get registers a new route with the GET and HEAD methods.
func (r *RouterV5) Get(uri string, handler HandlerV5) *RouteV5 {
	return r.registerRoute(http.MethodGet, uri, handler)
}

// Post registers a new route with the POST method.
func (r *RouterV5) Post(uri string, handler HandlerV5) *RouteV5 {
	return r.registerRoute(http.MethodPost, uri, handler)
}

// Put registers a new route with the PUT method.
func (r *RouterV5) Put(uri string, handler HandlerV5) *RouteV5 {
	return r.registerRoute(http.MethodPut, uri, handler)
}

// Patch registers a new route with the PATCH method.
func (r *RouterV5) Patch(uri string, handler HandlerV5) *RouteV5 {
	return r.registerRoute(http.MethodPatch, uri, handler)
}

// Delete registers a new route with the DELETE method.
func (r *RouterV5) Delete(uri string, handler HandlerV5) *RouteV5 {
	return r.registerRoute(http.MethodDelete, uri, handler)
}

// Options registers a new route wit the OPTIONS method.
func (r *RouterV5) Options(uri string, handler HandlerV5) *RouteV5 {
	return r.registerRoute(http.MethodOptions, uri, handler)
}

func (r *RouterV5) registerRoute(methods string, uri string, handler HandlerV5) *RouteV5 {
	// if r.corsOptions != nil && !strings.Contains(methods, "OPTIONS") {
	// 	methods += "|OPTIONS"
	// }

	if strings.Contains(methods, http.MethodGet) && !strings.Contains(methods, http.MethodHead) {
		methods += "|HEAD"
	}

	if uri == "/" && r.parent != nil && !(r.prefix == "" && r.parent.parent == nil) {
		uri = ""
	}

	route := &RouteV5{
		name:    "",
		uri:     uri,
		methods: strings.Split(methods, "|"),
		parent:  r,
		handler: handler,
		Meta:    make(map[string]any),
	}
	route.compileParameters(route.uri, true, r.regexCache)
	r.routes = append(r.routes, route)
	return route
}

func (r *RouterV5) requestHandler(match *routeMatchV5, w http.ResponseWriter, rawRequest *http.Request) {
	ctx := &Context{
		server:     r.server,
		route:      match.route,
		RequestV5:  newRequest(rawRequest),
		ResponseV5: newResponseV5(w, rawRequest),
		Extra:      make(map[string]any),
	}
	handler := match.route.handler

	// Validate last.
	// Allows custom middleware to be executed after core
	// middleware and before validation.
	// handler = validateRequestMiddleware(handler)
	// TODO re-enable validation middleware

	// Route-specific middleware is executed after router middleware
	handler = match.route.applyMiddleware(handler)

	parent := match.route.parent
	for parent != nil {
		handler = parent.applyMiddleware(handler)
		parent = parent.parent
	}

	handler = r.globalMiddleware.applyMiddleware(handler)

	handler(ctx)

	if err := r.finalize(ctx); err != nil {
		r.server.ErrLogger.Println(err)
	}
}

// finalize the request's life-cycle.
func (r *RouterV5) finalize(ctx *Context) error {
	if ctx.ResponseV5.empty {
		if ctx.ResponseV5.status == 0 {
			// If the response is empty, return status 204 to
			// comply with RFC 7231, 6.3.5
			ctx.Status(http.StatusNoContent)
		} else if statusHandler, ok := r.statusHandlers[ctx.ResponseV5.status]; ok {
			// Status has been set but body is empty.
			// Execute status handler if exists.
			statusHandler(ctx)
		}
	}

	if !ctx.ResponseV5.wroteHeader && !ctx.ResponseV5.hijacked {
		ctx.WriteHeader(ctx.ResponseV5.status)
	}

	return ctx.ResponseV5.close()
}
