package goyave

import (
	"net/http"
	"net/url"

	"github.com/System-Glitch/govalidator"
)

// Request struct represents an http request.
// Contains the validated body in the Data attribute if the route was defined with a request generator function
type Request struct {
	httpRequest *http.Request
	Rules       govalidator.MapData
	Data        map[string]interface{}
	Params      map[string]string
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
		r.httpRequest.ParseForm()
		r.Data = generateFlatMap(r.httpRequest.Form)
	}
	if len(errors) > 0 {
		return map[string]interface{}{"validationError": errors}
	}

	return nil
}

func generateFlatMap(values url.Values) map[string]interface{} {
	var flatMap map[string]interface{} = make(map[string]interface{})
	for field, value := range values {
		flatMap[field] = value[0]
	}
	return flatMap
}

// isRequestMalformed checks if the only error in the given errsBag is a unmarshal error.
// Used to determine if a 400 Bad Request or 422 Unprocessable Entity should be returned.
func isRequestMalformed(errsBag map[string]interface{}) bool {
	_, ok := errsBag["validationError"].(url.Values)["_error"]
	return ok
}
