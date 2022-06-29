package goyave

import "net/http"

// PanicStatusHandler for the HTTP 500 error.
// If debugging is enabled, writes the error details to the response and
// print stacktrace in the console.
// If debugging is not enabled, writes `{"error": "Internal Server Error"}`
// to the response.
func PanicStatusHandlerV5(c *Context) {
	c.error(c.Extra[ExtraError])
	if c.empty {
		message := map[string]string{
			"error": http.StatusText(c.GetStatus()),
		}
		c.JSON(c.GetStatus(), message)
	}
}

// ErrorStatusHandler a generic status handler for non-success codes.
// Writes the corresponding status message to the response.
func ErrorStatusHandlerV5(c *Context) {
	message := map[string]string{
		"error": http.StatusText(c.GetStatus()),
	}
	c.JSON(c.GetStatus(), message)
}

// ValidationStatusHandler for HTTP 400 and HTTP 422 errors.
// Writes the validation errors to the response.
func ValidationStatusHandlerV5(c *Context) {
	message := map[string]any{"validationError": c.Extra[ExtraError]}
	c.JSON(c.GetStatus(), message)
}
