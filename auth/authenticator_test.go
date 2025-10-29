package auth

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/util/testutil"

	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
)

type TestUser struct {
	gorm.Model
	Name     string `gorm:"type:varchar(100)"`
	Password string `gorm:"type:varchar(100)" auth:"password"`
	Email    string `gorm:"type:varchar(100);uniqueIndex" auth:"username"`
}

type MockUserService[T any] struct {
	user *T
	err  error
}

func (s MockUserService[T]) FindByUsername(_ context.Context, _ any) (*T, error) {
	return s.user, s.err
}

type TestBasicUnauthorizer struct {
	*BasicAuthenticator[TestUser]
}

func (a *TestBasicUnauthorizer) OnUnauthorized(response *goyave.Response, _ *goyave.Request, err error) {
	response.JSON(http.StatusUnauthorized, map[string]string{"custom error key": err.Error()})
}

type TestNoScheme struct {
	goyave.Component
}

func (a *TestNoScheme) Authenticate(request *goyave.Request) (*BasicUser, error) {
	return (&ConfigBasicAuthenticator{Component: a.Component}).Authenticate(request)
}

func prepareAuthenticatorTest(t *testing.T) (*testutil.TestServer, *TestUser) {
	cfg := config.LoadDefault()
	cfg.Set("database.connection", "sqlite3")
	cfg.Set("database.name", "testauthenticator.db")
	cfg.Set("database.options", "mode=memory")
	cfg.Set("app.debug", false)
	server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: cfg})
	password, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	user := &TestUser{
		Name:     "johndoe",
		Email:    "johndoe@example.org",
		Password: string(password),
	}

	return server, user
}

func TestAuthenticator(t *testing.T) {
	t.Run("Middleware", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		t.Cleanup(func() { server.CloseDB() })

		mockUserService := &MockUserService[TestUser]{user: user}
		authenticator := Middleware(NewBasicAuthenticator(mockUserService, "Password"))
		assert.Equal(t, defaultRealm, authenticator.Realm)

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "secret")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Equal(t, user.ID, request.User.(*TestUser).ID)
			assert.Equal(t, user.Name, request.User.(*TestUser).Name)
			assert.Equal(t, user.Email, request.User.(*TestUser).Email)
			assert.Same(t, request.User, UserFromContext[TestUser](request.Context()))
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
		assert.Empty(t, resp.Header.Get("WWW-Authenticate"))

		request = server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "incorrect password")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp = server.TestMiddleware(authenticator, request, func(response *goyave.Response, _ *goyave.Request) {
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.invalid-credentials")}, body)
		assert.Equal(t, `Basic realm="Authorization required", charset="UTF-8"`, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("NoAuth", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		t.Cleanup(func() { server.CloseDB() })

		mockUserService := &MockUserService[TestUser]{user: user}
		authenticator := Middleware(NewBasicAuthenticator(mockUserService, "Password"))

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "secret")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: false}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Nil(t, request.User)
			assert.Nil(t, UserFromContext[TestUser](request.Context()))
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
		assert.Empty(t, resp.Header.Get("WWW-Authenticate"))

		request.Route = &goyave.Route{Meta: map[string]any{}}
		resp = server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Nil(t, request.User)
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
		assert.Empty(t, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("MiddlewareUnauthorizer", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		t.Cleanup(func() { server.CloseDB() })

		mockUserService := &MockUserService[TestUser]{user: user}
		authenticator := Middleware(&TestBasicUnauthorizer{BasicAuthenticator: NewBasicAuthenticator(mockUserService, "Password")})

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "incorrect password")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, _ *goyave.Request) {
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"custom error key": server.Lang.GetDefault().Get("auth.invalid-credentials")}, body)
		assert.Equal(t, `Basic realm="Authorization required", charset="UTF-8"`, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("MiddlewareWithRealm", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		t.Cleanup(func() { server.CloseDB() })

		mockUserService := &MockUserService[TestUser]{user: user}
		authenticator := MiddlewareWithRealm(&TestBasicUnauthorizer{BasicAuthenticator: NewBasicAuthenticator(mockUserService, "Password")}, "custom realm")

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "incorrect password")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, _ *goyave.Request) {
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"custom error key": server.Lang.GetDefault().Get("auth.invalid-credentials")}, body)
		assert.Equal(t, `Basic realm="custom realm", charset="UTF-8"`, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("MiddlewareWithRealmNoScheme", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		t.Cleanup(func() { server.CloseDB() })

		server.Config().Set("auth.basic.username", "johndoe")
		server.Config().Set("auth.basic.password", "secret")

		authenticator := MiddlewareWithRealm(&TestNoScheme{}, "custom realm")

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "incorrect password")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, _ *goyave.Request) {
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.invalid-credentials")}, body)
		assert.Empty(t, resp.Header.Get("WWW-Authenticate")) // Authentication doesn't implement SchemeAuthenticator, header should not be added
	})
}

func TestUserContext(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, UserFromContext[TestUser](ctx))

	user := &TestUser{}
	withUser := ContextWithUser(ctx, user)
	assert.Same(t, user, UserFromContext[TestUser](withUser))

	withUser = ContextWithUser(ctx, &struct{}{})
	assert.Nil(t, UserFromContext[TestUser](withUser))
}
