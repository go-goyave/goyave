package auth

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/util/testutil"
)

func TestBasicAuthenticator(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		mockUserService := &MockUserService[TestUser]{user: user}
		authenticator := Middleware(NewBasicAuthenticator(mockUserService, "Password"))

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "secret")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Equal(t, user.ID, request.User.(*TestUser).ID)
			assert.Equal(t, user.Name, request.User.(*TestUser).Name)
			assert.Equal(t, user.Email, request.User.(*TestUser).Email)
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
	})

	t.Run("success_ptr", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		mockUserService := &MockUserService[*TestUser]{user: &user}
		authenticator := Middleware(NewBasicAuthenticator(mockUserService, "Password"))

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "secret")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			u := request.User.(**TestUser)
			assert.Equal(t, user.ID, (*u).ID)
			assert.Equal(t, user.Name, (*u).Name)
			assert.Equal(t, user.Email, (*u).Email)
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
	})

	t.Run("wrong_password", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		mockUserService := &MockUserService[TestUser]{user: user}
		authenticator := Middleware(NewBasicAuthenticator(mockUserService, "Password"))

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "wrong password")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, _ *goyave.Request) {
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.invalid-credentials")}, body)
	})

	t.Run("service_error", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		buf := &bytes.Buffer{}
		server.Logger = slog.New(slog.NewHandler(false, buf))
		mockUserService := &MockUserService[TestUser]{err: fmt.Errorf("service_error")}
		authenticator := Middleware(NewBasicAuthenticator(mockUserService, "Password"))

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "secret")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, _ *goyave.Request) {
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
	})

	t.Run("optional_success", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		mockUserService := &MockUserService[TestUser]{user: user}
		a := NewBasicAuthenticator(mockUserService, "Password")
		a.Optional = true
		authenticator := Middleware(a)

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "secret")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Equal(t, user.ID, request.User.(*TestUser).ID)
			assert.Equal(t, user.Name, request.User.(*TestUser).Name)
			assert.Equal(t, user.Email, request.User.(*TestUser).Email)
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
	})

	t.Run("optional_wrong_password", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		mockUserService := &MockUserService[TestUser]{user: user}
		a := NewBasicAuthenticator(mockUserService, "Password")
		a.Optional = true
		authenticator := Middleware(a)

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "wrong password")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, _ *goyave.Request) {
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.invalid-credentials")}, body)
	})

	t.Run("optional_no_auth", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		mockUserService := &MockUserService[TestUser]{user: user}
		a := NewBasicAuthenticator(mockUserService, "Password")
		a.Optional = true
		authenticator := Middleware(a)

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Nil(t, request.User)
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
	})

	t.Run("no_auth", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		mockUserService := &MockUserService[TestUser]{user: user}
		a := NewBasicAuthenticator(mockUserService, "Password")
		authenticator := Middleware(a)

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, _ *goyave.Request) {
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.no-credentials-provided")}, body)
	})

	t.Run("non-existing_password_field", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		buf := &bytes.Buffer{}
		server.Logger = slog.New(slog.NewHandler(false, buf))
		mockUserService := &MockUserService[TestUser]{user: user}
		authenticator := Middleware(NewBasicAuthenticator(mockUserService, "NotAColumn"))

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "secret")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, _ *goyave.Request) {
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
	})
}

func TestConfigBasicAuthenticator(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("auth.basic.username", "johndoe")
		cfg.Set("auth.basic.password", "secret")
		server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: cfg})
		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth("johndoe", "secret")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(ConfigBasicAuth(), request, func(response *goyave.Response, request *goyave.Request) {
			assert.Equal(t, "johndoe", request.User.(*BasicUser).Name)
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
	})

	t.Run("wrong_password", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("auth.basic.username", "johndoe")
		cfg.Set("auth.basic.password", "secret")
		server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: cfg})
		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth("johndoe", "wrong_password")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(ConfigBasicAuth(), request, func(response *goyave.Response, _ *goyave.Request) {
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.invalid-credentials")}, body)
	})

	t.Run("no_auth", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("auth.basic.username", "johndoe")
		cfg.Set("auth.basic.password", "secret")
		server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: cfg})
		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(ConfigBasicAuth(), request, func(response *goyave.Response, _ *goyave.Request) {
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.no-credentials-provided")}, body)
	})
}
