package auth

import (
	"reflect"

	"github.com/System-Glitch/goyave/v2"
	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/database"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// BasicAuthenticator implementation of Authenticator with the Basic
// authentication method, using config as credentials provider.
type BasicAuthenticator struct{}

var _ Authenticator = (*BasicAuthenticator)(nil) // implements Authenticator

// Authenticate fetch the user corresponding to the credentials
// found in the given request and puts the result in the given user pointer.
// If no user can be authenticated, returns false.
//
// The database request is executed based on the model name and the
// struct tags `auth:"username"` and `auth:"password"`.
// The password is checked using bcrypt. The username field should unique.
func (ba *BasicAuthenticator) Authenticate(request *goyave.Request, user interface{}) bool {
	username, password, ok := request.BasicAuth()

	if !ok {
		return false
	}

	columns := FindColumns(user, "username", "password")

	result := database.GetConnection().Where(columns[0].Name+" = ?", username).First(user)

	if errors := result.GetErrors(); len(errors) != 0 && !gorm.IsRecordNotFoundError(result.Error) {
		panic(errors)
	}

	pass := reflect.Indirect(reflect.ValueOf(user)).FieldByName(columns[1].Field.Name)
	if result.RecordNotFound() || bcrypt.CompareHashAndPassword([]byte(pass.String()), []byte(password)) != nil {
		return false
	}

	return true
}

//--------------------------------------------

// BasicUser a simple user for config-based basic authentication.
type BasicUser struct {
	Name string
}

type basicUserAuthenticator struct{}

var _ Authenticator = (*basicUserAuthenticator)(nil) // implements Authenticator

// Authenticate check if the request basic auth header matches the
// "authUsername" and "authPassword" config entries.
func (a *basicUserAuthenticator) Authenticate(request *goyave.Request, user interface{}) bool {
	username, password, ok := request.BasicAuth()

	if !ok ||
		username != config.GetString("authUsername") ||
		password != config.GetString("authPassword") {
		return false
	}
	user.(*BasicUser).Name = username
	return true
}

// ConfigBasicAuth create a new authenticator middleware for
// config-based Basic authentication. On auth success, the request
// user is set to a "BasicUser".
// The user is authenticated if the "authUsername" and "authPassword" config entries
// match the request's Authorization header.
func ConfigBasicAuth() goyave.Middleware {
	return Middleware(&BasicUser{}, &basicUserAuthenticator{})
}
