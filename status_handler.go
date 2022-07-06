package goyave

import "net/http"

// TODO status handler should also have access to server resources (config, lang, etc)

type StatusHandler interface {
	IController
	Handle(response *ResponseV5, request *RequestV5)
}

type PanicStatusHandlerV5 struct {
	Controller
}

// PanicStatusHandler for the HTTP 500 error.
// If debugging is enabled, writes the error details to the response and
// print stacktrace in the console.
// If debugging is not enabled, writes `{"error": "Internal Server Error"}`
// to the response.
func (*PanicStatusHandlerV5) Handle(response *ResponseV5, request *RequestV5) {
	response.error(request.Extra[ExtraError])
	if response.IsEmpty() {
		message := map[string]string{
			"error": http.StatusText(response.GetStatus()),
		}
		response.JSON(response.GetStatus(), message)
	}
}

type ErrorStatusHandlerV5 struct {
	Controller
}

// ErrorStatusHandler a generic status handler for non-success codes.
// Writes the corresponding status message to the response.
func (*ErrorStatusHandlerV5) Handle(response *ResponseV5, request *RequestV5) {
	message := map[string]string{
		"error": http.StatusText(response.GetStatus()),
	}
	response.JSON(response.GetStatus(), message)
}

type ValidationStatusHandlerV5 struct {
	Controller
}

// ValidationStatusHandler for HTTP 400 and HTTP 422 errors.
// Writes the validation errors to the response.
func (*ValidationStatusHandlerV5) Handle(response *ResponseV5, request *RequestV5) {
	message := map[string]any{"validationError": request.Extra[ExtraError]}
	response.JSON(response.GetStatus(), message)
}
