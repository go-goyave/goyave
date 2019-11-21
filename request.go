package goyave

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/System-Glitch/goyave/helpers/filesystem"
	"github.com/System-Glitch/goyave/validation"
	"github.com/google/uuid"
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
		log.Panicf("Field \"%s\" is not a string", field)
	}
	return str
}

// Numeric get a numeric field from the request data.
// Panics if the field is not numeric.
func (r *Request) Numeric(field string) float64 {
	str, ok := r.Data[field].(float64)
	if !ok {
		log.Panicf("Field \"%s\" is not numeric", field)
	}
	return str
}

// Integer get an integer field from the request data.
// Panics if the field is not an integer.
func (r *Request) Integer(field string) int {
	str, ok := r.Data[field].(int)
	if !ok {
		log.Panicf("Field \"%s\" is not an integer", field)
	}
	return str
}

// Bool get a bool field from the request data.
// Panics if the field is not a bool.
func (r *Request) Bool(field string) bool {
	str, ok := r.Data[field].(bool)
	if !ok {
		log.Panicf("Field \"%s\" is not a bool", field)
	}
	return str
}

// File get a file field from the request data.
// Panics if the field is not numeric.
func (r *Request) File(field string) []filesystem.File {
	str, ok := r.Data[field].([]filesystem.File)
	if !ok {
		log.Panicf("Field \"%s\" is not a file", field)
	}
	return str
}

// Timezone get a timezone field from the request data.
// Panics if the field is not a timezone.
func (r *Request) Timezone(field string) *time.Location {
	str, ok := r.Data[field].(*time.Location)
	if !ok {
		log.Panicf("Field \"%s\" is not a timezone", field)
	}
	return str
}

// IP get an IP field from the request data.
// Panics if the field is not an IP.
func (r *Request) IP(field string) net.IP {
	str, ok := r.Data[field].(net.IP)
	if !ok {
		log.Panicf("Field \"%s\" is not an IP", field)
	}
	return str
}

// URL get an URL field from the request data.
// Panics if the field is not an URL.
func (r *Request) URL(field string) *url.URL {
	str, ok := r.Data[field].(*url.URL)
	if !ok {
		log.Panicf("Field \"%s\" is not an URL", field)
	}
	return str
}

// UUID get a UUID field from the request data.
// Panics if the field is not a UUID.
func (r *Request) UUID(field string) uuid.UUID {
	str, ok := r.Data[field].(uuid.UUID)
	if !ok {
		log.Panicf("Field \"%s\" is not an UUID", field)
	}
	return str
}

// Date get a date field from the request data.
// Panics if the field is not a date.
func (r *Request) Date(field string) time.Time {
	str, ok := r.Data[field].(time.Time)
	if !ok {
		log.Panicf("Field \"%s\" is not a date", field)
	}
	return str
}

func (r *Request) validate() map[string]validation.Errors {
	if r.Rules == nil {
		return nil
	}

	errors := validation.Validate(r.httpRequest, r.Data, r.Rules, r.Lang)
	if len(errors) > 0 {
		return map[string]validation.Errors{"validationError": errors}
	}

	return nil
}
