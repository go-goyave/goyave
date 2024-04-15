package auth

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"reflect"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	errorutil "goyave.dev/goyave/v5/util/errors"
)

// BasicAuthenticator implementation of Authenticator with the Basic
// authentication method.
type BasicAuthenticator struct {
	goyave.Component

	// Optional defines if the authenticator allows requests that
	// don't provide credentials. Handlers should therefore check
	// if request.User is not nil before accessing it.
	Optional bool
}

var _ Authenticator = (*BasicAuthenticator)(nil) // implements Authenticator

// Authenticate fetch the user corresponding to the credentials
// found in the given request and puts the result in the given user pointer.
// If no user can be authenticated, returns an error.
//
// The database request is executed based on the model name and the
// struct tags `auth:"username"` and `auth:"password"`.
// The password is checked using bcrypt. The username field should unique.
func (a *BasicAuthenticator) Authenticate(request *goyave.Request, user any) error {
	username, password, ok := request.BasicAuth()

	if !ok {

		if a.Optional {
			return nil
		}
		return fmt.Errorf(request.Lang.Get("auth.no-credentials-provided"))
	}

	columns := FindColumns(a.DB(), user, "username", "password")

	result := a.DB().
		WithContext(request.Context()).
		Where(columns[0].Name, username).
		First(user)
	notFound := errors.Is(result.Error, gorm.ErrRecordNotFound)

	if result.Error != nil && !notFound {
		panic(errorutil.New(result.Error))
	}

	t := reflect.Indirect(reflect.ValueOf(user))
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	pass := t.FieldByName(columns[1].Field.Name)

	if notFound || bcrypt.CompareHashAndPassword([]byte(pass.String()), []byte(password)) != nil {
		return fmt.Errorf(request.Lang.Get("auth.invalid-credentials"))
	}

	return nil
}

//--------------------------------------------

func init() {
	config.Register("auth.basic.username", config.Entry{
		Value:            nil,
		Type:             reflect.String,
		IsSlice:          false,
		AuthorizedValues: []any{},
	})
	config.Register("auth.basic.password", config.Entry{
		Value:            nil,
		Type:             reflect.String,
		IsSlice:          false,
		AuthorizedValues: []any{},
	})
}

// BasicUser a simple user for config-based basic authentication.
type BasicUser struct {
	Name string
}

// ConfigBasicAuthenticator implementation of Authenticator with the Basic
// authentication method, using username and password from the configuration.
type ConfigBasicAuthenticator struct {
	goyave.Component
}

var _ Authenticator = (*ConfigBasicAuthenticator)(nil) // implements Authenticator

// Authenticate check if the request basic auth header matches the
// "auth.basic.username" and "auth.basic.password" config entries.
func (a *ConfigBasicAuthenticator) Authenticate(request *goyave.Request, user any) error {
	username, password, ok := request.BasicAuth()

	if !ok {
		return fmt.Errorf(request.Lang.Get("auth.no-credentials-provided"))
	}

	if subtle.ConstantTimeCompare([]byte(a.Config().GetString("auth.basic.username")), []byte(username)) != 1 ||
		subtle.ConstantTimeCompare([]byte(a.Config().GetString("auth.basic.password")), []byte(password)) != 1 {
		return fmt.Errorf(request.Lang.Get("auth.invalid-credentials"))
	}

	(*user.(**BasicUser)) = &BasicUser{
		Name: username,
	}
	return nil
}

// ConfigBasicAuth create a new authenticator middleware for
// config-based Basic authentication. On auth success, the request
// user is set to a `*BasicUser`.
// The user is authenticated if the "auth.basic.username" and "auth.basic.password" config entries
// match the request's Authorization header.
func ConfigBasicAuth() *Handler[*BasicUser] {
	return Middleware[*BasicUser](&ConfigBasicAuthenticator{})
}
