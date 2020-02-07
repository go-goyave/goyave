package goyave

import (
	"net/http"

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
	if params := r.parametrizeable.regex.FindStringSubmatch(req.URL.Path); params != nil {
		if helper.Contains(r.methods, req.Method) {
			match.parameters = r.makeParameters(params[1:])
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
