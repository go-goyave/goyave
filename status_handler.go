package goyave

import (
	"net/http"

	"goyave.dev/goyave/v4/validation"
)

// StatusHandler is a regular handler executed during the finalization step of the request's lifecycle
// if the response body is empty but a status code has been set.
// Status handlers are mainly used to implement a custom behavior for user or server errors (400 and 500 status codes).
type StatusHandler interface {
	Composable
	Handle(response *ResponseV5, request *RequestV5)
}

// PanicStatusHandler for the HTTP 500 error.
// If debugging is enabled, writes the error details to the response and
// print stacktrace in the console.
// If debugging is not enabled, writes `{"error": "Internal Server Error"}`
// to the response.
type PanicStatusHandlerV5 struct {
	Component
}

// Handle internal server error responses.
func (*PanicStatusHandlerV5) Handle(response *ResponseV5, request *RequestV5) {
	response.error(request.Extra[ExtraError])
	if response.IsEmpty() {
		message := map[string]string{
			"error": http.StatusText(response.GetStatus()),
		}
		response.JSON(response.GetStatus(), message)
	}
}

// ErrorStatusHandler a generic status handler for non-success codes.
// Writes the corresponding status message to the response.
type ErrorStatusHandlerV5 struct {
	Component
}

// Handle generic error reponses.
func (*ErrorStatusHandlerV5) Handle(response *ResponseV5, _ *RequestV5) {
	message := map[string]string{
		"error": http.StatusText(response.GetStatus()),
	}
	response.JSON(response.GetStatus(), message)
}

// ValidationStatusHandler for HTTP 422 errors.
// Writes the validation errors to the response.
type ValidationStatusHandlerV5 struct {
	Component
}

// Handle validation error responses.
func (*ValidationStatusHandlerV5) Handle(response *ResponseV5, request *RequestV5) {
	type ValidationErrorResponse struct {
		Body  *validation.Errors `json:"body,omitempty"`
		Query *validation.Errors `json:"query,omitempty"`
	}

	errs := &ValidationErrorResponse{}

	if e, ok := request.Extra[ExtraValidationError]; ok {
		errs.Body = e.(*validation.Errors)
	}

	if e, ok := request.Extra[ExtraQueryValidationError]; ok {
		errs.Query = e.(*validation.Errors)
	}

	message := map[string]*ValidationErrorResponse{"error": errs}
	response.JSON(response.GetStatus(), message)
}
