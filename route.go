package goyave

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/helper"
	"github.com/System-Glitch/goyave/v2/validation"
)

// Route stores information for matching and serving.
type Route struct {
	name            string
	uri             string
	methods         []string
	parent          *Router
	handler         Handler
	validationRules validation.RuleSet
	middlewareHolder
	parametrizeable
}

var _ routeMatcher = (*Route)(nil) // implements routeMatcher

// newRoute create a new route without any settings except its handler.
// This is used to generate a fake route for the Method Not Allowed and Not Found handlers.
func newRoute(handler Handler) *Route {
	return &Route{handler: handler}
}

func (r *Route) match(req *http.Request, match *routeMatch) bool {
	if params := r.parametrizeable.regex.FindStringSubmatch(match.currentPath); params != nil {
		if helper.Contains(r.methods, req.Method) {
			match.trimCurrentPath(params[0])
			match.mergeParams(r.makeParameters(params[1:]))
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

func (r *Route) makeParameters(match []string) map[string]string {
	return r.parametrizeable.makeParameters(match, r.parameters)
}

// Name set the name of the route.
// Panics if a route with the same name already exists.
func (r *Route) Name(name string) {
	r.name = name

	if _, ok := r.parent.namedRoutes[name]; ok {
		panic(fmt.Errorf("Route %q already exists", name))
	}
	r.parent.namedRoutes[name] = r
}

// BuildURL build a full URL pointing to this route.
// Panics if the amount of parameters doesn't match the amount of
// actual parameters for this route.
func (r *Route) BuildURL(parameters ...string) string {
	if len(parameters) != len(r.parameters) {
		panic(fmt.Errorf("BuildURL: route has %d parameters, %d given", len(r.parameters), len(parameters)))
	}

	address := getAddress(config.GetString("protocol"))

	var builder strings.Builder
	builder.Grow(len(r.uri) + len(address))

	builder.WriteString(address)

	idxs, _ := r.braceIndices(r.uri)
	length := len(idxs)
	end := 0
	currentParam := 0
	for i := 0; i < length; i += 2 {
		raw := r.uri[end:idxs[i]]
		end = idxs[i+1]
		builder.WriteString(raw)
		builder.WriteString(parameters[currentParam])
		currentParam++
		end++ // Skip closing braces
	}
	builder.WriteString(r.uri[end:])

	return builder.String()
}

// GetName get the name of this route.
func (r *Route) GetName() string {
	return r.name
}

// GetURI get the URI of this route.
// Note that this URI may contain route parameters in their définition format.
// Use the request's URI if you want to see the URI as it was requested by the client.
func (r *Route) GetURI() string {
	return r.uri
}

// TODO get full URI when tree-like router implementation done

// GetMethods returns the methods the route matches against.
func (r *Route) GetMethods() []string {
	cpy := make([]string, len(r.methods))
	copy(cpy, r.methods)
	return cpy
}
