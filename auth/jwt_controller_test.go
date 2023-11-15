package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/util/testutil"
	"goyave.dev/goyave/v5/validation"
)

func TestJWTController(t *testing.T) {

	t.Run("Login", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		server.Config().Set("auth.jwt.secret", "secret")

		controller := &JWTController[TestUser]{}
		server.RegisterRoutes(func(s *goyave.Server, router *goyave.Router) {
			router.Controller(controller)
		})

		data := map[string]any{
			"username": user.Email,
			"password": "secret",
		}
		body, err := json.Marshal(data)
		if !assert.NoError(t, err) {
			return
		}
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		resp := server.TestRequest(request)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		respBody, err := testutil.ReadJSONBody[map[string]any](resp.Body)
		_ = resp.Body.Close()
		if !assert.NoError(t, err) {
			return
		}
		assert.NotEmpty(t, respBody["token"])
	})

	t.Run("Login_invalid", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		server.Config().Set("auth.jwt.secret", "secret")

		controller := &JWTController[TestUser]{}
		server.RegisterRoutes(func(s *goyave.Server, router *goyave.Router) {
			router.Controller(controller)
		})

		data := map[string]any{
			"username": user.Email,
			"password": "wrong password",
		}
		body, err := json.Marshal(data)
		if !assert.NoError(t, err) {
			return
		}
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		resp := server.TestRequest(request)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		respBody, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		_ = resp.Body.Close()
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.invalid-credentials")}, respBody)
	})

	t.Run("Login_error_no_table", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("database.connection", "sqlite3")
		cfg.Set("database.name", "testauthenticator.db")
		cfg.Set("database.options", "mode=memory")
		cfg.Set("app.debug", false)
		server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: cfg})
		server.Config().Set("auth.jwt.secret", "secret")

		controller := &JWTController[TestUser]{}
		server.RegisterRoutes(func(s *goyave.Server, router *goyave.Router) {
			router.Controller(controller)
		})

		data := map[string]any{
			"username": "johndoe@example.org",
			"password": "secret",
		}
		body, err := json.Marshal(data)
		if !assert.NoError(t, err) {
			return
		}
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		resp := server.TestRequest(request)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		_ = resp.Body.Close()
	})

	t.Run("Login_token_func_error", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		server.Config().Set("auth.jwt.secret", "secret")

		controller := &JWTController[TestUser]{
			TokenFunc: func(request *goyave.Request, user *TestUser) (string, error) {
				return "", fmt.Errorf("test error")
			},
		}
		server.RegisterRoutes(func(s *goyave.Server, router *goyave.Router) {
			router.Controller(controller)
		})

		data := map[string]any{
			"username": user.Email,
			"password": "secret",
		}
		body, err := json.Marshal(data)
		if !assert.NoError(t, err) {
			return
		}
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		resp := server.TestRequest(request)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		_ = resp.Body.Close()
	})

	t.Run("Login_with_field_override", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		server.Config().Set("auth.jwt.secret", "secret")

		controller := &JWTController[TestUser]{
			UsernameField: "email",
			PasswordField: "pass",
		}
		server.RegisterRoutes(func(s *goyave.Server, router *goyave.Router) {
			router.Controller(controller)
		})

		data := map[string]any{
			"email": user.Email,
			"pass":  "secret",
		}
		body, err := json.Marshal(data)
		if !assert.NoError(t, err) {
			return
		}
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		resp := server.TestRequest(request)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		respBody, err := testutil.ReadJSONBody[map[string]any](resp.Body)
		_ = resp.Body.Close()
		if !assert.NoError(t, err) {
			return
		}
		assert.NotEmpty(t, respBody["token"])
	})

	t.Run("Login_validation", func(t *testing.T) {
		server, _ := prepareAuthenticatorTest(t)
		server.Config().Set("auth.jwt.secret", "secret")

		controller := &JWTController[TestUser]{}
		server.RegisterRoutes(func(s *goyave.Server, router *goyave.Router) {
			router.Controller(controller)
		})

		data := map[string]any{}
		body, err := json.Marshal(data)
		if !assert.NoError(t, err) {
			return
		}
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		resp := server.TestRequest(request)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		respBody, err := testutil.ReadJSONBody[map[string]*validation.ErrorResponse](resp.Body)
		_ = resp.Body.Close()
		if !assert.NoError(t, err) {
			return
		}
		if assert.Contains(t, respBody, "error") && assert.NotNil(t, respBody["error"]) {
			assert.Contains(t, respBody["error"].Body.Fields, "username")
			assert.Contains(t, respBody["error"].Body.Fields, "password")
		}
	})
}
