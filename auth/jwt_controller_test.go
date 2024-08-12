package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/util/testutil"
	"goyave.dev/goyave/v5/validation"
)

func TestJWTController(t *testing.T) {
	t.Run("Login", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		server.Config().Set("auth.jwt.secret", "secret")

		mockUserService := &MockUserService[TestUser]{user: user}
		controller := NewJWTController(mockUserService, "Password")
		server.RegisterRoutes(func(_ *goyave.Server, router *goyave.Router) {
			router.Controller(controller)
		})

		data := map[string]any{
			"username": user.Email,
			"password": "secret",
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		resp := server.TestRequest(request)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		respBody, err := testutil.ReadJSONBody[map[string]any](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.NotEmpty(t, respBody["token"])
	})

	t.Run("Login_ptr", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		server.Config().Set("auth.jwt.secret", "secret")

		mockUserService := &MockUserService[*TestUser]{user: &user}
		controller := NewJWTController(mockUserService, "Password")
		server.RegisterRoutes(func(_ *goyave.Server, router *goyave.Router) {
			router.Controller(controller)
		})

		data := map[string]any{
			"username": user.Email,
			"password": "secret",
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		resp := server.TestRequest(request)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		respBody, err := testutil.ReadJSONBody[map[string]any](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.NotEmpty(t, respBody["token"])
	})

	t.Run("Login_invalid_password", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		server.Config().Set("auth.jwt.secret", "secret")

		mockUserService := &MockUserService[TestUser]{user: user}
		controller := NewJWTController(mockUserService, "Password")
		server.RegisterRoutes(func(_ *goyave.Server, router *goyave.Router) {
			router.Controller(controller)
		})

		data := map[string]any{
			"username": user.Email,
			"password": "wrong password",
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		resp := server.TestRequest(request)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		respBody, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.invalid-credentials")}, respBody)
	})

	t.Run("Login_invalid_username", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		server.Config().Set("auth.jwt.secret", "secret")

		mockUserService := &MockUserService[TestUser]{err: fmt.Errorf("test errors: %w", gorm.ErrRecordNotFound)}
		controller := NewJWTController(mockUserService, "Password")
		server.RegisterRoutes(func(_ *goyave.Server, router *goyave.Router) {
			router.Controller(controller)
		})

		data := map[string]any{
			"username": "wrong username",
			"password": user.Password,
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		resp := server.TestRequest(request)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		respBody, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.invalid-credentials")}, respBody)
	})

	t.Run("Login_token_func_error", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		buf := &bytes.Buffer{}
		server.Logger = slog.New(slog.NewHandler(false, buf))
		server.Config().Set("auth.jwt.secret", "secret")

		mockUserService := &MockUserService[TestUser]{user: user}
		controller := NewJWTController(mockUserService, "Password")
		controller.TokenFunc = func(_ *goyave.Request, _ *TestUser) (string, error) {
			return "", fmt.Errorf("test error")
		}
		server.RegisterRoutes(func(_ *goyave.Server, router *goyave.Router) {
			router.Controller(controller)
		})

		data := map[string]any{
			"username": user.Email,
			"password": "secret",
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		resp := server.TestRequest(request)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
		assert.NotEmpty(t, buf.String())
	})

	t.Run("Login_non-existing_password_field", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		buf := &bytes.Buffer{}
		server.Logger = slog.New(slog.NewHandler(false, buf))
		server.Config().Set("auth.jwt.secret", "secret")

		mockUserService := &MockUserService[TestUser]{user: user}
		controller := NewJWTController(mockUserService, "NotAField")
		server.RegisterRoutes(func(_ *goyave.Server, router *goyave.Router) {
			router.Controller(controller)
		})

		data := map[string]any{
			"username": user.Email,
			"password": "secret",
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		resp := server.TestRequest(request)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
		assert.NotEmpty(t, buf.String())
	})

	t.Run("Login_service_error", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		buf := &bytes.Buffer{}
		server.Logger = slog.New(slog.NewHandler(false, buf))
		server.Config().Set("auth.jwt.secret", "secret")

		mockUserService := &MockUserService[TestUser]{err: fmt.Errorf("service error")}
		controller := NewJWTController(mockUserService, "NotAField")
		server.RegisterRoutes(func(_ *goyave.Server, router *goyave.Router) {
			router.Controller(controller)
		})

		data := map[string]any{
			"username": user.Email,
			"password": "secret",
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		resp := server.TestRequest(request)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
		assert.NotEmpty(t, buf.String())
	})

	t.Run("Login_with_field_override", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		server.Config().Set("auth.jwt.secret", "secret")

		mockUserService := &MockUserService[TestUser]{user: user}
		controller := NewJWTController(mockUserService, "Password")
		controller.UsernameRequestField = "email"
		controller.PasswordRequestField = "pass"
		server.RegisterRoutes(func(_ *goyave.Server, router *goyave.Router) {
			router.Controller(controller)
		})

		data := map[string]any{
			"email": user.Email,
			"pass":  "secret",
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		resp := server.TestRequest(request)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		respBody, err := testutil.ReadJSONBody[map[string]any](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.NotEmpty(t, respBody["token"])
	})

	t.Run("Login_validation", func(t *testing.T) {
		server, _ := prepareAuthenticatorTest(t)
		server.Config().Set("auth.jwt.secret", "secret")

		controller := &JWTController[TestUser]{}
		server.RegisterRoutes(func(_ *goyave.Server, router *goyave.Router) {
			router.Controller(controller)
		})

		data := map[string]any{}
		body, err := json.Marshal(data)
		require.NoError(t, err)
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		resp := server.TestRequest(request)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		respBody, err := testutil.ReadJSONBody[map[string]*validation.ErrorResponse](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		if assert.Contains(t, respBody, "error") && assert.NotNil(t, respBody["error"]) {
			assert.Contains(t, respBody["error"].Body.Fields, "username")
			assert.Contains(t, respBody["error"].Body.Fields, "password")
		}
	})
}
