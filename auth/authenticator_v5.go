package auth

import (
	"net/http"
	"reflect"
	"strings"

	"golang.org/x/exp/slices"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"goyave.dev/goyave/v4"
)

// Authenticator is an object in charge of authenticating a model.
type AuthenticatorV5 interface {
	goyave.Composable

	// Authenticate fetch the user corresponding to the credentials
	// found in the given request and puts the result in the given user pointer.
	// If no user can be authenticated, returns the error detailing why the
	// authentication failed. The error message is already localized.
	//
	// The `user` is a double pointer to a `nil` structure defined by the generic
	// parameter of the middleware.
	Authenticate(request *goyave.RequestV5, user any) error
}

// Unauthorizer can be implemented by Authenticators to define custom behavior
// when authentication fails.
type UnauthorizerV5 interface {
	OnUnauthorized(*goyave.ResponseV5, *goyave.RequestV5, error)
}

// Handler a middleware that automatically sets the request's User before
// executing the authenticator. Supports the `Unauthorizer` interface.
//
// The T parameter represents the user model and should be a pointer.
type Handler[T any] struct {
	AuthenticatorV5
}

// Handle set the request's user to a new instance of the model before
// executing the authenticator. Blocks if the authentication is not successful.
// If the authenticator implements `Unauthorizer`, `OnUnauthorized` is called,
// otherwise returns a default `401 Unauthorized` error.
func (m *Handler[T]) Handle(next goyave.HandlerV5) goyave.HandlerV5 {
	return func(response *goyave.ResponseV5, request *goyave.RequestV5) {
		user := new(T)
		if err := m.AuthenticatorV5.Authenticate(request, user); err != nil {
			if unauthorizer, ok := m.AuthenticatorV5.(UnauthorizerV5); ok {
				unauthorizer.OnUnauthorized(response, request, err)
				return
			}
			response.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		if user != nil {
			request.User = *user
		}
		next(response, request)
	}
}

// Middleware returns an authentication middleware which will use the given
// authenticator and set the request's user according to the given model.
func MiddlewareV5[T any](authenticator AuthenticatorV5) *Handler[T] {
	return &Handler[T]{
		AuthenticatorV5: authenticator,
	}
}

// FindColumns in given struct. A field matches if it has a "auth" tag with the given value.
// Returns a slice of found fields, ordered as the input "fields" slice.
// If the nth field is not found, the nth value of the returned slice will be nil.
//
// Promoted fields are matched as well.
//
// Given the following struct and "username", "notatag", "password":
//
//	 type TestUser struct {
//			gorm.Model
//			Name     string `gorm:"type:varchar(100)"`
//			Password string `gorm:"type:varchar(100)" auth:"password"`
//			Email    string `gorm:"type:varchar(100);uniqueIndex" auth:"username"`
//	 }
//
// The result will be the "Email" field, "nil" and the "Password" field.
func FindColumnsV5(db *gorm.DB, strct any, fields ...string) []*Column {
	return findColumnsV5(db, reflect.TypeOf(strct), fields)
}

func findColumnsV5(db *gorm.DB, t reflect.Type, fields []string) []*Column {
	length := len(fields)
	result := make([]*Column, length)

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		strctType := t.Field(i)
		fieldType := strctType.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Struct && strctType.Anonymous {
			// Check promoted fields recursively
			for i, v := range findColumnsV5(db, fieldType, fields) {
				if v != nil {
					result[i] = v
				}
			}
			continue
		}

		tag := strctType.Tag.Get("auth")
		if index := slices.Index(fields, tag); index != -1 {
			result[index] = &Column{
				Name:  columnNameV5(strctType, db.NamingStrategy),
				Field: &strctType,
			}
		}
	}

	return result
}

func columnNameV5(field reflect.StructField, namer schema.Namer) string {
	for _, t := range strings.Split(field.Tag.Get("gorm"), ";") { // Check for gorm column name override
		if strings.HasPrefix(t, "column") {
			i := strings.Index(t, ":")
			if i == -1 || i+1 >= len(t) {
				// Invalid syntax, fallback to auto-naming
				break
			}
			return strings.TrimSpace(t[i+1:])
		}
	}

	return namer.ColumnName("", field.Name)
}
