package auth

import (
	"errors"
	"net/http"
	"reflect"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"goyave.dev/goyave/v3"
	"goyave.dev/goyave/v3/config"
	"goyave.dev/goyave/v3/database"
	"goyave.dev/goyave/v3/lang"
	"goyave.dev/goyave/v3/validation"
)

// GenerateToken generate a new JWT.
// The token is created using the HMAC SHA256 method and signed using
// the "auth.jwt.secret" config entry.
// The token is set to expire in the amount of seconds defined by
// the "auth.jwt.expiry" config entry.
func GenerateToken(username interface{}) (string, error) {
	return GenerateTokenWithClaims(jwt.MapClaims{"userid": username})
}

// GenerateTokenWithClaims generates a new JWT with custom claims.
// The token is created using the HMAC SHA256 method and signed using
// the "auth.jwt.secret" config entry.
// The token is set to expire in the amount of seconds defined by
// the "auth.jwt.expiry" config entry.
func GenerateTokenWithClaims(claims jwt.MapClaims) (string, error) {
	expiry := time.Duration(config.GetInt("auth.jwt.expiry")) * time.Second
	now := time.Now()
	customClaims := jwt.MapClaims{
		"nbf": now.Unix(),             // Not Before
		"exp": now.Add(expiry).Unix(), // Expiry
	}
	for k, c := range claims {
		customClaims[k] = c
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims)

	return token.SignedString([]byte(config.GetString("auth.jwt.secret")))
}

// TokenFunc is the function used by JWTController to generate tokens
// during login process.
type TokenFunc func(request *goyave.Request, user interface{}) (string, error)

// JWTController a controller for JWT-based authentication, using HMAC SHA256.
// Its model fields are used for username and password retrieval.
type JWTController struct {
	model     interface{}
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
		return GenerateToken(r.String(controller.UsernameField))
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
