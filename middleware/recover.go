package middleware

import (
	"net/http"

	"github.com/System-Glitch/goyave/helpers/response"
)

// Recovery is a middleware that recovers from panic and sends a 500 error code.
// If debugging is enabled in the config, the error is also written in the response.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				response.Error(w, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
