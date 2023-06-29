package goyave

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"goyave.dev/goyave/v4/lang"
)

const (
	// ExtraError the key used in `Context.Extra` to store an error
	// reported with the Error function or via the recovery middleware.
	ExtraError = "goyave.error"

	// ExtraStacktrace the key used in `Context.Extra` to store the
	// stacktrace if debug is enabled and an error is reported.
	ExtraStacktrace = "goyave.stacktrace"

	// ExtraBodyValidationRules the key used in `Context.Extra` to
	// store the body validation rules.
	ExtraBodyValidationRules = "goyave.bodyValidationRules"

	// ExtraQueryValidationRules the key used in `Context.Extra` to
	// store the query validation rules.
	ExtraQueryValidationRules = "goyave.queryValidationRules"

	// ExtraValidationError the key used in `Context.Extra` to
	// store the body validation errors.
	ExtraValidationError = "goyave.validationError"

	// ExtraQueryValidationError the key used in `Context.Extra` to
	// store the query validation errors.
	ExtraQueryValidationError = "goyave.queryValidationError"

	// ExtraJWTClaims when using the built-in `JWTAuthenticator`, this
	// extra key can be used to retrieve the JWT claims.
	ExtraJWTClaims = "goyave.jwtClaims"
)

// Request represents an http request received by the server.
type Request struct {
	httpRequest *http.Request
	Now         time.Time
	Data        any
	User        any
	Query       map[string]any
	Lang        *lang.Language
	Extra       map[string]any
	Route       *Route
	RouteParams map[string]string
	cookies     []*http.Cookie
}

// NewRequest create a new Request from the given raw http request.
// Initializes Now with the current time and Extra with a non-nil map.
func NewRequest(httpRequest *http.Request) *Request {
	return &Request{
		httpRequest: httpRequest,
		Now:         time.Now(),
		Extra:       map[string]any{},
		// Route is set by the router
		// Lang is set inside the language middleware
		// Query is set inside the parse request middleware
	}
}

// Request return the raw http request.
// Prefer using the "goyave.Request" accessors.
func (r *Request) Request() *http.Request {
	return r.httpRequest
}

// Method specifies the HTTP method (GET, POST, PUT, etc.).
func (r *Request) Method() string {
	return r.httpRequest.Method
}

// Protocol the protocol used by this request, "HTTP/1.1" for example.
func (r *Request) Protocol() string {
	return r.httpRequest.Proto
}

// URL specifies the URL being requested.
func (r *Request) URL() *url.URL {
	return r.httpRequest.URL
}

// Header contains the request header fields either received
// by the server or to be sent by the client.
// Header names are case-insensitive.
//
// If the raw request has the following header lines,
//
//	Host: example.com
//	accept-encoding: gzip, deflate
//	Accept-Language: en-us
//	fOO: Bar
//	foo: two
//
// then the header map will look like this:
//
//	Header = map[string][]string{
//		"Accept-Encoding": {"gzip, deflate"},
//		"Accept-Language": {"en-us"},
//		"Foo": {"Bar", "two"},
//	}
func (r *Request) Header() http.Header {
	return r.httpRequest.Header
}

// ContentLength records the length of the associated content.
// The value -1 indicates that the length is unknown.
func (r *Request) ContentLength() int64 {
	return r.httpRequest.ContentLength
}

// RemoteAddress allows to record the network address that
// sent the request, usually for logging.
func (r *Request) RemoteAddress() string {
	return r.httpRequest.RemoteAddr
}

// Cookies returns the HTTP cookies sent with the request.
func (r *Request) Cookies() []*http.Cookie {
	if r.cookies == nil {
		r.cookies = r.httpRequest.Cookies()
	}
	return r.cookies
}

// Referrer returns the referring URL, if sent in the request.
func (r *Request) Referrer() string {
	return r.httpRequest.Referer()
}

// UserAgent returns the client's User-Agent, if sent in the request.
func (r *Request) UserAgent() string {
	return r.httpRequest.UserAgent()
}

// BasicAuth returns the username and password provided in the request's
// Authorization header, if the request uses HTTP Basic Authentication.
func (r *Request) BasicAuth() (username, password string, ok bool) {
	return r.httpRequest.BasicAuth()
}

// BearerToken extract the auth token from the "Authorization" header.
// Only takes tokens of type "Bearer".
// Returns empty string if no token found or the header is invalid.
func (r *Request) BearerToken() (string, bool) {
	const schema = "Bearer "
	header := r.Header().Get("Authorization")
	if !strings.HasPrefix(header, schema) {
		return "", false
	}
	return strings.TrimSpace(header[len(schema):]), true
}

// Body the request body.
// Always non-nil, but will return EOF immediately when no body is present.
// The server will close the request body so handlers don't need to.
func (r *Request) Body() io.ReadCloser {
	return r.httpRequest.Body
}
