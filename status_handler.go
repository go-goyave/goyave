package goyave

import (
	"errors"
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

// ParseErrorStatusHandler a generic (error) status handler for requests.
type ParseErrorStatusHandler struct {
	Component
}

// Handle generic request (error) responses.
func (h *ParseErrorStatusHandler) Handle(response *Response, request *Request) {
	var errorMessage string
	lang := request.Lang

	err, ok := request.Extra[ExtraParseError{}].(error)
	if ok {
		switch {
		case errors.Is(err, ErrInvalidJSONBody):
			errorMessage = lang.Get("parse.json-invalid-body")
		case errors.Is(err, ErrInvalidQuery):
			errorMessage = lang.Get("parse.invalid-query")
		case errors.Is(err, ErrInvalidContentForType):
			errorMessage = lang.Get("parse.invalid-content-for-type")
		case errors.Is(err, ErrErrorInRequestBody):
			errorMessage = lang.Get("parse.error-in-request-body")
		default:
			errorMessage = lang.Get(err.Error())
		}
	} else {
		errorMessage = http.StatusText(response.GetStatus())
	}

	message := map[string]string{
		"error": errorMessage,
	}
	response.JSON(response.GetStatus(), message)
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
