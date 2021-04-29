package auth

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/gorm"
	"goyave.dev/goyave/v3"
	"goyave.dev/goyave/v3/config"
	"goyave.dev/goyave/v3/database"
	"goyave.dev/goyave/v3/lang"
)

func init() {
	config.Register("auth.jwt.expiry", config.Entry{
		Value:            300,
		Type:             reflect.Int,
		IsSlice:          false,
		AuthorizedValues: []interface{}{},
	})
	registerKeyConfigEntry("auth.jwt.secret")
	registerKeyConfigEntry("auth.jwt.rsa.public")
	registerKeyConfigEntry("auth.jwt.rsa.private")
	registerKeyConfigEntry("auth.jwt.rsa.password")
	registerKeyConfigEntry("auth.jwt.ecdsa.public")
	registerKeyConfigEntry("auth.jwt.ecdsa.private")
}

func registerKeyConfigEntry(name string) {
	config.Register(name, config.Entry{
		Value:            nil,
		Type:             reflect.String,
		IsSlice:          false,
		AuthorizedValues: []interface{}{},
	})
}

// GenerateToken generate a new JWT.
// The token is created using the HMAC SHA256 method and signed using
// the `auth.jwt.secret` config entry.
// The token is set to expire in the amount of seconds defined by
// the `auth.jwt.expiry` config entry.
//
// The generated token will contain the following claims:
// - `userid`: has the value of the `id` parameter
// - `nbf`: "Not before", the current timestamp is used
// - `exp`: "Expiry", the current timestamp plus the `auth.jwt.expiry` config entry.
func GenerateToken(username interface{}) (string, error) {
	return GenerateTokenWithClaims(jwt.MapClaims{"userid": username}, jwt.SigningMethodHS256)
}

// GenerateTokenWithClaims generates a new JWT with custom claims.
// The token is set to expire in the amount of seconds defined by
// the `auth.jwt.expiry` config entry.
// Depending on the given signing method, the following configuration entries
// will be used:
// - RSA:
//   - `auth.jwt.rsa.private`: path to the private PEM-encoded RSA key.
//   - `auth.jwt.rsa.password`: optional password for the private RSA key.
// - ECDSA: `auth.jwt.ecdsa.private`: path to the private PEM-encoded ECDSA key.
// - HMAC: `auth.jwt.secret`: HMAC secret
//
// The generated token will also contain the following claims:
// - `nbf`: "Not before", the current timestamp is used
// - `exp`: "Expiry", the current timestamp plus the `auth.jwt.expiry` config entry.
//
// `nbf` and `exp` can be overridden if they are set in the `claims` parameter.
func GenerateTokenWithClaims(claims jwt.MapClaims, signingMethod jwt.SigningMethod) (string, error) {
	expiry := time.Duration(config.GetInt("auth.jwt.expiry")) * time.Second
	now := time.Now()
	customClaims := jwt.MapClaims{
		"nbf": now.Unix(),             // Not Before
		"exp": now.Add(expiry).Unix(), // Expiry
	}
	for k, c := range claims {
		customClaims[k] = c
	}
	token := jwt.NewWithClaims(signingMethod, customClaims)

	key, err := getKey(signingMethod)
	if err != nil {
		panic(err)
	}
	return token.SignedString(key)
}

func getKey(signingMethod jwt.SigningMethod) (interface{}, error) {
	switch signingMethod.(type) {
	case *jwt.SigningMethodRSA:
		return loadKey("auth.jwt.rsa.private")
	case *jwt.SigningMethodECDSA:
		return loadKey("auth.jwt.ecdsa.private")
	case *jwt.SigningMethodHMAC:
		return []byte(config.GetString("auth.jwt.secret")), nil
	default:
		return nil, errors.New("Unsupported JWT signing method: " + signingMethod.Alg())
	}
}

// JWTAuthenticator implementation of Authenticator using a JSON Web Token.
type JWTAuthenticator struct {

	// SigningMethod expected by this authenticator when parsing JWT.
	// Defaults to HMAC.
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
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		key, err := loadKey("auth.jwt.rsa.public")
		if err != nil {
			panic(err)
		}
		return key, nil
	case *jwt.SigningMethodECDSA:
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		key, err := loadKey("auth.jwt.ecdsa.public")
		if err != nil {
			panic(err)
		}
		return key, nil
	case *jwt.SigningMethodHMAC, nil:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.GetString("auth.jwt.secret")), nil
	default:
		panic(errors.New("Unsupported JWT Signing method: " + a.SigningMethod.Alg()))
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
