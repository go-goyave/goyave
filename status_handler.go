package goyave

import (
	"net/http"

	"goyave.dev/goyave/v4/validation"
)

type StatusHandler interface {
	IController
	Handle(response *ResponseV5, request *RequestV5)
}

// PanicStatusHandler for the HTTP 500 error.
// If debugging is enabled, writes the error details to the response and
// print stacktrace in the console.
// If debugging is not enabled, writes `{"error": "Internal Server Error"}`
// to the response.
type PanicStatusHandlerV5 struct {
	Controller
}

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
	Controller
}

func (*ErrorStatusHandlerV5) Handle(response *ResponseV5, request *RequestV5) {
	message := map[string]string{
		"error": http.StatusText(response.GetStatus()),
	}
	response.JSON(response.GetStatus(), message)
}

// ValidationStatusHandler for HTTP 422 errors.
// Writes the validation errors to the response.
type ValidationStatusHandlerV5 struct {
	Controller
}

func (*ValidationStatusHandlerV5) Handle(response *ResponseV5, request *RequestV5) {
	type ValidationErrorResponse struct {
		Body  *validation.ErrorsV5 `json:"body,omitempty"`
		Query *validation.ErrorsV5 `json:"query,omitempty"`
	}

	errs := &ValidationErrorResponse{}

	if e, ok := request.Extra[ExtraValidationError]; ok {
		errs.Body = e.(*validation.ErrorsV5)
	}

	if e, ok := request.Extra[ExtraQueryValidationError]; ok {
		errs.Query = e.(*validation.ErrorsV5)
	}

	message := map[string]*ValidationErrorResponse{"error": errs}
	response.JSON(response.GetStatus(), message)
}
