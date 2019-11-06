package goyave

import (
	"net/http"
	"strings"

	"github.com/System-Glitch/goyave/config"
	"github.com/System-Glitch/goyave/helpers/filesystem"
	"github.com/System-Glitch/goyave/validation"
	"github.com/gorilla/mux"
)

// Router registers routes to be matched and dispatches a handler.
type Router struct {
	muxRouter   *mux.Router
	middlewares []Middleware
}

// Handler is a controller function
type Handler func(*Response, *Request)

func newRouter() *Router {
	muxRouter := mux.NewRouter()
	muxRouter.Schemes(config.GetString("protocol"))
	router := &Router{muxRouter: muxRouter}
	router.Middleware(recoveryMiddleware, parseRequestMiddleware, languageMiddleware)
	return router
}

// Subrouter create a new sub-router from this router.
// Use subrouters to create route groups and to apply middlewares to multiple routes.
func (r *Router) Subrouter(prefix string) *Router {
	router := &Router{muxRouter: r.muxRouter.PathPrefix(prefix).Subrouter()}

	// Apply parent middlewares to subrouter
	router.Middleware(r.middlewares...)
	return router
}

// Middleware apply one or more middleware(s) to the route group.
func (r *Router) Middleware(middlewares ...Middleware) {
	r.middlewares = append(r.middlewares, middlewares...)
}

// Route register a new route.
func (r *Router) Route(method string, endpoint string, handler Handler, validationRules validation.RuleSet) {
	r.muxRouter.HandleFunc(endpoint, func(w http.ResponseWriter, rawRequest *http.Request) {
		r.requestHandler(w, rawRequest, handler, validationRules)
	}).Methods(method)
}

// Static serve a directory and its subdirectories of static resources.
// Set the "download" attribute to true if you want the files to be sent as an attachment.
//
// If no file is given in the url, or if the given file is a directory, the handler will
// send the "index.html" file if it exists.
func (r *Router) Static(endpoint string, directory string, download bool) {
	r.Route("GET", endpoint+"{resource:.*}", staticHandler(directory, download), nil)
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
	if len(file) <= 0 || filesystem.IsDirectory(path) {
		if strings.Count(file, "/") > 0 && !strings.HasSuffix(file, "/") {
			file += "/"
		}
		file += "index.html"
		path = directory + "/" + file
	}
	return path
}

func (r *Router) requestHandler(w http.ResponseWriter, rawRequest *http.Request, handler Handler, rules validation.RuleSet) {
	request := &Request{
		httpRequest: rawRequest,
		Rules:       rules,
		Params:      mux.Vars(rawRequest),
	}
	response := &Response{
		httpRequest: rawRequest,
		writer:      w,
		empty:       true,
	}

	// Validate last.
	// Allows custom middlewares to be executed after core
	// middlewares and before validation.
	handler = validateRequestMiddleware(handler)
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		handler = r.middlewares[i](handler)
	}

	handler(response, request)

	// If the response is empty, return status 204 to
	// comply with RFC 7231, 6.3.5
	if response.empty {
		response.Status(http.StatusNoContent)
	}
}
