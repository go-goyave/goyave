package goyave

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/System-Glitch/goyave/v2/helper"
	"github.com/System-Glitch/goyave/v2/validation"
)

type parametrizeable struct {
	regex      *regexp.Regexp
	parameters []string
}

// compileParameters parse the route parameters and compiles their regexes if needed.
func (p *parametrizeable) compileParameters(uri string) {
	idxs, err := p.braceIndices(uri)
	if err != nil {
		panic(err)
	}

	fullPattern := "^"
	length := len(idxs)
	if length > 0 {
		end := 0
		for i := 0; i < length; i += 2 {
			raw := uri[end:idxs[i]]
			end = idxs[i+1]
			sub := uri[idxs[i]+1 : end]
			parts := strings.SplitN(sub, ":", 2)
			if parts[0] == "" {
				panic(fmt.Errorf("invalid route parameter, missing name in %q", sub))
			}
			pattern := "[^/]+" // default pattern
			if len(parts) == 2 {
				pattern = parts[1]
				if pattern == "" {
					panic(fmt.Errorf("invalid route parameter, missing pattern in %q", sub))
				}
			}

			fullPattern += raw
			fullPattern += "(" + pattern + ")" // TODO find more efficient way of building the string
			end++                              // Skip closing braces
			p.parameters = append(p.parameters, parts[0])
		}
		fullPattern += uri[end:]
	} else {
		fullPattern += uri
	}

	fullPattern += "$"

	p.regex = regexp.MustCompile(fullPattern) // TODO optimize by checking if pattern already exists

	if p.regex.NumSubexp() != length/2 {
		panic(fmt.Sprintf("route %s contains capture groups in its regexp. ", uri) +
			"Only non-capturing groups are accepted: e.g. (?:pattern) instead of (pattern)")
	}
}

// braceIndices returns the first level curly brace indices from a string.
// It returns an error in case of unbalanced braces.
func (p *parametrizeable) braceIndices(s string) ([]int, error) {
	var level, idx int
	indices := make([]int, 0, 2)
	length := len(s)
	for i := 0; i < length; i++ {
		if s[i] == '{' {
			level++
			if level == 1 {
				idx = i
			}
		} else if s[i] == '}' {
			level--
			if level == 0 {
				if i == idx+1 {
					return nil, fmt.Errorf("empty route parameter in %q", s)
				}
				indices = append(indices, idx, i)
			} else if level < 0 {
				return nil, fmt.Errorf("unbalanced braces in %q", s)
			}
		}
	}
	if level != 0 {
		return nil, fmt.Errorf("unbalanced braces in %q", s)
	}
	return indices, nil
}

// makeParameters from a regex match and the given parameter names.
func (p *parametrizeable) makeParameters(match []string, names []string) map[string]string {
	length := len(match)
	params := make(map[string]string, length)
	for i, v := range match {
		params[names[i]] = v
	}
	return params
}

//------------------------------------------------------

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

func newRoute(handler Handler) *Route {
	return &Route{handler: handler}
}

func (r *Route) match(req *http.Request, match *routeMatch) bool {
	if params := r.parametrizeable.regex.FindStringSubmatch(req.URL.Path); params != nil && len(params)-1 == len(r.parameters) {
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

// Name get the name of this route.
func (r *Route) Name() string {
	return r.name
}

// URI get the URI of this route.
// Note that this URI may contain route parameters in their définition format.
// Use the request's URI if you want to see the URI as it was requested by the client.
func (r *Route) URI() string {
	return r.uri
}

// TODO implement more getters
