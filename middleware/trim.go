package middleware

import (
	"strings"

	"github.com/System-Glitch/goyave"
)

// Trim removes all leading and trailing white space from string fields.
func Trim(next goyave.Handler) goyave.Handler {
	return func(response *goyave.Response, request *goyave.Request) {
		if request.Data != nil {
			for field, val := range request.Data {
				str, ok := val.(string)
				if ok {
					request.Data[field] = strings.TrimSpace(str)
				}
			}
		}
		next(response, request)
	}
}
