package auth

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"reflect"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5"
	errorutil "goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/validation"
)

// BasicAuthenticator implementation of Authenticator with the Basic
// authentication method.
//
// The T parameter represents the user DTO and should not be a pointer. The DTO used should be
// different from the DTO returned to clients as a response because it needs to contain the user's password.
type BasicAuthenticator[T any] struct {
	UserService UserService[T]

	// PasswordField the name of T's struct field that holds the user's hashed password.
	// It will be used to compare the password hash with the user input.
	PasswordField string

	// Optional defines if the authenticator allows requests that
	// don't provide credentials. Handlers should therefore check
	// if `request.User` is not `nil` before accessing it.
	Optional bool
}

// NewBasicAuthenticator create a new authenticator for the Basic authentication flow.
//
// The T parameter represents the user DTO and should not be a pointer. The DTO used should be
// different from the DTO returned to clients as a response because it needs to contain the user's password.
//
// The `passwordField` corresponds to the name of T's struct field that holds the user's hashed password.
// It will be used to compare the password hash with the user input.
func NewBasicAuthenticator[T any](userService UserService[T], passwordField string) *BasicAuthenticator[T] {
	return &BasicAuthenticator[T]{
		UserService:   userService,
		PasswordField: passwordField,
	}
}

// Authenticate fetch the user corresponding to the credentials
// found in the given request and returns it.
// If no user can be authenticated, returns an error.
// The password is checked using bcrypt.
func (a *BasicAuthenticator[T]) Authenticate(request *goyave.Request) (*T, error) {
	username, password, ok := request.BasicAuth()

	if !ok {
		if a.Optional {
			return nil, nil
		}
		return nil, fmt.Errorf("%s", request.Lang.Get("auth.no-credentials-provided"))
	}

	user, err := a.UserService.FindByUsername(request.Context(), username)

	notFound := errors.Is(err, gorm.ErrRecordNotFound)
	if err != nil && !notFound {
		panic(errorutil.New(err))
	}

	if notFound {
		return nil, fmt.Errorf("%s", request.Lang.Get("auth.invalid-credentials"))
	}

	t := reflect.Indirect(reflect.ValueOf(user))
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	pass := t.FieldByName(a.PasswordField)
	if pass.Kind() == reflect.Invalid {
		panic(errorutil.Errorf("could not find valid field/column %q in type %T", a.PasswordField, user))
	}

	if bcrypt.CompareHashAndPassword([]byte(pass.String()), []byte(password)) != nil {
		return nil, fmt.Errorf("%s", request.Lang.Get("auth.invalid-credentials"))
	}

	return user, nil
}

func (a *BasicAuthenticator[T]) Scheme() string {
	return "Basic"
}

//--------------------------------------------

type BasicConfig struct {
	Username string
	Password string
}

func (BasicConfig) RuleSet() validation.RuleSet {
	return validation.RuleSet{
		{Path: validation.CurrentElement, Rules: validation.List{validation.Object()}},
		{Path: "Username", Rules: validation.List{validation.Required(), validation.String()}},
		{Path: "Password", Rules: validation.List{validation.Required(), validation.String()}},
	}
}

// BasicUser a simple user for config-based basic authentication.
type BasicUser struct {
	Name string
}

// ConfigBasicAuthenticator implementation of Authenticator with the Basic
// authentication method, using username and password from the configuration.
type ConfigBasicAuthenticator struct {
	config *BasicConfig
}

// Authenticate check if the request basic auth header matches the
// "auth.basic.username" and "auth.basic.password" config entries.
func (a *ConfigBasicAuthenticator) Authenticate(request *goyave.Request) (*BasicUser, error) {
	username, password, ok := request.BasicAuth()

	if !ok {
		return nil, fmt.Errorf("%s", request.Lang.Get("auth.no-credentials-provided"))
	}

	if subtle.ConstantTimeCompare([]byte(a.config.Username), []byte(username)) != 1 ||
		subtle.ConstantTimeCompare([]byte(a.config.Password), []byte(password)) != 1 {
		return nil, fmt.Errorf("%s", request.Lang.Get("auth.invalid-credentials"))
	}

	return &BasicUser{
		Name: username,
	}, nil
}

func (a *ConfigBasicAuthenticator) Scheme() string {
	return "Basic"
}

// ConfigBasicAuth create a new authenticator middleware for
// config-based Basic authentication. On auth success, the request
// user is set to a `*BasicUser`.
// The user is authenticated if the "config.Username" and "config.Password"
// match the request's Authorization header.
func ConfigBasicAuth(config *BasicConfig) *Handler[BasicUser] {
	return Middleware(&ConfigBasicAuthenticator{
		config: config,
	})
}

// ConfigBasicAuthWithRealm is the same as ConfigBasicAuth but with a custom realm description.
// The realm describes the protected area and is returned in the `WWW-Authenticate` header
// when the authentication fails.
func ConfigBasicAuthWithRealm(config *BasicConfig, realm string) *Handler[BasicUser] {
	return MiddlewareWithRealm(&ConfigBasicAuthenticator{
		config: config,
	}, realm)
}
