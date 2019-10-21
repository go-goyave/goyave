package goyave

import (
	"net/http"

	"github.com/System-Glitch/goyave/config"
	"github.com/gorilla/mux"
)

// Router registers routes to be matched and dispatches a handler.
type Router struct {
	muxRouter *mux.Router
}

func newRouter() *Router {
	muxRouter := mux.NewRouter()
	muxRouter.Schemes(config.Get("protocol").(string))
	return &Router{muxRouter: muxRouter}
}

// Subrouter create a new sub-router from this router.
// Use subrouters to create route groups and to apply middlewares to multiple routes.
func (r *Router) Subrouter(prefix string) *Router {
	return &Router{muxRouter: r.muxRouter.PathPrefix(prefix).Subrouter()}
}

// Middleware apply one or more middleware(s) to the route group.
func (r *Router) Middleware(middlewares ...func(http.Handler) http.Handler) {
	// TODO implement middleware
}

// Route register a new route.
func (r *Router) Route(method string, endpoint string, handler func(http.ResponseWriter, *Request), request *Request) {
	// TODO implement route
	r.muxRouter.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		// TODO handle url params
		handler(w, nil) // TODO handle request and pass param
	}).Methods(method)
}
