package middleware

import (
	"net/http"

	"github.com/System-Glitch/goyave"
	"github.com/System-Glitch/goyave/validation"
)

// DisallowNonValidatedFields validates that all fields in the request
// are validated by the RuleSet.
// Returns "422 Unprocessable Entity" and an error message if the user
// has sent non-validated field(s).
func DisallowNonValidatedFields(next goyave.Handler) goyave.Handler {
	return func(response goyave.Response, request *goyave.Request) {
		nonValidated := validation.Errors{}
		for field := range request.Data {
			if _, exists := request.Rules[field]; !exists {
				nonValidated[field] = append(nonValidated[field], "Non-validated fields are forbidden.") // TODO get lang message
			}
		}

		if len(nonValidated) != 0 {
			response.JSON(http.StatusUnprocessableEntity, map[string]validation.Errors{"validationError": nonValidated})
		} else {
			next(response, request)
		}
	}
}
