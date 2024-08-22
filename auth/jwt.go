package auth

import (
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/lang"

	errorutil "goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/fsutil/osfs"
)

const (
	// JWTServiceName identifier for the `JWTService`.
	JWTServiceName = "goyave.jwt"
)

// ExtraJWTClaims when using the built-in `JWTAuthenticator`, this
// key can be used to retrieve the JWT claims in the request's `Extra`.
type ExtraJWTClaims struct{}

func init() {
	config.Register("auth.jwt.expiry", config.Entry{
		Value:            300,
		Type:             reflect.Int,
		IsSlice:          false,
		AuthorizedValues: []any{},
	})
	registerKeyConfigEntry("auth.jwt.secret")
	registerKeyConfigEntry("auth.jwt.rsa.public")
	registerKeyConfigEntry("auth.jwt.rsa.private")
	registerKeyConfigEntry("auth.jwt.ecdsa.public")
	registerKeyConfigEntry("auth.jwt.ecdsa.private")
}

func registerKeyConfigEntry(name string) {
	config.Register(name, config.Entry{
		Value:            nil,
		Type:             reflect.String,
		IsSlice:          false,
		AuthorizedValues: []any{},
	})
}

// JWTService providing signature keys cache and JWT generation.
//
// This service is identified by `auth.JWTServiceName`.
type JWTService struct {
	fs     fs.FS
	config *config.Config
	cache  sync.Map
}

// NewJWTService create a new `JWTService` with the given config and file system.
// The file system is used to get the signing keys.
func NewJWTService(config *config.Config, fs fs.FS) *JWTService {
	return &JWTService{
		config: config,
		fs:     fs,
	}
}

// Name returns the name of the service.
func (s *JWTService) Name() string {
	return JWTServiceName
}

// GenerateToken generate a new JWT.
// The token is created using the HMAC SHA256 method and signed using
// the `auth.jwt.secret` config entry.
// The token is set to expire in the amount of seconds defined by
// the `auth.jwt.expiry` config entry.
//
// The generated token will contain the following claims:
//   - `sub`: has the value of the `id` parameter
//   - `nbf`: "Not before", the current timestamp is used
//   - `exp`: "Expiry", the current timestamp plus the `auth.jwt.expiry` config entry.
func (s *JWTService) GenerateToken(username any) (string, error) {
	return s.GenerateTokenWithClaims(jwt.MapClaims{"sub": username}, jwt.SigningMethodHS256)
}

// GenerateTokenWithClaims generates a new JWT with custom claims.
// The token is set to expire in the amount of seconds defined by
// the `auth.jwt.expiry` config entry.
// Depending on the given signing method, the following configuration entries
// will be used:
//   - RSA: `auth.jwt.rsa.private`: path to the private PEM-encoded RSA key.
//   - ECDSA: `auth.jwt.ecdsa.private`: path to the private PEM-encoded ECDSA key.
//   - HMAC: `auth.jwt.secret`: HMAC secret
//
// The generated token will also contain the following claims:
//   - `nbf`: "Not before", the current timestamp is used
//   - `exp`: "Expiry", the current timestamp plus the `auth.jwt.expiry` config entry.
//
// `nbf` and `exp` can be overridden if they are set in the `claims` parameter.
func (s *JWTService) GenerateTokenWithClaims(claims jwt.MapClaims, signingMethod jwt.SigningMethod) (string, error) {
	exp := time.Duration(s.config.GetInt("auth.jwt.expiry")) * time.Second
	now := time.Now()
	customClaims := jwt.MapClaims{
		"nbf": now.Unix(),          // Not Before
		"exp": now.Add(exp).Unix(), // Expiry
	}
	for k, c := range claims {
		customClaims[k] = c
	}
	token := jwt.NewWithClaims(signingMethod, customClaims)

	key, err := s.GetPrivateKey(signingMethod)
	if err != nil {
		return "", err
	}
	result, err := token.SignedString(key)
	return result, errorutil.New(err)
}

// GetKey load a JWT signature key from the config.
// List of `entry` parameter possible values:
//
//   - `auth.jwt.rsa.public`
//   - `auth.jwt.rsa.private`
//   - `auth.jwt.ecdsa.public`
//   - `auth.jwt.ecdsa.private`
//   - `auth.jwt.secret`
//
// To optimize subsequent requests and avoid IO for keys that are stored on the
// disk, the keys are cached.
func (s *JWTService) GetKey(entry string) (any, error) {
	if k, ok := s.cache.Load(entry); ok {
		return k, nil
	}

	data, err := fs.ReadFile(s.fs, s.config.GetString(entry))
	if err != nil {
		return nil, errorutil.New(err)
	}

	var key any
	switch entry {
	case "auth.jwt.rsa.private":
		key, err = jwt.ParseRSAPrivateKeyFromPEM(data)
	case "auth.jwt.rsa.public":
		key, err = jwt.ParseRSAPublicKeyFromPEM(data)
	case "auth.jwt.ecdsa.private":
		key, err = jwt.ParseECPrivateKeyFromPEM(data)
	case "auth.jwt.ecdsa.public":
		key, err = jwt.ParseECPublicKeyFromPEM(data)
	}

	if err == nil {
		s.cache.Store(entry, key)
	}
	return key, errorutil.New(err)
}

// GetPrivateKey loads the private key that corresponds to the given `signingMethod`.
func (s *JWTService) GetPrivateKey(signingMethod jwt.SigningMethod) (any, error) {
	switch signingMethod.(type) {
	case *jwt.SigningMethodRSA:
		return s.GetKey("auth.jwt.rsa.private")
	case *jwt.SigningMethodECDSA:
		return s.GetKey("auth.jwt.ecdsa.private")
	case *jwt.SigningMethodHMAC:
		return []byte(s.config.GetString("auth.jwt.secret")), nil
	default:
		return nil, errorutil.New("unsupported JWT signing method: " + signingMethod.Alg())
	}
}

// JWTAuthenticator implementation of Authenticator using a JSON Web Token.
//
// The T parameter represents the user DTO and should not be a pointer.
type JWTAuthenticator[T any] struct {
	goyave.Component

	service *JWTService

	UserService UserService[T]

	// SigningMethod expected by this authenticator when parsing JWT.
	// Defaults to HMAC.
	SigningMethod jwt.SigningMethod

	// ClaimName the name of the claim used to retrieve the user.
	// Defaults to "sub".
	ClaimName string

	// Optional defines if the authenticator allows requests that
	// don't provide credentials. Handlers should therefore check
	// if `request.User` is not `nil` before accessing it.
	Optional bool
}

// NewJWTAuthenticator create a new authenticator for the JSON Web Token authentication flow.
//
// The T parameter represents the user DTO and should not be a pointer.
func NewJWTAuthenticator[T any](userService UserService[T]) *JWTAuthenticator[T] {
	return &JWTAuthenticator[T]{
		UserService: userService,
	}
}

// Init the authenticator. Automatically registers the `JWTService` if not already registered,
// using `osfs.FS` as file system for the keys.
func (a *JWTAuthenticator[T]) Init(server *goyave.Server) {
	a.Component.Init(server)

	service, ok := server.LookupService(JWTServiceName)
	if !ok {
		service = NewJWTService(server.Config(), &osfs.FS{})
		server.RegisterService(service)
	}
	a.service = service.(*JWTService)
}

// Authenticate fetch the user corresponding to the token
// found in the given request and returns it.
// If no user can be authenticated, returns an error.
//
// If the token is valid and has claims, those claims will be added to `request.Extra` with the key "jwt_claims".
func (a *JWTAuthenticator[T]) Authenticate(request *goyave.Request) (*T, error) {
	tokenString, ok := request.BearerToken()
	if tokenString == "" || !ok {
		if a.Optional {
			return nil, nil
		}
		return nil, fmt.Errorf("%s", request.Lang.Get("auth.no-credentials-provided"))
	}

	token, err := jwt.Parse(tokenString, a.keyFunc)

	if err == nil && token.Valid {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			request.Extra[ExtraJWTClaims{}] = claims

			claimName := a.ClaimName
			if claimName == "" {
				claimName = "sub"
			}
			user, err := a.UserService.FindByUsername(request.Context(), claims[claimName])
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, fmt.Errorf("%s", request.Lang.Get("auth.invalid-credentials"))
				}
				panic(errorutil.New(err))
			}

			return user, nil
		}
	}

	return nil, a.makeError(request.Lang, err.(*jwt.ValidationError).Errors)
}

func (a *JWTAuthenticator[T]) keyFunc(token *jwt.Token) (any, error) {
	switch a.SigningMethod.(type) {
	case *jwt.SigningMethodRSA:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		key, err := a.service.GetKey("auth.jwt.rsa.public")
		if err != nil {
			panic(err)
		}
		return key, nil
	case *jwt.SigningMethodECDSA:
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		key, err := a.service.GetKey("auth.jwt.ecdsa.public")
		if err != nil {
			panic(err)
		}
		return key, nil
	case *jwt.SigningMethodHMAC, nil:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.Config().GetString("auth.jwt.secret")), nil
	default:
		panic(errorutil.New("unsupported JWT Signing method: " + a.SigningMethod.Alg()))
	}
}

func (a *JWTAuthenticator[T]) makeError(language *lang.Language, bitfield uint32) error {
	if bitfield&jwt.ValidationErrorNotValidYet != 0 {
		return fmt.Errorf("%s", language.Get("auth.jwt-not-valid-yet"))
	} else if bitfield&jwt.ValidationErrorExpired != 0 {
		return fmt.Errorf("%s", language.Get("auth.jwt-expired"))
	}
	return fmt.Errorf("%s", language.Get("auth.jwt-invalid"))
}
