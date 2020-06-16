package auth

import (
	"net/http"
	"reflect"
	"time"

	"github.com/System-Glitch/goyave/v2"
	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/database"
	"github.com/System-Glitch/goyave/v2/lang"
	"github.com/System-Glitch/goyave/v2/validation"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// GenerateToken generate a new JWT.
// The token is created using the HMAC SHA256 method and signed using
// the "jwtSecret" config entry.
// The token is set to expire in the amount of seconds defined by
// the "jwtExpiry" config entry.
func GenerateToken(id interface{}) (string, error) {
	expiry := time.Duration(config.Get("jwtExpiry").(float64)) * time.Second
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userid": id,
		"nbf":    now.Unix(),             // Not Before
		"exp":    now.Add(expiry).Unix(), // Expiry
	})

	return token.SignedString([]byte(config.GetString("jwtSecret")))
}

// JWTController a controller for JWT-based authentication, using HMAC SHA256.
// Its model fields are used for username and password retrieval.
type JWTController struct {
	model interface{}
}

// NewJWTController create a new JWTController that will
// be using the given model for login and token generation.
func NewJWTController(model interface{}) *JWTController {
	return &JWTController{model}
}

// Login POST handler for token-based authentication.
// Creates a new token for the user authenticated with the "username" and
// "password" body fields and returns it as a response.
//
// The database request is executed based on the model name and the
// struct tags `auth:"username"` and `auth:"password"`.
// The password is checked using bcrypt. The username field should unique.
func (c *JWTController) Login(response *goyave.Response, request *goyave.Request) {
	userType := reflect.Indirect(reflect.ValueOf(c.model)).Type()
	user := reflect.New(userType).Interface()
	username := request.String("username")
	columns := FindColumns(user, "username", "password")

	result := database.GetConnection().Where(columns[0].Name+" = ?", username).First(user)

	if errors := result.GetErrors(); len(errors) != 0 && !gorm.IsRecordNotFoundError(result.Error) {
		panic(errors)
	}

	pass := reflect.Indirect(reflect.ValueOf(user)).FieldByName(columns[1].Field.Name)
	if !result.RecordNotFound() && bcrypt.CompareHashAndPassword([]byte(pass.String()), []byte(request.String("password"))) == nil {
		token, err := GenerateToken(username)
		if err != nil {
			panic(err)
		}
		response.JSON(http.StatusOK, map[string]string{"token": token})
		return
	}

	response.JSON(http.StatusUnprocessableEntity, map[string]string{"validationError": lang.Get(request.Lang, "auth.invalid-credentials")})
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
	jwtRouter.Route("POST", "/login", NewJWTController(model).Login).Validate(validation.Rules{
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
	})
	return jwtRouter
}
