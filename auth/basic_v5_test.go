package auth

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v4"
	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/util/testutil"
)

func TestBasicAuthenticator(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		server, user := prepareAuthenticatorTest()
		authenticator := MiddlewareV5[*TestUser](&BasicAuthenticatorV5{})

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "secret")
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.ResponseV5, request *goyave.RequestV5) {
			assert.Equal(t, user.ID, request.User.(*TestUser).ID)
			assert.Equal(t, user.Name, request.User.(*TestUser).Name)
			assert.Equal(t, user.Email, request.User.(*TestUser).Email)
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close()
	})

	t.Run("wrong_password", func(t *testing.T) {
		server, user := prepareAuthenticatorTest()
		authenticator := MiddlewareV5[*TestUser](&BasicAuthenticatorV5{})

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "wrong password")
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.ResponseV5, request *goyave.RequestV5) {
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		_ = resp.Body.Close()
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.invalid-credentials")}, body)
	})

	t.Run("optional_success", func(t *testing.T) {
		server, user := prepareAuthenticatorTest()
		authenticator := MiddlewareV5[*TestUser](&BasicAuthenticatorV5{Optional: true})

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "secret")
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.ResponseV5, request *goyave.RequestV5) {
			assert.Equal(t, user.ID, request.User.(*TestUser).ID)
			assert.Equal(t, user.Name, request.User.(*TestUser).Name)
			assert.Equal(t, user.Email, request.User.(*TestUser).Email)
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close()
	})

	t.Run("optional_wrong_password", func(t *testing.T) {
		server, user := prepareAuthenticatorTest()
		authenticator := MiddlewareV5[*TestUser](&BasicAuthenticatorV5{Optional: true})

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "wrong password")
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.ResponseV5, request *goyave.RequestV5) {
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		_ = resp.Body.Close()
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.invalid-credentials")}, body)
	})

	t.Run("optional_no_auth", func(t *testing.T) {
		server, _ := prepareAuthenticatorTest()
		authenticator := MiddlewareV5[*TestUser](&BasicAuthenticatorV5{Optional: true})

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.ResponseV5, request *goyave.RequestV5) {
			assert.Nil(t, request.User)
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close()
	})

	t.Run("error_no_table", func(t *testing.T) {
		assert.Panics(t, func() {
			cfg := config.LoadDefault()
			cfg.Set("database.connection", "sqlite3")
			cfg.Set("database.name", "testbasicauthenticator_no_table.db")
			cfg.Set("database.options", "mode=memory")
			cfg.Set("app.debug", false)
			server, err := testutil.NewTestServerWithConfig(cfg, nil)
			if err != nil {
				assert.Fail(t, err.Error())
				return
			}
			authenticator := MiddlewareV5[*TestUserPromoted](&BasicAuthenticatorV5{})
			authenticator.Init(server.Server)
			request := server.NewTestRequest(http.MethodGet, "/protected", nil)
			request.Request().SetBasicAuth("johndoe", "secret")

			// Panic here because table doesn't exist
			user := &TestUserPromoted{}
			_ = authenticator.Authenticate(request, &user)
		})
	})

	t.Run("no_auth", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("database.connection", "sqlite3")
		cfg.Set("database.name", "testbasicauthenticator_no_table.db")
		cfg.Set("database.options", "mode=memory")
		cfg.Set("app.debug", false)
		server, err := testutil.NewTestServerWithConfig(cfg, nil)
		if err != nil {
			assert.Fail(t, err.Error())
			return
		}
		authenticator := MiddlewareV5[*TestUser](&BasicAuthenticatorV5{})
		authenticator.Init(server.Server)
		request := server.NewTestRequest(http.MethodGet, "/protected", nil)

		err = authenticator.Authenticate(request, &TestUserPromoted{})
		assert.Error(t, err)
		assert.Equal(t, server.Lang.GetDefault().Get("auth.no-credentials-provided"), err.Error())
	})
}

func TestConfigBasicAuthenticator(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("auth.basic.username", "johndoe")
		cfg.Set("auth.basic.password", "secret")
		server, err := testutil.NewTestServerWithConfig(cfg, nil)
		if err != nil {
			assert.Fail(t, err.Error())
			return
		}
		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth("johndoe", "secret")
		resp := server.TestMiddleware(ConfigBasicAuthV5(), request, func(response *goyave.ResponseV5, request *goyave.RequestV5) {
			assert.Equal(t, "johndoe", request.User.(*BasicUser).Name)
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close()
	})

	t.Run("wrong_password", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("auth.basic.username", "johndoe")
		cfg.Set("auth.basic.password", "secret")
		server, err := testutil.NewTestServerWithConfig(cfg, nil)
		if err != nil {
			assert.Fail(t, err.Error())
			return
		}
		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth("johndoe", "wrong_password")
		resp := server.TestMiddleware(ConfigBasicAuthV5(), request, func(response *goyave.ResponseV5, request *goyave.RequestV5) {
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		_ = resp.Body.Close()
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.invalid-credentials")}, body)
	})

	t.Run("no_auth", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("auth.basic.username", "johndoe")
		cfg.Set("auth.basic.password", "secret")
		server, err := testutil.NewTestServerWithConfig(cfg, nil)
		if err != nil {
			assert.Fail(t, err.Error())
			return
		}
		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		resp := server.TestMiddleware(ConfigBasicAuthV5(), request, func(response *goyave.ResponseV5, request *goyave.RequestV5) {
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		_ = resp.Body.Close()
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.no-credentials-provided")}, body)
	})
}
