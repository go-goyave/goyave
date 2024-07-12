package goyave

import (
	"fmt"
	"net/http"

	"goyave.dev/goyave/v5/validation"
)

// StatusHandler is a regular handler executed during the finalization step of the request's lifecycle
// if the response body is empty but a status code has been set.
// Status handlers are mainly used to implement a custom behavior for user or server errors (400 and 500 status codes).
type StatusHandler interface {
	Composable
	Handle(response *Response, request *Request)
}

// PanicStatusHandler for the HTTP 500 error.
// If debugging is enabled, writes the error details to the response and
// print stacktrace in the console.
// If debugging is not enabled, writes `{"error": "Internal Server Error"}`
// to the response.
type PanicStatusHandler struct {
	Component
}

// Handle internal server error responses.
func (*PanicStatusHandler) Handle(response *Response, _ *Request) {
	response.error(response.GetError())
	if response.IsEmpty() && !response.Hijacked() {
		message := map[string]string{
			"error": http.StatusText(response.GetStatus()),
		}
		response.JSON(response.GetStatus(), message)
	}
}

// ErrorStatusHandler a generic status handler for non-success codes.
// Writes the corresponding status message to the response.
type ErrorStatusHandler struct {
	Component
}

// Handle generic error responses.
func (*ErrorStatusHandler) Handle(response *Response, _ *Request) {
	message := map[string]string{
		"error": http.StatusText(response.GetStatus()),
	}
	response.JSON(response.GetStatus(), message)
}

// RequestErrorStatusHandler a generic (error) status handler for requests.
type RequestErrorStatusHandler struct {
	Component
}

// Handle generic request (error) responses.
func (h *RequestErrorStatusHandler) Handle(response *Response, request *Request) {
	var errorMessages []string
	var langName string

	if request.Lang == nil {
		langName = h.Lang().GetDefault().Name()
	} else {
		langName = request.Lang.Name()
	}
	lang := h.Lang().GetLanguage(langName)

	if e, ok := request.Extra[ExtraRequestError{}]; ok {
		switch v := e.(type) {
		case []error:
			for _, err := range v {
				errorMessages = append(errorMessages, lang.Get(err.Error()))
			}
		case error:
			errorMessages = append(errorMessages, lang.Get(v.Error()))
		case string:
			errorMessages = append(errorMessages, lang.Get(v))
		default:
			errorMessages = append(errorMessages, lang.Get(fmt.Sprintf("%v", v)))
		}
	}

	messages := map[string][]string{
		"error": errorMessages,
	}
	response.JSON(response.GetStatus(), messages)
}

// ValidationStatusHandler for HTTP 422 errors.
// Writes the validation errors to the response.
type ValidationStatusHandler struct {
	Component
}

// Handle validation error responses.
func (*ValidationStatusHandler) Handle(response *Response, request *Request) {
	errs := &validation.ErrorResponse{}

	if e, ok := request.Extra[ExtraValidationError{}]; ok {
		errs.Body = e.(*validation.Errors)
	}

	if e, ok := request.Extra[ExtraQueryValidationError{}]; ok {
		errs.Query = e.(*validation.Errors)
	}

	message := map[string]*validation.ErrorResponse{"error": errs}
	response.JSON(response.GetStatus(), message)
}
