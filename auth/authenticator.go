package auth

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/System-Glitch/goyave/v2"
	"github.com/System-Glitch/goyave/v2/helper"
	"github.com/jinzhu/gorm"
)

// Column matches a column name with a struct field.
type Column struct {
	Name  string
	Field *reflect.StructField
}

// Authenticator is an object in charge of authenticating a model.
type Authenticator interface {

	// Authenticate fetch the user corresponding to the credentials
	// found in the given request and puts the result in the given user pointer.
	// If no user can be authenticated, returns the error detailing why the
	// authentication failed. The error message is already localized.
	Authenticate(request *goyave.Request, user interface{}) error
}

// Middleware create a new authenticator middleware to authenticate
// the given model using the given authenticator.
func Middleware(model interface{}, authenticator Authenticator) goyave.Middleware {
	return func(next goyave.Handler) goyave.Handler {
		return func(response *goyave.Response, r *goyave.Request) {
			userType := reflect.Indirect(reflect.ValueOf(model)).Type()
			user := reflect.New(userType).Interface()
			r.User = user
			if err := authenticator.Authenticate(r, r.User); err != nil {
				response.JSON(http.StatusUnauthorized, map[string]string{"authError": err.Error()})
				return
			}
			next(response, r)
		}
	}
}

// FindColumns in given struct. A field matches if it has a "auth" tag with the given value.
// Returns a slice of found fields, ordered as the input "fields" slice.
// If the nth field is not found, the nth value of the returned slice will be nil.
func FindColumns(strct interface{}, fields ...string) []*Column {
	length := len(fields)
	result := make([]*Column, length, length)

	value := reflect.ValueOf(strct)
	t := reflect.TypeOf(strct)
	if t.Kind() == reflect.Ptr {
		value = value.Elem()
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		field := value.Field(i)
		fieldType := t.Field(i)
		if field.Kind() == reflect.Struct && fieldType.Anonymous {
			// Check promoted fields recursively
			for i, v := range FindColumns(field.Interface(), fields...) {
				if v != nil {
					result[i] = v
				}
			}
			continue
		}

		tag := fieldType.Tag.Get("auth")
		if index := helper.IndexOf(fields, tag); index != -1 {
			result[index] = &Column{
				Name:  columnName(&fieldType),
				Field: &fieldType,
			}
		}
	}

	return result
}

func columnName(field *reflect.StructField) string {
	for _, t := range strings.Split(field.Tag.Get("gorm"), ";") { // Check for gorm column name override
		if strings.HasPrefix(t, "column") {
			v := strings.Split(t, ":")
			return strings.TrimSpace(v[1])
		}
	}

	return gorm.ToColumnName(field.Name)
}
