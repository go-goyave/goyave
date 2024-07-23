package goyave

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"goyave.dev/goyave/v5/lang"
)

type (
	// ExtraBodyValidationRules the key used in `Context.Extra` to
	// store the body validation rules.
	ExtraBodyValidationRules struct{}

	// ExtraQueryValidationRules the key used in `Context.Extra` to
	// store the query validation rules.
	ExtraQueryValidationRules struct{}

	// ExtraValidationError the key used in `Context.Extra` to
	// store the body validation errors.
	ExtraValidationError struct{}

	// ExtraQueryValidationError the key used in `Context.Extra` to
	// store the query validation errors.
	ExtraQueryValidationError struct{}

	// ExtraParseError the key used in `Context.Extra` to
	// store specific parsing errors.
	ExtraParseError struct{}
)

var (
	// ErrInvalidQuery error when an invalid query string is passed.
	ErrInvalidQuery = errors.New("parse middleware: could not parse query")

	// ErrInvalidJSONBody error when an empty or malformed JSON body is sent.
	ErrInvalidJSONBody = errors.New("parse middleware: could not JSON unmarshal body")

	// ErrInvalidContentForType error when e.g. a multipart form is not actually multipart, or empty.
	ErrInvalidContentForType = errors.New("parse middleware: could not parse form")

	// ErrErrorInRequestBody error when e.g. a incoming request is not received properly.
	ErrErrorInRequestBody = errors.New("parse middleware: could not read body")
)

// Request represents a http request received by the server.
type Request struct {
	httpRequest *http.Request
	Now         time.Time
	Data        any
	User        any
	Query       map[string]any
	Lang        *lang.Language

	// Extra can be used to store any extra information related to the request.
	// For example, the JWT middleware stores the token claim in the extras.
	//
	// The keys must be comparable and should not be of type
	// string or any other built-in type to avoid collisions.
	// To avoid allocating when assigning to an `interface{}`, context keys often have
	// concrete type `struct{}`. Alternatively, exported context key variables' static
	// type should be a pointer or interface.
	Extra       map[any]any
	Route       *Route
	RouteParams map[string]string
	cookies     []*http.Cookie
}

var requestPool = sync.Pool{
	New: func() any {
		return &Request{}
	},
}

// NewRequest create a new Request from the given raw http request.
// Initializes Now with the current time and Extra with a non-nil map.
func NewRequest(httpRequest *http.Request) *Request {
	req := requestPool.Get().(*Request)
	req.reset(httpRequest)
	return req
}

func (r *Request) reset(httpRequest *http.Request) {
	r.httpRequest = httpRequest
	r.Now = time.Now()
	r.Extra = map[any]any{}
	r.cookies = nil
	r.Data = nil
	r.Lang = nil
	r.Query = nil
	r.Route = nil
	r.RouteParams = nil
	r.User = nil
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

// Context returns the request's context. To change the context, use `WithContext`.
//
// The returned context is always non-nil; it defaults to the
// background context.
//
// The context is canceled when the client's connection closes, the request is canceled (with HTTP/2),
// or when the `ServeHTTP` method returns (after the finalization step of the request lifecycle).
func (r *Request) Context() context.Context {
	return r.httpRequest.Context()
}

// WithContext creates a shallow copy of the underlying `*http.Request` with
// its context changed to `ctx` then returns itself.
// The provided ctx must be non-nil.
func (r *Request) WithContext(ctx context.Context) *Request {
	r.httpRequest = r.httpRequest.WithContext(ctx)
	return r
}
