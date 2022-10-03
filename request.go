package goyave

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/imdario/mergo"
	"goyave.dev/goyave/v4/cors"
	"goyave.dev/goyave/v4/util/fsutil"

	"github.com/google/uuid"
	"goyave.dev/goyave/v4/validation"
)

// Request struct represents an http request.
// Contains the validated body in the Data attribute if the route was defined with a request generator function
type Request struct {
	httpRequest *http.Request
	corsOptions *cors.Options
	route       *Route
	Rules       *validation.Rules
	Params      map[string]string
	Data        map[string]interface{}
	Extra       map[string]interface{}
	User        interface{}
	Lang        string
	cookies     []*http.Cookie
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

// URI specifies the URI being requested.
// Use this if you absolutely need the raw query params, url, etc.
// Otherwise use the provided methods and fields of the "goyave.Request".
func (r *Request) URI() *url.URL {
	return r.httpRequest.URL
}

// Route returns the current route.
func (r *Request) Route() *Route {
	return r.route
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

// CORSOptions returns the CORS options applied to this request, or nil.
// The returned object is a copy of the options applied to the router.
// Therefore, altering the returned object will not alter the router's options.
func (r *Request) CORSOptions() *cors.Options {
	if r.corsOptions == nil {
		return nil
	}

	cpy := *r.corsOptions
	return &cpy
}

// Has check if the given field exists in the request data.
func (r *Request) Has(field string) bool {
	_, exists := r.Data[field]
	return exists
}

// String get a string field from the request data.
// Panics if the field is not a string.
func (r *Request) String(field string) string {
	str, ok := r.Data[field].(string)
	if !ok {
		panic(fmt.Sprintf("Field \"%s\" is not a string", field))
	}
	return str
}

// Numeric get a numeric field from the request data.
// Panics if the field is not numeric.
func (r *Request) Numeric(field string) float64 {
	str, ok := r.Data[field].(float64)
	if !ok {
		panic(fmt.Sprintf("Field \"%s\" is not numeric", field))
	}
	return str
}

// Integer get an integer field from the request data.
// Panics if the field is not an integer.
func (r *Request) Integer(field string) int {
	str, ok := r.Data[field].(int)
	if !ok {
		panic(fmt.Sprintf("Field \"%s\" is not an integer", field))
	}
	return str
}

// Bool get a bool field from the request data.
// Panics if the field is not a bool.
func (r *Request) Bool(field string) bool {
	str, ok := r.Data[field].(bool)
	if !ok {
		panic(fmt.Sprintf("Field \"%s\" is not a bool", field))
	}
	return str
}

// File get a file field from the request data.
// Panics if the field is not numeric.
func (r *Request) File(field string) []fsutil.File {
	str, ok := r.Data[field].([]fsutil.File)
	if !ok {
		panic(fmt.Sprintf("Field \"%s\" is not a file", field))
	}
	return str
}

// Timezone get a timezone field from the request data.
// Panics if the field is not a timezone.
func (r *Request) Timezone(field string) *time.Location {
	str, ok := r.Data[field].(*time.Location)
	if !ok {
		panic(fmt.Sprintf("Field \"%s\" is not a timezone", field))
	}
	return str
}

// IP get an IP field from the request data.
// Panics if the field is not an IP.
func (r *Request) IP(field string) net.IP {
	str, ok := r.Data[field].(net.IP)
	if !ok {
		panic(fmt.Sprintf("Field \"%s\" is not an IP", field))
	}
	return str
}

// URL get a URL field from the request data.
// Panics if the field is not a URL.
func (r *Request) URL(field string) *url.URL {
	str, ok := r.Data[field].(*url.URL)
	if !ok {
		panic(fmt.Sprintf("Field \"%s\" is not a URL", field))
	}
	return str
}

// UUID get a UUID field from the request data.
// Panics if the field is not a UUID.
func (r *Request) UUID(field string) uuid.UUID {
	str, ok := r.Data[field].(uuid.UUID)
	if !ok {
		panic(fmt.Sprintf("Field \"%s\" is not an UUID", field))
	}
	return str
}

// Date get a date field from the request data.
// Panics if the field is not a date.
func (r *Request) Date(field string) time.Time {
	str, ok := r.Data[field].(time.Time)
	if !ok {
		panic(fmt.Sprintf("Field \"%s\" is not a date", field))
	}
	return str
}

// Object get an object field from the request data.
// Panics if the field is not an object.
func (r *Request) Object(field string) map[string]interface{} {
	str, ok := r.Data[field].(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("Field \"%s\" is not an object", field))
	}
	return str
}

// ToStruct map the request data to a struct.
//
//	 type UserInsertRequest struct {
//		 Username string
//		 Email string
//	 }
//	 //...
//	 userInsertRequest := UserInsertRequest{}
//	 if err := request.ToStruct(&userInsertRequest); err != nil {
//	  panic(err)
//	 }
func (r *Request) ToStruct(dst interface{}) error {
	return mergo.Map(dst, r.Data)
}

func (r *Request) validate() validation.Errors {
	if r.Rules == nil {
		return nil
	}

	extra := map[string]interface{}{
		"request": r,
	}
	contentType := r.httpRequest.Header.Get("Content-Type")
	return validation.ValidateWithExtra(r.Data, r.Rules, strings.HasPrefix(contentType, "application/json"), r.Lang, extra)
}
