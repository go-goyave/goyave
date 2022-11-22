package auth

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/golang-jwt/jwt"
	"github.com/samber/lo"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"goyave.dev/goyave/v4"
	"goyave.dev/goyave/v4/validation"
)

// TokenFunc is the function used by JWTController to generate tokens
// during login process.
type TokenFuncV5 func(request *goyave.RequestV5, user any) (string, error)

// JWTController controller adding a login route returning a JWT.
type JWTControllerV5 struct { // TODO refresh token
	goyave.Component

	model any

	// SigningMethod used to generate the token using the default
	// TokenFunc. By default, uses `jwt.SigningMethodHS256`.
	SigningMethod jwt.SigningMethod

	// The function generating the token on a successful authentication.
	// Defaults to a JWT signed with HS256 and containing the username as the
	// "sub" claim.
	TokenFunc TokenFuncV5

	// UsernameField the name of the request's body field
	// used as username in the authentication process.
	// Defaults to "username"
	UsernameField string
	// PasswordField the name of the request's body field
	// used as password in the authentication process.
	// Defaults to "password"
	PasswordField string
}

func (c *JWTControllerV5) defaultTokenFunc(r *goyave.RequestV5, user any) (string, error) {
	signingMethod := c.SigningMethod
	if signingMethod == nil {
		signingMethod = jwt.SigningMethodHS256
	}
	body := r.Data.(map[string]any)
	usernameField := lo.Ternary(c.UsernameField == "", "username", c.UsernameField)
	return GenerateTokenWithClaimsV5(c.Config(), jwt.MapClaims{"sub": body[usernameField]}, signingMethod)
}

// RegisterRoutes register the "/login" route (with validation) on the given router.
func (c *JWTControllerV5) RegisterRoutes(router *goyave.RouterV5) {
	router.Post("/login", c.Login).ValidateBody(c.validationRules)
}

func (c *JWTControllerV5) validationRules(r *goyave.RequestV5) validation.RuleSetV5 {
	return validation.RuleSetV5{
		{validation.CurrentElement, validation.ListV5{validation.Required(), validation.Object()}},
		{lo.Ternary(c.UsernameField == "", "username", c.UsernameField), validation.ListV5{validation.Required(), validation.String()}},
		{lo.Ternary(c.PasswordField == "", "password", c.PasswordField), validation.ListV5{validation.Required(), validation.String()}},
	}
}

// Login POST handler for token-based authentication.
// Creates a new token for the user authenticated with the body fields
// defined in the controller and returns it as a response.
// (the "username" and "password" body field are used by default)
//
// The database request is executed based on the model name and the
// struct tags `auth:"username"` and `auth:"password"`.
// The password is checked using bcrypt. The username field should unique.
func (c *JWTControllerV5) Login(response *goyave.ResponseV5, request *goyave.RequestV5) {
	userType := reflect.Indirect(reflect.ValueOf(c.model)).Type()
	user := reflect.New(userType).Interface()
	body := request.Data.(map[string]any)
	username := body[lo.Ternary(c.UsernameField == "", "username", c.UsernameField)].(string)
	password := body[lo.Ternary(c.PasswordField == "", "password", c.PasswordField)].(string)
	columns := FindColumnsV5(c.DB(), user, "username", "password")

	result := c.DB().Where(columns[0].Name, username).First(user)
	notFound := errors.Is(result.Error, gorm.ErrRecordNotFound)

	if result.Error != nil && !notFound {
		panic(result.Error)
	}

	pass := reflect.Indirect(reflect.ValueOf(user)).FieldByName(columns[1].Field.Name)
	if !notFound && bcrypt.CompareHashAndPassword([]byte(pass.String()), []byte(password)) == nil {
		tokenFunc := lo.Ternary(c.TokenFunc == nil, c.defaultTokenFunc, c.TokenFunc)
		token, err := tokenFunc(request, user)
		if err != nil {
			panic(err)
		}
		response.JSON(http.StatusOK, map[string]string{"token": token})
		return
	}

	response.JSON(http.StatusUnauthorized, map[string]string{"error": request.Lang.Get("auth.invalid-credentials")})
}
