package goyave

import (
	"fmt"
	"net/http"
	"strings"

	"goyave.dev/goyave/v4/validation"
)

type RouteV5 struct {
	name    string
	uri     string
	methods []string
	parent  *RouterV5
	Meta    map[string]any
	handler HandlerV5
	middlewareHolderV5
	parameterizable
}

var _ routeMatcherV5 = (*RouteV5)(nil) // implements routeMatcher

// newRoute create a new route without any settings except its handler.
// This is used to generate a fake route for the Method Not Allowed and Not Found handlers.
// This route has the core middleware enabled and can be used without a parent router.
// Thus, custom status handlers can use language and body.
func newRouteV5(handler HandlerV5) *RouteV5 {
	return &RouteV5{
		handler: handler,
		Meta:    make(map[string]any),
		middlewareHolderV5: middlewareHolderV5{
			middleware: nil,
		},
	}
}

func (r *RouteV5) match(req *http.Request, match *routeMatchV5) bool {
	if params := r.parameterizable.regex.FindStringSubmatch(match.currentPath); params != nil {
		if r.checkMethod(req.Method) {
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

func (r *RouteV5) checkMethod(method string) bool {
	for _, m := range r.methods {
		if m == method {
			return true
		}
	}
	return false
}

func (r *RouteV5) makeParameters(match []string) map[string]string {
	return r.parameterizable.makeParameters(match, r.parameters)
}

// Name set the name of the route.
// Panics if a route with the same name already exists.
// Returns itself.
func (r *RouteV5) Name(name string) *RouteV5 {
	if r.name != "" {
		panic(fmt.Errorf("Route name is already set"))
	}

	if _, ok := r.parent.namedRoutes[name]; ok {
		panic(fmt.Errorf("Route %q already exists", name))
	}

	r.name = name
	r.parent.namedRoutes[name] = r
	return r
}

func (r *RouteV5) SetMeta(key string, value any) *RouteV5 {
	r.Meta[key] = value
	return r
}

func (r *RouteV5) RemoveMeta(key string) *RouteV5 {
	delete(r.Meta, key)
	return r
}

// Validate adds validation rules to this route. If the user-submitted data
// doesn't pass validation, the user will receive an error and messages explaining
// what is wrong.
//
// Returns itself.
func (r *RouteV5) Validate(validationRules validation.Ruler) *RouteV5 {
	// TODO ValidateQuery too!
	r.Meta[MetaValidationRules] = validationRules.AsRules()
	return r
}

// Middleware register middleware for this route only.
//
// Returns itself.
func (r *RouteV5) Middleware(middleware ...MiddlewareV5) *RouteV5 {
	r.middleware = append(r.middleware, middleware...)
	return r
}

// BuildURL build a full URL pointing to this route.
// Panics if the amount of parameters doesn't match the amount of
// actual parameters for this route.
// func (r *RouteV5) BuildURL(parameters ...string) string {
// 	return r.parent.BaseURL() + r.BuildURI(parameters...)
//  FIXME BaseURL accessed from?
// }

// BuildURI build a full URI pointing to this route. The returned
// string doesn't include the protocol and domain. (e.g. "/user/login")
// Panics if the amount of parameters doesn't match the amount of
// actual parameters for this route.
func (r *RouteV5) BuildURI(parameters ...string) string {
	fullURI, fullParameters := r.GetFullURIAndParameters()

	if len(parameters) != len(fullParameters) {
		panic(fmt.Errorf("BuildURI: route has %d parameters, %d given", len(fullParameters), len(parameters)))
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
func (r *RouteV5) GetName() string {
	return r.name
}

// GetURI get the URI of this route.
// The returned URI is relative to the parent router of this route, it is NOT
// the full path to this route.
//
// Note that this URI may contain route parameters in their définition format.
// Use the request's URI if you want to see the URI as it was requested by the client.
func (r *RouteV5) GetURI() string {
	return r.uri
}

// GetFullURI get the full URI of this route.
//
// Note that this URI may contain route parameters in their définition format.
// Use the request's URI if you want to see the URI as it was requested by the client.
func (r *RouteV5) GetFullURI() string {
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
func (r *RouteV5) GetMethods() []string {
	cpy := make([]string, len(r.methods))
	copy(cpy, r.methods)
	return cpy
}

// GetHandler returns the Handler associated with this route.
func (r *RouteV5) GetHandler() HandlerV5 {
	return r.handler
}

func (r *RouteV5) GetParent() *RouterV5 {
	return r.parent
}

// GetValidationRules returns the validation rules associated with this route.
func (r *RouteV5) GetValidationRules() *validation.Rules {
	return r.Meta[MetaValidationRules].(*validation.Rules)
}

// GetFullURIAndParameters get the full uri and parameters for this route and all its parent routers.
func (r *RouteV5) GetFullURIAndParameters() (string, []string) {
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
