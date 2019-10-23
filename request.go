package goyave

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/System-Glitch/goyave/config"
	"github.com/System-Glitch/goyave/helpers/filesystem"

	"github.com/System-Glitch/govalidator"
)

// Request struct represents an http request.
// Contains the validated body in the Data attribute if the route was defined with a request generator function
type Request struct {
	httpRequest *http.Request
	cookies     []*http.Cookie
	Rules       govalidator.MapData
	Data        map[string]interface{}
	Params      map[string]string
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

func (r *Request) validate() map[string]interface{} {
	if r.Rules == nil {
		return nil
	}

	r.Data = make(map[string]interface{}, 0)
	validator := govalidator.New(govalidator.Options{
		Request:         r.httpRequest,
		Rules:           r.Rules,
		Data:            &r.Data,
		RequiredDefault: false,
	})

	var errors url.Values
	if r.httpRequest.Header.Get("Content-Type") == "application/json" {
		errors = validator.ValidateJSON()
	} else {
		errors = validator.Validate()
		err := r.httpRequest.ParseMultipartForm(int64(config.Get("maxUploadSize").(float64)) << 20)
		if err != nil {
			panic(err)
		}
		if len(errors) == 0 {
			r.Data = generateFlatMap(r.httpRequest, r.Rules)
		}
	}
	if len(errors) > 0 {
		return map[string]interface{}{"validationError": errors}
	}

	return nil
}

func generateFlatMap(request *http.Request, rules govalidator.MapData) map[string]interface{} {
	var flatMap map[string]interface{} = make(map[string]interface{})
	for field, value := range request.MultipartForm.Value {
		flatMap[field] = value[0]
	}

	for field := range rules {
		if strings.HasPrefix(field, "file:") {
			name := field[5:]
			f, h, err := request.FormFile(name)
			if err != nil {
				panic(err)
			}
			file := filesystem.File{
				Header: h,
				Data:   f,
			}
			flatMap[name] = file
		}
	}
	return flatMap
}

// isRequestMalformed checks if the only error in the given errsBag is a unmarshal error.
// Used to determine if a 400 Bad Request or 422 Unprocessable Entity should be returned.
func isRequestMalformed(errsBag map[string]interface{}) bool {
	_, ok := errsBag["validationError"].(url.Values)["_error"]
	return ok
}
