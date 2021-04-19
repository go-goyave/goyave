package auth

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/gorm"
	"goyave.dev/goyave/v3"
	"goyave.dev/goyave/v3/config"
	"goyave.dev/goyave/v3/database"
	"goyave.dev/goyave/v3/lang"
)

// JWTAuthenticator implementation of Authenticator using a JSON Web Token.
type JWTAuthenticator struct {
	SigningMethod jwt.SigningMethod

	// ClaimName the name of the claim used to retrieve the user.
	// Defaults to "userid".
	ClaimName string

	// Optional defines if the authenticator allows requests that
	// don't provide credentials. Handlers should therefore check
	// if request.User is not nil before accessing it.
	Optional bool
}

var _ Authenticator = (*JWTAuthenticator)(nil) // implements Authenticator

func init() {
	config.Register("auth.jwt.secret", config.Entry{
		Value:            nil,
		Type:             reflect.String,
		IsSlice:          false,
		AuthorizedValues: []interface{}{},
	})
	config.Register("auth.jwt.expiry", config.Entry{
		Value:            300,
		Type:             reflect.Int,
		IsSlice:          false,
		AuthorizedValues: []interface{}{},
	})
}

// Authenticate fetch the user corresponding to the token
// found in the given request and puts the result in the given user pointer.
// If no user can be authenticated, returns an error.
//
// The database request is executed based on the model name and the
// struct tag `auth:"username"`.
//
// If the token is valid and has claims, those claims will be added to `request.Extra` with the key "jwt_claims".
//
// This implementation is a JWT-based authentication using HMAC SHA256, supporting only one active token.
func (a *JWTAuthenticator) Authenticate(request *goyave.Request, user interface{}) error {

	tokenString, ok := request.BearerToken()
	if tokenString == "" || !ok {
		if a.Optional {
			return nil
		}
		return fmt.Errorf(lang.Get(request.Lang, "auth.no-credentials-provided"))
	}

	token, err := jwt.Parse(tokenString, a.keyFunc)

	if err == nil && token.Valid {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			request.Extra["jwt_claims"] = claims
			column := FindColumns(user, "username")[0]
			claimName := a.ClaimName
			if claimName == "" {
				claimName = "userid"
			}
			result := database.GetConnection().Where(column.Name+" = ?", claims[claimName]).First(user)

			if result.Error != nil {
				if errors.Is(result.Error, gorm.ErrRecordNotFound) {
					return fmt.Errorf(lang.Get(request.Lang, "auth.invalid-credentials"))
				}
				panic(result.Error)
			}

			return nil
		}
	}

	return a.makeError(request.Lang, err.(*jwt.ValidationError).Errors)
}

func (a *JWTAuthenticator) keyFunc(token *jwt.Token) (interface{}, error) {
	switch a.SigningMethod.(type) {
	case *jwt.SigningMethodRSA:
		// TODO avoid redundancy
		// TODO cache keys to avoid re-parsing all the time
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		data, _ := ioutil.ReadFile(config.GetString("auth.jwt.rsa.public"))
		return jwt.ParseRSAPublicKeyFromPEM(data)
	case *jwt.SigningMethodECDSA:
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		data, _ := ioutil.ReadFile(config.GetString("auth.jwt.ecdsa.public"))
		return jwt.ParseECPublicKeyFromPEM(data)
	case *jwt.SigningMethodHMAC:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.GetString("auth.jwt.secret")), nil
	default:
		return nil, errors.New("Unsupported JWT Signing method: " + a.SigningMethod.Alg())
	}
}

func (a *JWTAuthenticator) makeError(language string, bitfield uint32) error {
	if bitfield&jwt.ValidationErrorNotValidYet != 0 {
		return fmt.Errorf(lang.Get(language, "auth.jwt-not-valid-yet"))
	} else if bitfield&jwt.ValidationErrorExpired != 0 {
		return fmt.Errorf(lang.Get(language, "auth.jwt-expired"))
	}
	return fmt.Errorf(lang.Get(language, "auth.jwt-invalid"))
}
