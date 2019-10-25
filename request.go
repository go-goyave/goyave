package goyave

import (
	"net/http"
	"net/url"

	"github.com/System-Glitch/goyave/validation"
)

// Request struct represents an http request.
// Contains the validated body in the Data attribute if the route was defined with a request generator function
type Request struct {
	httpRequest *http.Request
	cookies     []*http.Cookie
	Rules       validation.RuleSet
	Data        map[string]interface{}
	Params      map[string]string
	Lang        string
}

// Method specifies the HTTP method (GET, POST, PUT, etc.).
func (r *Request) Method() string {
	return r.httpRequest.Method
}

// Protocol the protocol used by this request, "HTTP/1.0" for example
func (r *Request) Protocol() string {
	return r.httpRequest.Proto
}

// URL specifies either the URI being requested (for server
// requests) or the URL to access (for client requests).
// Use this if you absolutely need the raw query params, url, etc.
// Otherwise use the provided methods and fields of the "goyave.Request"
func (r *Request) URL() *url.URL {
	return r.httpRequest.URL
}

// Header contains the request header fields either received
// by the server or to be sent by the client.
//
// If a server received a request with header lines,
//
//	Host: example.com
//	accept-encoding: gzip, deflate
//	Accept-Language: en-us
//	fOO: Bar
//	foo: two
//
// then
//
//	Header = map[string][]string{
//		"Accept-Encoding": {"gzip, deflate"},
//		"Accept-Language": {"en-us"},
//		"Foo": {"Bar", "two"},
//	}
//
// HTTP defines that header names are case-insensitive. The
// request parser implements this by using CanonicalHeaderKey,
// making the first character and any characters following a
// hyphen uppercase and the rest lowercase.
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

// Cookies returns the HTTP cookies sent with the request
func (r *Request) Cookies(name string) []*http.Cookie {
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

// Redirect send a permanent redirect response
//
// This method is not part of the response helpers to keep the original
// request encapsulated.
func (r *Request) Redirect(w http.ResponseWriter, url string) {
	http.Redirect(w, r.httpRequest, url, http.StatusPermanentRedirect)
}

// TemporaryRedirect send a temporary redirect response
//
// This method is not part of the response helpers to keep the original
// request encapsulated.
func (r *Request) TemporaryRedirect(w http.ResponseWriter, url string) {
	http.Redirect(w, r.httpRequest, url, http.StatusTemporaryRedirect)
}

func (r *Request) validate() map[string]validation.Errors {
	if r.Rules == nil {
		return nil
	}

	errors := validation.Validate(r.httpRequest, r.Data, r.Rules)
	if len(errors) > 0 {
		return map[string]validation.Errors{"validationError": errors}
	}

	return nil
}
