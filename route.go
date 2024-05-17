package goyave

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/samber/lo"
	"goyave.dev/goyave/v5/cors"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/validation"
)

// Route stores information for route matching and serving and can be
// used to generate dynamic URLs/URIs. Routes can, just like routers,
// hold Meta information that can be used by generic middleware to
// alter their behavior depending on the route being served.
type Route struct {
	name    string
	uri     string
	methods []string
	parent  *Router
	Meta    map[string]any
	handler Handler
	middlewareHolder
	parameterizable
}

var _ routeMatcher = (*Route)(nil) // implements routeMatcher

// RuleSetFunc function generating a new validation rule set.
// This function is called for every validated request.
// The returned value is expected to be fresh, not re-used across
// multiple requests nor concurrently.
type RuleSetFunc func(*Request) validation.RuleSet

// newRoute create a new route without any settings except its handler.
// This is used to generate a fake route for the Method Not Allowed and Not Found handlers.
// This route has the core middleware enabled and can be used without a parent router.
// Thus, custom status handlers can use language and body.
func newRoute(handler Handler, name string) *Route {
	return &Route{
		name:    name,
		handler: handler,
		Meta:    make(map[string]any),
		middlewareHolder: middlewareHolder{
			middleware: nil,
		},
	}
}

func (r *Route) match(method string, match *routeMatch) bool {
	if params := r.parameterizable.regex.FindStringSubmatch(match.currentPath); params != nil {
		if r.checkMethod(method) {
			if len(params) > 1 {
				match.mergeParams(r.makeParameters(params))
			}
			match.route = r
			return true
		}
		match.err = errMatchMethodNotAllowed
		return false
	}

	if match.err == nil {
		// Don't override error if already set.
		// Not nil error means it's either already errMatchNotFound
		// or it's errMatchMethodNotAllowed, implying that a route has
		// already been matched but with wrong method.
		match.err = errMatchNotFound
	}
	return false
}

func (r *Route) checkMethod(method string) bool {
	for _, m := range r.methods {
		if m == method {
			return true
		}
	}
	return false
}

func (r *Route) makeParameters(match []string) map[string]string {
	return r.parameterizable.makeParameters(match, r.parameters)
}

// Name set the name of the route.
// Panics if a route with the same name already exists.
// Returns itself.
func (r *Route) Name(name string) *Route {
	if r.name != "" {
		panic(errors.NewSkip("route name is already set", 3))
	}

	if _, ok := r.parent.namedRoutes[name]; ok {
		panic(errors.NewSkip(fmt.Errorf("route %q already exists", name), 3))
	}

	r.name = name
	r.parent.namedRoutes[name] = r
	return r
}

// SetMeta attach a value to this route identified by the given key.
//
// This value can override a value inherited by the parent routers for this route only.
func (r *Route) SetMeta(key string, value any) *Route {
	r.Meta[key] = value
	return r
}

// RemoveMeta detach the meta value identified by the given key from this route.
// This doesn't remove meta using the same key from the parent routers.
func (r *Route) RemoveMeta(key string) *Route {
	delete(r.Meta, key)
	return r
}

// LookupMeta value identified by the given key. If not found in this route,
// the value is recursively fetched in the parent routers.
//
// Returns the value and `true` if found in the current route or one of the
// parent routers, `nil` and `false` otherwise.
func (r *Route) LookupMeta(key string) (any, bool) {
	val, ok := r.Meta[key]
	if ok {
		return val, ok
	}
	if r.parent == nil {
		return nil, false
	}
	return r.parent.LookupMeta(key)
}

// ValidateBody adds (or replace) validation rules for the request body.
func (r *Route) ValidateBody(validationRules RuleSetFunc) *Route {
	validationMiddleware := findMiddleware[*validateRequestMiddleware](r.middleware)
	if validationMiddleware == nil {
		r.Middleware(&validateRequestMiddleware{BodyRules: validationRules})
	} else {
		validationMiddleware.BodyRules = validationRules
	}
	return r
}

// ValidateQuery adds (or replace) validation rules for the request query.
func (r *Route) ValidateQuery(validationRules RuleSetFunc) *Route {
	validationMiddleware := findMiddleware[*validateRequestMiddleware](r.middleware)
	if validationMiddleware == nil {
		r.Middleware(&validateRequestMiddleware{QueryRules: validationRules})
	} else {
		validationMiddleware.QueryRules = validationRules
	}
	return r
}

// CORS set the CORS options for this route only.
// The "OPTIONS" method is added if this route doesn't already support it.
//
// If the options are not `nil`, the CORS middleware is automatically added globally.
// To disable CORS, give `nil` options. The "OPTIONS" method will be removed
// if it isn't the only method for this route.
func (r *Route) CORS(options *cors.Options) *Route {
	i := lo.IndexOf(r.methods, http.MethodOptions)
	if options == nil {
		r.Meta[MetaCORS] = nil
		if len(r.methods) > 1 && i != -1 {
			r.methods = append(r.methods[:i], r.methods[i+1:]...)
		}
		return r
	}
	r.Meta[MetaCORS] = options
	if !hasMiddleware[*corsMiddleware](r.parent.globalMiddleware.middleware) {
		r.parent.GlobalMiddleware(&corsMiddleware{})
	}
	if i == -1 {
		r.methods = append(r.methods, http.MethodOptions)
	}
	return r
}

// Middleware register middleware for this route only.
//
// Returns itself.
func (r *Route) Middleware(middleware ...Middleware) *Route {
	r.middleware = append(r.middleware, middleware...)
	for _, m := range middleware {
		m.Init(r.parent.server)
	}
	return r
}

// BuildURL build a full URL pointing to this route.
// Panics if the amount of parameters doesn't match the amount of
// actual parameters for this route.
func (r *Route) BuildURL(parameters ...string) string {
	return r.parent.server.BaseURL() + r.BuildURI(parameters...)
}

// BuildProxyURL build a full URL pointing to this route using the proxy base URL.
// Panics if the amount of parameters doesn't match the amount of
// actual parameters for this route.
func (r *Route) BuildProxyURL(parameters ...string) string {
	return r.parent.server.ProxyBaseURL() + r.BuildURI(parameters...)
}

// BuildURI build a full URI pointing to this route. The returned
// string doesn't include the protocol and domain. (e.g. "/user/login")
// Panics if the amount of parameters doesn't match the amount of
// actual parameters for this route.
func (r *Route) BuildURI(parameters ...string) string {
	fullURI, fullParameters := r.GetFullURIAndParameters()

	if len(parameters) != len(fullParameters) {
		panic(errors.Errorf("BuildURI: route has %d parameters, %d given", len(fullParameters), len(parameters)))
	}

	var builder strings.Builder
	builder.Grow(len(fullURI))

	idxs, _ := r.braceIndices(fullURI)
	length := len(idxs)
	end := 0
	currentParam := 0
	for i := 0; i < length; i += 2 {
		raw := fullURI[end:idxs[i]]
		end = idxs[i+1]
		builder.WriteString(raw)
		builder.WriteString(parameters[currentParam])
		currentParam++
		end++ // Skip closing braces
	}
	builder.WriteString(fullURI[end:])

	return builder.String()
}

// GetName get the name of this route.
func (r *Route) GetName() string {
	return r.name
}

// GetURI get the URI of this route.
// The returned URI is relative to the parent router of this route, it is NOT
// the full path to this route.
//
// Note that this URI may contain route parameters in their définition format.
// Use the request's URI if you want to see the URI as it was requested by the client.
func (r *Route) GetURI() string {
	return r.uri
}

// GetFullURI get the full URI of this route.
//
// Note that this URI may contain route parameters in their définition format.
// Use the request's URI if you want to see the URI as it was requested by the client.
func (r *Route) GetFullURI() string {
	router := r.parent
	segments := make([]string, 0, 3)
	segments = append(segments, r.uri)

	for router != nil {
		segments = append(segments, router.prefix)
		router = router.parent
	}

	// Revert segements
	for i := len(segments)/2 - 1; i >= 0; i-- {
		opp := len(segments) - 1 - i
		segments[i], segments[opp] = segments[opp], segments[i]
	}

	return strings.Join(segments, "")
}

// GetMethods returns the methods the route matches against.
func (r *Route) GetMethods() []string {
	cpy := make([]string, len(r.methods))
	copy(cpy, r.methods)
	return cpy
}

// GetHandler returns the Handler associated with this route.
func (r *Route) GetHandler() Handler {
	return r.handler
}

// GetParent returns the parent Router of this route.
func (r *Route) GetParent() *Router {
	return r.parent
}

// GetFullURIAndParameters get the full uri and parameters for this route and all its parent routers.
func (r *Route) GetFullURIAndParameters() (string, []string) {
	router := r.parent
	segments := make([]string, 0, 3)
	segments = append(segments, r.uri)

	parameters := make([]string, 0, len(r.parameters))
	for i := len(r.parameters) - 1; i >= 0; i-- {
		parameters = append(parameters, r.parameters[i])
	}

	for router != nil {
		segments = append(segments, router.prefix)
		for i := len(router.parameters) - 1; i >= 0; i-- {
			parameters = append(parameters, router.parameters[i])
		}
		router = router.parent
	}

	// Revert segements
	for i := len(segments)/2 - 1; i >= 0; i-- {
		opp := len(segments) - 1 - i
		segments[i], segments[opp] = segments[opp], segments[i]
	}

	// Revert parameters
	for i := len(parameters)/2 - 1; i >= 0; i-- {
		opp := len(parameters) - 1 - i
		parameters[i], parameters[opp] = parameters[opp], parameters[i]
	}

	return strings.Join(segments, ""), parameters
}
