package auth

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"goyave.dev/goyave/v3"
	"goyave.dev/goyave/v3/database"
	"goyave.dev/goyave/v3/lang"
	"goyave.dev/goyave/v3/validation"
)

// TokenFunc is the function used by JWTController to generate tokens
// during login process.
type TokenFunc func(request *goyave.Request, user interface{}) (string, error)

// JWTController a controller for JWT-based authentication, using HMAC SHA256.
// Its model fields are used for username and password retrieval.
type JWTController struct {
	model interface{}

	// SigningMethod used to generate the token using the default
	// TokenFunc. By default, uses `jwt.SigningMethodHS256`.
	SigningMethod jwt.SigningMethod

	TokenFunc TokenFunc

	// UsernameField the name of the request's body field
	// used as username in the authentication process
	UsernameField string
	// PasswordField the name of the request's body field
	// used as password in the authentication process
	PasswordField string
}

// NewJWTController create a new JWTController that will
// be using the given model for login and token generation.
func NewJWTController(model interface{}) *JWTController {
	controller := &JWTController{
		model:         model,
		UsernameField: "username",
		PasswordField: "password",
	}
	controller.TokenFunc = func(r *goyave.Request, user interface{}) (string, error) {
		signingMethod := controller.SigningMethod
		if signingMethod == nil {
			signingMethod = jwt.SigningMethodHS256
		}
		return GenerateTokenWithClaims(jwt.MapClaims{"userid": r.String(controller.UsernameField)}, signingMethod)
	}
	return controller
}

// Login POST handler for token-based authentication.
// Creates a new token for the user authenticated with the body fields
// defined in the controller and returns it as a response.
// (the "username" and "password" body field are used by default)
//
// The database request is executed based on the model name and the
// struct tags `auth:"username"` and `auth:"password"`.
// The password is checked using bcrypt. The username field should unique.
func (c *JWTController) Login(response *goyave.Response, request *goyave.Request) {
	userType := reflect.Indirect(reflect.ValueOf(c.model)).Type()
	user := reflect.New(userType).Interface()
	username := request.String(c.UsernameField)
	columns := FindColumns(user, "username", "password")

	result := database.GetConnection().Where(columns[0].Name+" = ?", username).First(user)
	notFound := errors.Is(result.Error, gorm.ErrRecordNotFound)

	if result.Error != nil && !notFound {
		panic(result.Error)
	}

	pass := reflect.Indirect(reflect.ValueOf(user)).FieldByName(columns[1].Field.Name)
	if !notFound && bcrypt.CompareHashAndPassword([]byte(pass.String()), []byte(request.String(c.PasswordField))) == nil {
		token, err := c.TokenFunc(request, user)
		if err != nil {
			panic(err)
		}
		response.JSON(http.StatusOK, map[string]string{"token": token})
		return
	}

	response.JSON(http.StatusUnauthorized, map[string]string{"validationError": lang.Get(request.Lang, "auth.invalid-credentials")})
}

// Refresh refresh the current token.
// func (c *TokenAuthController) Refresh(response *goyave.Response, request *goyave.Request) {
// TODO refresh token
// }

// JWTRoutes create a "/auth" route group and registers the "POST /auth/login"
// validated route. Returns the new route group.
//
// Validation rules are as follows:
//  - "username": required string
//  - "password": required string
//
// The given model is used for username and password retrieval and for
// instantiating an authenticated request's user.
func JWTRoutes(router *goyave.Router, model interface{}) *goyave.Router {
	jwtRouter := router.Subrouter("/auth")
	jwtRouter.Route("POST", "/login", NewJWTController(model).Login).Validate(&validation.Rules{
		Fields: validation.FieldMap{
			"username": {
				Rules: []*validation.Rule{
					{Name: "required"},
					{Name: "string"},
				},
			},
			"password": {
				Rules: []*validation.Rule{
					{Name: "required"},
					{Name: "string"},
				},
			},
		},
	})
	return jwtRouter
}
