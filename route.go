package goyave

import (
	"fmt"
	"net/http"
	"strings"

	"goyave.dev/goyave/v4/validation"
)

// Route stores information for matching and serving.
type Route struct {
	name            string
	uri             string
	methods         []string
	parent          *Router
	handler         Handler
	validationRules *validation.Rules
	middlewareHolder
	parameterizable
}

var _ routeMatcher = (*Route)(nil) // implements routeMatcher

// newRoute create a new route without any settings except its handler.
// This is used to generate a fake route for the Method Not Allowed and Not Found handlers.
// This route has the core middleware enabled and can be used without a parent router.
// Thus, custom status handlers can use language and body.
func newRoute(handler Handler) *Route {
	return &Route{
		handler: handler,
		middlewareHolder: middlewareHolder{
			middleware: []Middleware{recoveryMiddleware, parseRequestMiddleware, languageMiddleware},
		},
	}
}

func (r *Route) match(req *http.Request, match *routeMatch) bool {
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
		panic(fmt.Errorf("Route name is already set"))
	}

	if _, ok := r.parent.namedRoutes[name]; ok {
		panic(fmt.Errorf("Route %q already exists", name))
	}

	r.name = name
	r.parent.namedRoutes[name] = r
	return r
}

// Validate adds validation rules to this route. If the user-submitted data
// doesn't pass validation, the user will receive an error and messages explaining
// what is wrong.
//
// Returns itself.
func (r *Route) Validate(validationRules validation.Ruler) *Route {
	r.validationRules = validationRules.AsRules()
	return r
}

// Middleware register middleware for this route only.
//
// Returns itself.
func (r *Route) Middleware(middleware ...Middleware) *Route {
	r.middleware = middleware
	return r
}

// BuildURL build a full URL pointing to this route.
// Panics if the amount of parameters doesn't match the amount of
// actual parameters for this route.
func (r *Route) BuildURL(parameters ...string) string {
	return BaseURL() + r.BuildURI(parameters...)
}

// BuildURI build a full URI pointing to this route. The returned
// string doesn't include the protocol and domain. (e.g. "/user/login")
// Panics if the amount of parameters doesn't match the amount of
// actual parameters for this route.
func (r *Route) BuildURI(parameters ...string) string {
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

// GetValidationRules returns the validation rules associated with this route.
func (r *Route) GetValidationRules() *validation.Rules {
	return r.validationRules
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
