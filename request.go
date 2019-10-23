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
