package auth

import (
	"net/http"
	"reflect"
	"strings"

	"slices"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"goyave.dev/goyave/v5"
)

// Column matches a column name with a struct field.
type Column struct {
	Field *reflect.StructField
	Name  string
}

// Authenticator is an object in charge of authenticating a model.
type Authenticator interface {
	goyave.Composable

	// Authenticate fetch the user corresponding to the credentials
	// found in the given request and puts the result in the given user pointer.
	// If no user can be authenticated, returns the error detailing why the
	// authentication failed. The error message is already localized.
	//
	// The `user` is a double pointer to a `nil` structure defined by the generic
	// parameter of the middleware.
	Authenticate(request *goyave.Request, user any) error
}

// Unauthorizer can be implemented by Authenticators to define custom behavior
// when authentication fails.
type Unauthorizer interface {
	OnUnauthorized(*goyave.Response, *goyave.Request, error)
}

// Handler a middleware that automatically sets the request's User before
// executing the authenticator. Supports the `Unauthorizer` interface.
//
// The T parameter represents the user model and should be a pointer.
type Handler[T any] struct {
	Authenticator
}

// Handle set the request's user to a new instance of the model before
// executing the authenticator. Blocks if the authentication is not successful.
// If the authenticator implements `Unauthorizer`, `OnUnauthorized` is called,
// otherwise returns a default `401 Unauthorized` error.
func (m *Handler[T]) Handle(next goyave.Handler) goyave.Handler {
	return func(response *goyave.Response, request *goyave.Request) {
		user := new(T)
		if err := m.Authenticator.Authenticate(request, user); err != nil {
			if unauthorizer, ok := m.Authenticator.(Unauthorizer); ok {
				unauthorizer.OnUnauthorized(response, request, err)
				return
			}
			response.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		request.User = *user
		next(response, request)
	}
}

// Middleware returns an authentication middleware which will use the given
// authenticator and set the request's user according to the given model.
func Middleware[T any](authenticator Authenticator) *Handler[T] {
	return &Handler[T]{
		Authenticator: authenticator,
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
func FindColumns(db *gorm.DB, strct any, fields ...string) []*Column {
	return findColumns(db, reflect.TypeOf(strct), fields)
}

func findColumns(db *gorm.DB, t reflect.Type, fields []string) []*Column {
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
			for i, v := range findColumns(db, fieldType, fields) {
				if v != nil {
					result[i] = v
				}
			}
			continue
		}

		tag := strctType.Tag.Get("auth")
		if index := slices.Index(fields, tag); index != -1 {
			result[index] = &Column{
				Name:  columnName(strctType, db.NamingStrategy),
				Field: &strctType,
			}
		}
	}

	return result
}

func columnName(field reflect.StructField, namer schema.Namer) string {
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
