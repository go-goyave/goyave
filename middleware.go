package goyave

import (
	"net/http"
)

// Middleware function generating middleware handler function
type Middleware func(Handler) Handler

// recoveryMiddleware is a middleware that recovers from panic and sends a 500 error code.
// If debugging is enabled in the config, the error is also written in the response.
func recoveryMiddleware(next Handler) Handler {
	return func(response Response, r *Request) {
		defer func() {
			if err := recover(); err != nil {
				response.Error(err)
			}
		}()

		next(response, r)
	}
}

// validateRequestMiddleware is a middleware that validates the request and sends a 422 error code
// if the validation rules are not met.
func validateRequestMiddleware(next Handler) Handler {
	return func(response Response, r *Request) {
		errsBag := r.validate()
		if errsBag == nil {
			next(response, r)
			return
		}

		var code int
		if isRequestMalformed(errsBag) {
			code = http.StatusBadRequest
		} else {
			code = http.StatusUnprocessableEntity
		}
		response.JSON(code, errsBag)
	}
}
