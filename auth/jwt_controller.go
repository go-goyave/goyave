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
	errorutil "goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/fsutil/osfs"
	"goyave.dev/goyave/v5/validation"
)

// TokenFunc is the function used by JWTController to generate tokens
// during login process.
type TokenFunc[T any] func(request *goyave.Request, user *T) (string, error)

// JWTController controller adding a login route returning a JWT for quick prototyping.
//
// The T parameter represents the user DTO and should not be a pointer. The DTO used should be
// different from the DTO returned to clients as a response because it needs to contain the user's password.
type JWTController[T any] struct { // TODO refresh token
	goyave.Component

	jwtService *JWTService

	UserService UserService[T]

	// SigningMethod used to generate the token using the default
	// TokenFunc. By default, uses `jwt.SigningMethodHS256`.
	SigningMethod jwt.SigningMethod

	// The function generating the token on a successful authentication.
	// Defaults to a JWT signed with HS256 and containing the username as the
	// "sub" claim.
	TokenFunc TokenFunc[T]

	// UsernameRequestField the name of the request's body field
	// used as username in the authentication process.
	// Defaults to "username"
	UsernameRequestField string
	// PasswordRequestField the name of the request's body field
	// used as password in the authentication process.
	// Defaults to "password"
	PasswordRequestField string
	// PasswordField the name of T's struct field that holds the user's hashed password.
	// It will be used to compare the password hash with the user input.
	PasswordField string
}

// NewJWTController create a new JWTController that registers a login route returning a JWT for quick prototyping.
//
// The `passwordField` corresponds to the name of T's struct field that holds the user's hashed password.
// It will be used to compare the password hash with the user input.
func NewJWTController[T any](userService UserService[T], passwordField string) *JWTController[T] {
	return &JWTController[T]{
		UserService:   userService,
		PasswordField: passwordField,
	}
}

// Init the controller. Automatically registers the `JWTService` if not already registered,
// using `osfs.FS` as file system for the signing keys.
func (c *JWTController[T]) Init(server *goyave.Server) {
	c.Component.Init(server)

	service, ok := server.LookupService(JWTServiceName)
	if !ok {
		service = NewJWTService(server.Config(), &osfs.FS{})
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
		{Path: lo.Ternary(c.UsernameRequestField == "", "username", c.UsernameRequestField), Rules: validation.List{
			validation.Required(),
			validation.String(),
		}},
		{Path: lo.Ternary(c.PasswordRequestField == "", "password", c.PasswordRequestField), Rules: validation.List{
			validation.Required(),
			validation.String(),
		}},
	}
}

// Login POST handler for token-based authentication.
// Creates a new token for the user authenticated with the body fields
// defined in the controller and returns it as a response.
// The password is checked using bcrypt.
func (c *JWTController[T]) Login(response *goyave.Response, request *goyave.Request) {
	body := request.Data.(map[string]any)
	username := body[lo.Ternary(c.UsernameRequestField == "", "username", c.UsernameRequestField)].(string)
	password := body[lo.Ternary(c.PasswordRequestField == "", "password", c.PasswordRequestField)].(string)

	user, err := c.UserService.FindByUsername(request.Context(), username)

	notFound := errors.Is(err, gorm.ErrRecordNotFound)
	if err != nil && !notFound {
		response.Error(errorutil.New(err))
		return
	}

	if notFound {
		response.JSON(http.StatusUnauthorized, map[string]string{"error": request.Lang.Get("auth.invalid-credentials")})
		return
	}

	t := reflect.Indirect(reflect.ValueOf(user))
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	pass := t.FieldByName(c.PasswordField)
	if pass.Kind() == reflect.Invalid {
		response.Error(errorutil.Errorf("Could not find valid field/column %q in type %T", c.PasswordField, user))
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(pass.String()), []byte(password)) == nil {
		tokenFunc := lo.Ternary(c.TokenFunc == nil, c.defaultTokenFunc, c.TokenFunc)
		token, err := tokenFunc(request, user)
		if err != nil {
			response.Error(errorutil.New(err))
			return
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
	usernameField := lo.Ternary(c.UsernameRequestField == "", "username", c.UsernameRequestField)
	return c.jwtService.GenerateTokenWithClaims(jwt.MapClaims{"sub": body[usernameField]}, signingMethod)
}
