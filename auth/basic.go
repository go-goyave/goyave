package auth

import (
	"reflect"

	"github.com/System-Glitch/goyave/v2"
	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/database"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// BasicAuthenticatable implementation of Authenticatable with the Basic
// authentication method, using config as credentials provider.
type BasicAuthenticatable struct{}

var _ Authenticatable = (*BasicAuthenticatable)(nil) // implements Authenticatable

// Authenticate fetch the user corresponding to the credentials
// found in the given request and puts the result in the given user pointer.
// If no user can be authenticated, returns false.
//
// The database request is executed based on the model name and the
// struct tags `auth:"username"` and `auth:"password"`.
// The password is checked using bcrypt. The username field should unique.
func (ba *BasicAuthenticatable) Authenticate(request *goyave.Request, user interface{}) bool {
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

var _ Authenticatable = (*BasicUser)(nil) // implements Authenticatable

// TODO write test for basic auth via config

// Authenticate check if the given username and password match the
// "auth:username" and "auth:password" config entries.
func (u *BasicUser) Authenticate(request *goyave.Request, user interface{}) bool {
	username, password, ok := request.BasicAuth()

	if !ok ||
		username != config.GetString("auth.username") ||
		password != config.GetString("auth.password") {
		return false
	}
	user = &BasicUser{Name: username}
	return true
}
