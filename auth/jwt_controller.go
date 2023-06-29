package auth

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/golang-jwt/jwt"
	"github.com/samber/lo"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/middleware/parse"
	"goyave.dev/goyave/v5/validation"
)

// TokenFunc is the function used by JWTController to generate tokens
// during login process.
type TokenFunc[T any] func(request *goyave.Request, user *T) (string, error)

// JWTController controller adding a login route returning a JWT for quick prototyping.
//
// The T parameter represents the user model and should NOT be a pointer.
type JWTController[T any] struct { // TODO refresh token
	goyave.Component

	jwtService *JWTService

	// SigningMethod used to generate the token using the default
	// TokenFunc. By default, uses `jwt.SigningMethodHS256`.
	SigningMethod jwt.SigningMethod

	// The function generating the token on a successful authentication.
	// Defaults to a JWT signed with HS256 and containing the username as the
	// "sub" claim.
	TokenFunc TokenFunc[T]

	// UsernameField the name of the request's body field
	// used as username in the authentication process.
	// Defaults to "username"
	UsernameField string
	// PasswordField the name of the request's body field
	// used as password in the authentication process.
	// Defaults to "password"
	PasswordField string
}

// Init the controller. Automatically registers the `JWTService` if not already registered.
func (c *JWTController[T]) Init(server *goyave.Server) {
	c.Component.Init(server)

	service, ok := server.LookupService(JWTServiceName)
	if !ok {
		service = &JWTService{}
		server.RegisterService(service)
	}
	c.jwtService = service.(*JWTService)
}

// RegisterRoutes register the "/login" route (with validation) on the given router.
func (c *JWTController[T]) RegisterRoutes(router *goyave.Router) {
	router.Post("/login", c.Login).Middleware(&parse.Middleware{}).ValidateBody(c.validationRules)
}

func (c *JWTController[T]) validationRules(_ *goyave.Request) validation.RuleSet {
	return validation.RuleSet{
		{Path: validation.CurrentElement, Rules: validation.List{
			validation.Required(),
			validation.Object(),
		}},
		{Path: lo.Ternary(c.UsernameField == "", "username", c.UsernameField), Rules: validation.List{
			validation.Required(),
			validation.String(),
		}},
		{Path: lo.Ternary(c.PasswordField == "", "password", c.PasswordField), Rules: validation.List{
			validation.Required(),
			validation.String(),
		}},
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
func (c *JWTController[T]) Login(response *goyave.Response, request *goyave.Request) {
	user := new(T)
	body := request.Data.(map[string]any)
	username := body[lo.Ternary(c.UsernameField == "", "username", c.UsernameField)].(string)
	password := body[lo.Ternary(c.PasswordField == "", "password", c.PasswordField)].(string)
	columns := FindColumns(c.DB(), user, "username", "password")

	result := c.DB().Where(columns[0].Name, username).First(user)
	notFound := errors.Is(result.Error, gorm.ErrRecordNotFound)

	if result.Error != nil && !notFound {
		panic(result.Error)
	}

	t := reflect.Indirect(reflect.ValueOf(user))
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	pass := t.FieldByName(columns[1].Field.Name)
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

func (c *JWTController[T]) defaultTokenFunc(r *goyave.Request, _ *T) (string, error) {
	signingMethod := c.SigningMethod
	if signingMethod == nil {
		signingMethod = jwt.SigningMethodHS256
	}
	body := r.Data.(map[string]any)
	usernameField := lo.Ternary(c.UsernameField == "", "username", c.UsernameField)
	return c.jwtService.GenerateTokenWithClaims(jwt.MapClaims{"sub": body[usernameField]}, signingMethod)
}
