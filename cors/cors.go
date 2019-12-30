package cors

import (
	"net/http"
	"time"
)

// Options holds the CORS configuration for a router.
type Options struct {
	// AllowedOrigins is a list of origins a cross-domain request can be executed from.
	// If the first value in the slice is "*" or if the slice is empty, all origins will be allowed.
	// Default value is ["*"]
	AllowedOrigins []string

	// AllowedMethods is a list of methods the client is allowed to use with
	// cross-domain requests. Default value is ["HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"].
	AllowedMethods []string

	// AllowedHeaders is list of non simple headers the client is allowed to use with
	// cross-domain requests.
	// If the first value in the slice is "*", all headers will be allowed.
	// If the slice is empty, the request's headers will be reflected.
	// Default value is ["Origin", "Accept", "Content-Type", "X-Requested-With", "Authorization"].
	AllowedHeaders []string

	// ExposedHeaders indicates which headers are safe to expose to the API of a CORS
	// API specification
	ExposedHeaders []string

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached. Default is 12 hours.
	MaxAge time.Duration

	// AllowCredentials indicates whether the request can include user credentials like
	// cookies, HTTP authentication or client side SSL certificates.
	AllowCredentials bool

	// OptionsPassthrough instructs preflight to let other potential next handlers to
	// process the OPTIONS method. Turn this on if your application handles OPTIONS.
	OptionsPassthrough bool
}

// Default create new CORS options with default settings.
// The returned value can be used as a starting point for
// customized options.
func Default() *Options {
	return &Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"Origin", "Accept", "Content-Type", "X-Requested-With", "Authorization"},
		AllowCredentials: false,
		MaxAge:           time.Hour * 12,
	}
}
