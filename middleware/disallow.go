package middleware

import (
	"net/http"

	"github.com/System-Glitch/goyave/lang"

	"github.com/System-Glitch/goyave"
	"github.com/System-Glitch/goyave/validation"
)

// DisallowNonValidatedFields validates that all fields in the request
// are validated by the RuleSet.
// Returns "422 Unprocessable Entity" and an error message if the user
// has sent non-validated field(s).
// If the body parsing failed, this middleware immediately passes to the next handler.
func DisallowNonValidatedFields(next goyave.Handler) goyave.Handler {
	return func(response *goyave.Response, request *goyave.Request) {
		nonValidated := validation.Errors{}
		if request.Data != nil {
			for field := range request.Data {
				if _, exists := request.Rules[field]; !exists {
					nonValidated[field] = append(nonValidated[field], lang.Get(request.Lang, "disallow-non-validated-fields"))
				}
			}

			if len(nonValidated) != 0 {
				response.JSON(http.StatusUnprocessableEntity, map[string]validation.Errors{"validationError": nonValidated})
				return
			}
		}
		next(response, request)
	}
}
