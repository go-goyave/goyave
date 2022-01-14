package cors

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"goyave.dev/goyave/v4/util/sliceutil"
)

// Options holds the CORS configuration for a router.
type Options struct {
	// AllowedOrigins is a list of origins a cross-domain request can be executed from.
	// If the first value in the slice is "*" or if the slice is empty, all origins will be allowed.
	// Default value is ["*"]
	AllowedOrigins []string

	// AllowedMethods is a list of methods the client is allowed to use with cross-domain requests.
	// Default value is ["HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"].
	AllowedMethods []string

	// AllowedHeaders is a list of non simple headers the client is allowed to use with
	// cross-domain requests.
	// If the first value in the slice is "*", all headers will be allowed.
	// If the slice is empty, the request's headers will be reflected.
	// Default value is ["Origin", "Accept", "Content-Type", "X-Requested-With", "Authorization"].
	AllowedHeaders []string

	// ExposedHeaders indicates which headers are safe to expose to the API of a CORS
	// API specification
	ExposedHeaders []string

	// MaxAge indicates how long the results of a preflight request can be cached.
	// Default is 12 hours.
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

// ConfigureCommon configures common headers between regular and preflight requests:
// Origin, Credentials and Exposed Headers.
func (o *Options) ConfigureCommon(headers http.Header, requestHeaders http.Header) {
	o.configureOrigin(headers, requestHeaders)
	o.configureCredentials(headers)
	o.configureExposedHeaders(headers)
}

func (o *Options) configureOrigin(headers http.Header, requestHeaders http.Header) {
	if len(o.AllowedOrigins) == 0 || o.AllowedOrigins[0] == "*" {
		headers.Set("Access-Control-Allow-Origin", "*")
	} else {
		if o.validateOrigin(requestHeaders) {
			headers.Set("Access-Control-Allow-Origin", requestHeaders.Get("Origin"))
		}
		headers.Add("Vary", "Origin")
	}
}

func (o *Options) configureCredentials(headers http.Header) {
	if o.AllowCredentials {
		headers.Set("Access-Control-Allow-Credentials", "true")
	}
}

func (o *Options) configureExposedHeaders(headers http.Header) {
	if len(o.ExposedHeaders) > 0 {
		headers.Set("Access-Control-Expose-Headers", strings.Join(o.ExposedHeaders, ", "))
	}
}

func (o *Options) configureAllowedMethods(headers http.Header) {
	headers.Set("Access-Control-Allow-Methods", strings.Join(o.AllowedMethods, ", "))
}

func (o *Options) configureAllowedHeaders(headers http.Header, requestHeaders http.Header) {
	if len(o.AllowedHeaders) == 0 {
		headers.Add("Vary", "Access-Control-Request-Headers")
		headers.Set("Access-Control-Allow-Headers", requestHeaders.Get("Access-Control-Request-Headers"))
	} else {
		headers.Set("Access-Control-Allow-Headers", strings.Join(o.AllowedHeaders, ", "))
	}

}

func (o *Options) configureMaxAge(headers http.Header) {
	headers.Set("Access-Control-Max-Age", strconv.FormatUint(uint64(o.MaxAge.Seconds()), 10))
}

// HandlePreflight configures headers for preflight requests:
// Allowed Methods, Allowed Headers and Max Age.
func (o *Options) HandlePreflight(headers http.Header, requestHeaders http.Header) {
	o.configureAllowedMethods(headers)
	o.configureAllowedHeaders(headers, requestHeaders)
	o.configureMaxAge(headers)
}

func (o *Options) validateOrigin(requestHeaders http.Header) bool {
	return len(o.AllowedOrigins) == 0 ||
		o.AllowedOrigins[0] == "*" ||
		sliceutil.ContainsStr(o.AllowedOrigins, requestHeaders.Get("Origin"))
}
