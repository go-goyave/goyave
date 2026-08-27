package auth

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/lang"
	"goyave.dev/goyave/v5/validation"

	errorutil "goyave.dev/goyave/v5/util/errors"
)

// TODO feature support for golang-jwt/jwt/v5

const (
	// JWTServiceName identifier for the `JWTService`.
	JWTServiceName = "goyave.jwt"
)

// ExtraJWTClaims when using the built-in `JWTAuthenticator`, this
// key can be used to retrieve the JWT claims in the request's `Extra`.
type ExtraJWTClaims struct{}

// TODO JWT config struct
type JWTConfig struct {
	// Expiry defined the number of seconds a token is valid for.
	// Defaults to 300s.
	Expiry int

	// Secret the secret value used for HMAC signatures.
	Secret string

	RSA   KeyPairConfig
	ECDSA KeyPairConfig
}

func (c JWTConfig) RuleSet() validation.RuleSet {
	return validation.RuleSet{
		{Path: validation.CurrentElement, Rules: validation.List{validation.Object()}},
		{Path: "Expiry", Rules: validation.List{validation.Required(), validation.Int(), validation.Min(0)}},
		{Path: "Secret", Rules: validation.List{validation.Required(), validation.String()}},
		{Path: "RSA", Rules: c.RSA.RuleSet()},
		{Path: "ECDSA", Rules: c.ECDSA.RuleSet()},
	}
}

func (c JWTConfig) Default() JWTConfig {
	return JWTConfig{
		Expiry: 300,
	}
}

type KeyPairConfig struct {
	// Public the path to the file containing the PEM-encoded public key
	// in the JWT service filesystem.
	Public string
	// Private the path to the file containing the PEM-encoded private key
	// in the JWT service filesystem.
	Private string
}

func (KeyPairConfig) RuleSet() validation.RuleSet {
	return validation.RuleSet{
		{Path: validation.CurrentElement, Rules: validation.List{validation.Object()}},
		{Path: "Public", Rules: validation.List{validation.Required(), validation.String()}},
		{Path: "Private", Rules: validation.List{validation.Required(), validation.String()}},
	}
}

// JWTService providing signature keys cache and JWT generation.
//
// This service is identified by `auth.JWTServiceName`.
type JWTService struct {
	fs     fs.FS
	config *JWTConfig
	cache  sync.Map
}

// NewJWTService create a new `JWTService` with the given config and file system.
// The file system is used to get the signing keys.
func NewJWTService(config *JWTConfig, fs fs.FS) *JWTService {
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
// the `Secret` config entry.
// The token is set to expire in the amount of seconds defined by
// the `Expiry` config entry.
//
// The generated token will contain the following claims:
//   - `sub`: has the value of the `id` parameter
//   - `nbf`: "Not before", the current timestamp is used
//   - `exp`: "Expiry", the current timestamp plus the `Expiry` config entry.
func (s *JWTService) GenerateToken(username any) (string, error) {
	return s.GenerateTokenWithClaims(jwt.MapClaims{"sub": username}, jwt.SigningMethodHS256)
}

// GenerateTokenWithClaims generates a new JWT with custom claims.
// The token is set to expire in the amount of seconds defined by
// the configuration. See [JWTConfig].
//
// The generated token will also contain the following claims:
//   - `nbf`: "Not before", the current timestamp is used
//   - `exp`: "Expiry", the current timestamp plus the `Expiry` config entry.
//
// `nbf` and `exp` can be overridden if they are set in the `claims` parameter.
func (s *JWTService) GenerateTokenWithClaims(claims jwt.MapClaims, signingMethod jwt.SigningMethod) (string, error) {
	exp := time.Duration(s.config.Expiry) * time.Second
	now := time.Now()
	customClaims := jwt.MapClaims{
		"nbf": now.Unix(),          // Not Before
		"exp": now.Add(exp).Unix(), // Expiry
	}
	maps.Copy(customClaims, claims)
	token := jwt.NewWithClaims(signingMethod, customClaims)

	key, err := s.GetPrivateKey(signingMethod)
	if err != nil {
		return "", err
	}
	result, err := token.SignedString(key)
	return result, errorutil.New(err)
}

// getKey read the file from the service's filesystem and returns the raw data.
// The returned data isn't usable directly, it needs to be parsed first. Parsing
// depends on the signing method.
// The key is expected to be PEM-encoded.
//
// To optimize subsequent requests and avoid IO for keys that are stored on the
// disk, the keys are cached.
func (s *JWTService) getKey(filePath string) ([]byte, error) {
	if k, ok := s.cache.Load(filePath); ok {
		return k.([]byte), nil
	}

	key, err := fs.ReadFile(s.fs, filePath)
	if err != nil {
		return nil, errorutil.New(err)
	}

	if err == nil {
		s.cache.Store(filePath, key)
	}
	return key, errorutil.New(err)
}

// GetPrivateKey loads the private key that corresponds to the given `signingMethod`.
func (s *JWTService) GetPrivateKey(signingMethod jwt.SigningMethod) (any, error) {
	switch signingMethod.(type) {
	case *jwt.SigningMethodRSA:
		key, err := s.getKey(s.config.RSA.Private)
		if err != nil {
			return nil, errorutil.New(err)
		}
		return jwt.ParseRSAPrivateKeyFromPEM(key)
	case *jwt.SigningMethodECDSA:
		key, err := s.getKey(s.config.ECDSA.Private)
		if err != nil {
			return nil, errorutil.New(err)
		}
		return jwt.ParseECPrivateKeyFromPEM(key)
	case *jwt.SigningMethodHMAC:
		return []byte(s.config.Secret), nil
	default:
		return nil, errorutil.New("unsupported JWT signing method: " + signingMethod.Alg())
	}
}

// GetPublicKey loads the public key that corresponds to the given `signingMethod`.
func (s *JWTService) GetPublicKey(signingMethod jwt.SigningMethod) (any, error) {
	switch signingMethod.(type) {
	case *jwt.SigningMethodRSA:
		key, err := s.getKey(s.config.RSA.Public)
		if err != nil {
			return nil, errorutil.New(err)
		}
		return jwt.ParseRSAPublicKeyFromPEM(key)
	case *jwt.SigningMethodECDSA:
		key, err := s.getKey(s.config.ECDSA.Public)
		if err != nil {
			return nil, errorutil.New(err)
		}
		return jwt.ParseECPublicKeyFromPEM(key)
	case *jwt.SigningMethodHMAC:
		return []byte(s.config.Secret), nil
	default:
		return nil, errorutil.New("unsupported JWT signing method: " + signingMethod.Alg())
	}
}

func (s *JWTService) keyFunc(signingMethod jwt.SigningMethod) jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		switch signingMethod.(type) {
		case *jwt.SigningMethodRSA:
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
		case *jwt.SigningMethodECDSA:
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
		case *jwt.SigningMethodHMAC, nil:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
		default:
			return nil, errorutil.New("unsupported JWT Signing method: " + signingMethod.Alg())
		}
		return s.GetPublicKey(signingMethod)
	}
}

// Parse a JWT in string form using the given signingMethod. If the JWT isn't signed with the same
// method, an error is returned.
//
// Note that only errors of type `*errors.Error` returned by this function should be considered as
// system errors. Other error types simply represent user errors (invalid token, unexpected signing method, ...)
func (s *JWTService) Parse(tokenString string, signingMethod jwt.SigningMethod) (*jwt.Token, error) {
	return jwt.Parse(tokenString, s.keyFunc(signingMethod))
}

// JWTAuthenticator implementation of Authenticator using a JSON Web Token.
//
// The T parameter represents the user DTO and should not be a pointer.
type JWTAuthenticator[T any] struct {
	JWTService *JWTService // TODO use an interface for the JWT service

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
func NewJWTAuthenticator[T any](jwtService *JWTService, userService UserService[T]) *JWTAuthenticator[T] {
	return &JWTAuthenticator[T]{
		JWTService:  jwtService,
		UserService: userService,
	}
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

	token, err := jwt.Parse(tokenString, a.JWTService.keyFunc(a.SigningMethod))

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
				return nil, errorutil.New(err)
			}

			return user, nil
		}
	}

	return nil, a.makeError(request.Lang, err)
}

func (a *JWTAuthenticator[T]) makeError(language *lang.Language, err error) error {
	if _, ok := err.(*errorutil.Error); ok { // System error
		return err
	}
	switch {
	case errors.Is(err, jwt.ErrTokenNotValidYet):
		return fmt.Errorf("%s", language.Get("auth.jwt-not-valid-yet"))
	case errors.Is(err, jwt.ErrTokenExpired):
		return fmt.Errorf("%s", language.Get("auth.jwt-expired"))
	default:
		return fmt.Errorf("%s", language.Get("auth.jwt-invalid"))
	}
}

func (a *JWTAuthenticator[T]) Scheme() string {
	return "Bearer"
}
