package auth

import (
	"fmt"

	"github.com/System-Glitch/goyave/v2"
	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/database"
	"github.com/System-Glitch/goyave/v2/lang"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
)

// JWTAuthenticator implementation of Authenticator using a simple Bearer token.
type JWTAuthenticator struct{}

var _ Authenticator = (*JWTAuthenticator)(nil) // implements Authenticator

// Authenticate fetch the user corresponding to the token
// found in the given request and puts the result in the given user pointer.
// If no user can be authenticated, returns false.
//
// The database request is executed based on the model name and the
// struct tag `auth:"username"`.
//
// This implementation is a JWT-based authentication using HMAC SHA256, supporting only one active token.
func (a *JWTAuthenticator) Authenticate(request *goyave.Request, user interface{}) error {

	tokenString := GetBearerToken(request)
	if tokenString == "" {
		return fmt.Errorf(lang.Get(request.Lang, "no-credentials-provided"))
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(config.GetString("jwtSecret")), nil
	})

	if err == nil && token.Valid {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			column := FindColumns(user, "username")[0]
			result := database.GetConnection().Where(column.Name+" = ?", claims["userid"]).First(user)

			if errors := result.GetErrors(); len(errors) != 0 && !gorm.IsRecordNotFoundError(result.Error) {
				panic(errors)
			}

			if result.RecordNotFound() {
				return fmt.Errorf(lang.Get(request.Lang, "invalid-credentials"))
			}

			return nil
		}
	}

	return a.makeError(request.Lang, err.(*jwt.ValidationError).Errors)
}

func (a *JWTAuthenticator) makeError(language string, bitfield uint32) error {
	if bitfield&jwt.ValidationErrorNotValidYet != 0 {
		return fmt.Errorf(lang.Get(language, "jwt-not-valid-yet"))
	} else if bitfield&jwt.ValidationErrorExpired != 0 {
		return fmt.Errorf(lang.Get(language, "jwt-expired"))
	}
	return fmt.Errorf(lang.Get(language, "jwt-invalid"))
}
