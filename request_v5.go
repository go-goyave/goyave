package goyave

import (
	"net/http"
	"net/url"
	"strings"
)

const (
	// ExtraError the key used in Context.Extra to store an error
	// reported with the Error function or via the recovery middleware.
	ExtraError = "goyave.error"

	// ExtraStacktrace the key used in Context.Extra to store the
	// stacktrace if debug is enabled and an error is reported.
	ExtraStacktrace = "goyave.stacktrace"
)

type RequestV5 struct {
	httpRequest *http.Request
	Data        any // TODO separate data from query (two middleware: ValidateQuery, ValidateBody)
	Query       url.Values
	Lang        string
	Extra       map[string]any
	route       *RouteV5
	RouteParams map[string]string
	cookies     []*http.Cookie
}

func newRequest(httpRequest *http.Request, route *RouteV5) *RequestV5 {
	return &RequestV5{
		httpRequest: httpRequest,
		route:       route,
		Query:       httpRequest.URL.Query(),
		Extra:       map[string]any{},
		// Lang is set inside the language middleware
	}
}

// Request return the raw http request.
// Prefer using the "goyave.Request" accessors.
func (r *RequestV5) Request() *http.Request {
	return r.httpRequest
}

// Method specifies the HTTP method (GET, POST, PUT, etc.).
func (r *RequestV5) Method() string {
	return r.httpRequest.Method
}

// Protocol the protocol used by this request, "HTTP/1.1" for example.
func (r *RequestV5) Protocol() string {
	return r.httpRequest.Proto
}

// URL specifies the URL being requested.
func (r *RequestV5) URL() *url.URL {
	return r.httpRequest.URL
}

// RequestHeader contains the request header fields either received
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
func (r *RequestV5) RequestHeader() http.Header {
	return r.httpRequest.Header
}

// ContentLength records the length of the associated content.
// The value -1 indicates that the length is unknown.
func (r *RequestV5) ContentLength() int64 {
	return r.httpRequest.ContentLength
}

// RemoteAddress allows to record the network address that
// sent the request, usually for logging.
func (r *RequestV5) RemoteAddress() string {
	return r.httpRequest.RemoteAddr
}

// Cookies returns the HTTP cookies sent with the request.
func (r *RequestV5) Cookies() []*http.Cookie {
	if r.cookies == nil {
		r.cookies = r.httpRequest.Cookies()
	}
	return r.cookies
}

// Referrer returns the referring URL, if sent in the request.
func (r *RequestV5) Referrer() string {
	return r.httpRequest.Referer()
}

// UserAgent returns the client's User-Agent, if sent in the request.
func (r *RequestV5) UserAgent() string {
	return r.httpRequest.UserAgent()
}

// BasicAuth returns the username and password provided in the request's
// Authorization header, if the request uses HTTP Basic Authentication.
func (r *RequestV5) BasicAuth() (username, password string, ok bool) {
	return r.httpRequest.BasicAuth()
}

// BearerToken extract the auth token from the "Authorization" header.
// Only takes tokens of type "Bearer".
// Returns empty string if no token found or the header is invalid.
func (r *RequestV5) BearerToken() (string, bool) {
	const schema = "Bearer "
	header := r.RequestHeader().Get("Authorization")
	if !strings.HasPrefix(header, schema) {
		return "", false
	}
	return strings.TrimSpace(header[len(schema):]), true
}

// Route returns the current route.
func (r *RequestV5) Route() *RouteV5 {
	return r.route
}
