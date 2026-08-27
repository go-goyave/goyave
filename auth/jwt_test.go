package auth

import (
	"bytes"
	"fmt"
	"net/http"
	"path"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/util/fsutil/osfs"
	"goyave.dev/goyave/v5/util/testutil"
)

func TestJWTService(t *testing.T) {
	t.Run("GenerateToken", func(t *testing.T) {
		config := &JWTConfig{
			Expiry: 20,
			Secret: "secret",
		}
		service := NewJWTService(config, &osfs.FS{})

		now := time.Now()
		expiry := time.Duration(20) * time.Second

		tokenString, err := service.GenerateToken("johndoe")
		require.NoError(t, err)
		parsedToken, err := jwt.Parse(tokenString, func(_ *jwt.Token) (any, error) {
			return []byte(config.Secret), nil
		})

		require.NoError(t, err)
		assert.True(t, parsedToken.Valid)
		assert.Equal(t, jwt.SigningMethodHS256, parsedToken.Method)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if assert.True(t, ok) {
			assert.Equal(t, "johndoe", claims["sub"])
			assert.GreaterOrEqual(t, float64(now.Unix()), claims["nbf"])
			assert.True(t, time.Unix(int64(claims["exp"].(float64)), 0).After(now))
			assert.Equal(t, int64(expiry.Seconds()), int64(claims["exp"].(float64)-claims["nbf"].(float64)))
		}
	})

	t.Run("GenerateTokenWithClaims_HS256", func(t *testing.T) {
		config := &JWTConfig{
			Expiry: 20,
			Secret: "secret",
		}
		service := NewJWTService(config, &osfs.FS{})

		now := time.Now()
		expiry := time.Duration(20) * time.Second

		srcClaims := jwt.MapClaims{
			"sub":         "johndoe",
			"customClaim": "customValue",
		}
		tokenString, err := service.GenerateTokenWithClaims(srcClaims, jwt.SigningMethodHS256)
		require.NoError(t, err)
		parsedToken, err := jwt.Parse(tokenString, func(_ *jwt.Token) (any, error) {
			return []byte(config.Secret), nil
		})

		require.NoError(t, err)
		assert.True(t, parsedToken.Valid)
		assert.Equal(t, jwt.SigningMethodHS256, parsedToken.Method)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if assert.True(t, ok) {
			assert.Equal(t, "johndoe", claims["sub"])
			assert.Equal(t, "customValue", claims["customClaim"])
			assert.GreaterOrEqual(t, float64(now.Unix()), claims["nbf"])
			assert.True(t, time.Unix(int64(claims["exp"].(float64)), 0).After(now))
			assert.Equal(t, int64(expiry.Seconds()), int64(claims["exp"].(float64)-claims["nbf"].(float64)))
		}
	})

	t.Run("GenerateTokenWithClaims_RSA", func(t *testing.T) {
		rootDir := testutil.FindRootDirectory()
		config := &JWTConfig{
			Expiry: 20,
			RSA: KeyPairConfig{
				Public:  path.Join(rootDir, "resources/rsa/public.pem"),
				Private: path.Join(rootDir, "resources/rsa/private.pem"),
			},
		}
		service := NewJWTService(config, &osfs.FS{})

		now := time.Now()
		expiry := time.Duration(20) * time.Second

		srcClaims := jwt.MapClaims{
			"sub":         "johndoe",
			"customClaim": "customValue",
		}
		tokenString, err := service.GenerateTokenWithClaims(srcClaims, jwt.SigningMethodRS256)
		require.NoError(t, err)
		parsedToken, err := jwt.Parse(tokenString, func(_ *jwt.Token) (any, error) {
			return service.GetPublicKey(jwt.SigningMethodRS256)
		})

		require.NoError(t, err)
		assert.True(t, parsedToken.Valid)
		assert.Equal(t, jwt.SigningMethodRS256, parsedToken.Method)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if assert.True(t, ok) {
			assert.Equal(t, "johndoe", claims["sub"])
			assert.Equal(t, "customValue", claims["customClaim"])
			assert.GreaterOrEqual(t, float64(now.Unix()), claims["nbf"])
			assert.True(t, time.Unix(int64(claims["exp"].(float64)), 0).After(now))
			assert.Equal(t, int64(expiry.Seconds()), int64(claims["exp"].(float64)-claims["nbf"].(float64)))
		}
	})

	t.Run("GenerateTokenWithClaims_ECDSA", func(t *testing.T) {
		rootDir := testutil.FindRootDirectory()
		config := &JWTConfig{
			Expiry: 20,
			RSA: KeyPairConfig{
				Public:  path.Join(rootDir, "resources/ecdsa/public.pem"),
				Private: path.Join(rootDir, "resources/ecdsa/private.pem"),
			},
		}
		service := NewJWTService(config, &osfs.FS{})

		now := time.Now()
		expiry := time.Duration(20) * time.Second

		srcClaims := jwt.MapClaims{
			"sub":         "johndoe",
			"customClaim": "customValue",
		}
		tokenString, err := service.GenerateTokenWithClaims(srcClaims, jwt.SigningMethodES256)
		require.NoError(t, err)
		parsedToken, err := jwt.Parse(tokenString, func(_ *jwt.Token) (any, error) {
			return service.GetPublicKey(jwt.SigningMethodES256)
		})

		require.NoError(t, err)
		assert.True(t, parsedToken.Valid)
		assert.Equal(t, jwt.SigningMethodES256, parsedToken.Method)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if assert.True(t, ok) {
			assert.Equal(t, "johndoe", claims["sub"])
			assert.Equal(t, "customValue", claims["customClaim"])
			assert.GreaterOrEqual(t, float64(now.Unix()), claims["nbf"])
			assert.True(t, time.Unix(int64(claims["exp"].(float64)), 0).After(now))
			assert.Equal(t, int64(expiry.Seconds()), int64(claims["exp"].(float64)-claims["nbf"].(float64)))
		}
	})

	t.Run("GenerateTokenWithClaims_Unsupported", func(t *testing.T) {
		config := &JWTConfig{
			Expiry: 20,
		}
		service := NewJWTService(config, &osfs.FS{})

		_, err := service.GenerateTokenWithClaims(nil, jwt.SigningMethodPS256)
		require.Error(t, err)
	})
}

func TestJWTAuthenticator(t *testing.T) {
	t.Run("success_hs256", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)

		config := &JWTConfig{
			Expiry: 20,
			Secret: "secret",
		}
		jwtService := NewJWTService(config, &osfs.FS{})
		mockUserService := &MockUserService[TestUser]{user: user}
		authenticator := Middleware(NewJWTAuthenticator(jwtService, mockUserService))

		token, err := jwtService.GenerateToken(user.Email)
		require.NoError(t, err)

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().Header.Set("Authorization", "Bearer "+token)
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Equal(t, user.ID, request.User.(*TestUser).ID)
			assert.Equal(t, user.Name, request.User.(*TestUser).Name)
			assert.Equal(t, user.Email, request.User.(*TestUser).Email)
			assert.Contains(t, request.Extra, ExtraJWTClaims{})
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
		assert.Empty(t, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("success_rsa", func(t *testing.T) {
		rootDir := testutil.FindRootDirectory()
		server, user := prepareAuthenticatorTest(t)
		config := &JWTConfig{
			Expiry: 20,
			RSA: KeyPairConfig{
				Public:  path.Join(rootDir, "resources/rsa/public.pem"),
				Private: path.Join(rootDir, "resources/rsa/private.pem"),
			},
		}
		jwtService := NewJWTService(config, &osfs.FS{})
		mockUserService := &MockUserService[TestUser]{user: user}
		a := NewJWTAuthenticator(jwtService, mockUserService)
		a.SigningMethod = jwt.SigningMethodRS256
		authenticator := Middleware(a)

		token, err := jwtService.GenerateTokenWithClaims(jwt.MapClaims{"sub": user.Email}, jwt.SigningMethodRS256)
		require.NoError(t, err)

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().Header.Set("Authorization", "Bearer "+token)
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Equal(t, user.ID, request.User.(*TestUser).ID)
			assert.Equal(t, user.Name, request.User.(*TestUser).Name)
			assert.Equal(t, user.Email, request.User.(*TestUser).Email)
			assert.Contains(t, request.Extra, ExtraJWTClaims{})
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
		assert.Empty(t, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("success_ecdsa", func(t *testing.T) {
		rootDir := testutil.FindRootDirectory()
		server, user := prepareAuthenticatorTest(t)
		config := &JWTConfig{
			Expiry: 20,
			RSA: KeyPairConfig{
				Public:  path.Join(rootDir, "resources/ecdsa/public.pem"),
				Private: path.Join(rootDir, "resources/ecdsa/private.pem"),
			},
		}
		jwtService := NewJWTService(config, &osfs.FS{})

		mockUserService := &MockUserService[TestUser]{user: user}
		a := NewJWTAuthenticator(jwtService, mockUserService)
		a.SigningMethod = jwt.SigningMethodES256
		authenticator := Middleware(a)

		token, err := jwtService.GenerateTokenWithClaims(jwt.MapClaims{"sub": user.Email}, jwt.SigningMethodES256)
		require.NoError(t, err)

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().Header.Set("Authorization", "Bearer "+token)
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Equal(t, user.ID, request.User.(*TestUser).ID)
			assert.Equal(t, user.Name, request.User.(*TestUser).Name)
			assert.Equal(t, user.Email, request.User.(*TestUser).Email)
			assert.Contains(t, request.Extra, ExtraJWTClaims{})
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
		assert.Empty(t, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("invalid_token", func(t *testing.T) {
		server, _ := prepareAuthenticatorTest(t)
		config := &JWTConfig{
			Expiry: 20,
			Secret: "secret",
		}
		jwtService := NewJWTService(config, &osfs.FS{})
		mockUserService := &MockUserService[TestUser]{}
		authenticator := Middleware(NewJWTAuthenticator(jwtService, mockUserService))

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().Header.Set("Authorization", "Bearer invalidtoken")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Nil(t, request.User)
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.jwt-invalid")}, body)
		assert.Equal(t, `Bearer realm="Authorization required", charset="UTF-8"`, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("token_not_valid_yet", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		config := &JWTConfig{
			Expiry: 20,
			Secret: "secret",
		}
		jwtService := NewJWTService(config, &osfs.FS{})
		mockUserService := &MockUserService[TestUser]{}
		authenticator := Middleware(NewJWTAuthenticator(jwtService, mockUserService))

		token, err := jwtService.GenerateTokenWithClaims(jwt.MapClaims{
			"sub": user.Email,
			"nbf": time.Now().Add(time.Hour).Unix(),
		}, jwt.SigningMethodHS256)
		require.NoError(t, err)

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().Header.Set("Authorization", "Bearer "+token)
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Nil(t, request.User)
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.jwt-not-valid-yet")}, body)
		assert.Equal(t, `Bearer realm="Authorization required", charset="UTF-8"`, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("token_expired", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		config := &JWTConfig{
			Expiry: 20,
			Secret: "secret",
		}
		jwtService := NewJWTService(config, &osfs.FS{})
		mockUserService := &MockUserService[TestUser]{}
		authenticator := Middleware(NewJWTAuthenticator(jwtService, mockUserService))

		token, err := jwtService.GenerateTokenWithClaims(jwt.MapClaims{
			"sub": user.Email,
			"exp": time.Now().Add(-time.Hour).Unix(),
		}, jwt.SigningMethodHS256)
		require.NoError(t, err)

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().Header.Set("Authorization", "Bearer "+token)
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Nil(t, request.User)
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.jwt-expired")}, body)
		assert.Equal(t, `Bearer realm="Authorization required", charset="UTF-8"`, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("unknown_user", func(t *testing.T) {
		server, _ := prepareAuthenticatorTest(t)
		config := &JWTConfig{
			Expiry: 20,
			Secret: "secret",
		}
		jwtService := NewJWTService(config, &osfs.FS{})
		mockUserService := &MockUserService[TestUser]{
			err: gorm.ErrRecordNotFound,
		}
		authenticator := Middleware(NewJWTAuthenticator(jwtService, mockUserService))

		token, err := jwtService.GenerateToken("notjohndoe@example.org")
		require.NoError(t, err)

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().Header.Set("Authorization", "Bearer "+token)
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Nil(t, request.User)
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.invalid-credentials")}, body)
		assert.Equal(t, `Bearer realm="Authorization required", charset="UTF-8"`, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("service_error", func(t *testing.T) {
		server, _ := prepareAuthenticatorTest(t)
		buf := &bytes.Buffer{}
		server.Logger = slog.New(slog.NewHandler(false, buf))
		config := &JWTConfig{
			Expiry: 20,
			Secret: "secret",
		}
		jwtService := NewJWTService(config, &osfs.FS{})
		mockUserService := &MockUserService[TestUser]{err: fmt.Errorf("service_error")}
		authenticator := Middleware(NewJWTAuthenticator(jwtService, mockUserService))

		token, err := jwtService.GenerateToken("notjohndoe@example.org")
		require.NoError(t, err)

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().Header.Set("Authorization", "Bearer "+token)
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, _ *goyave.Request) {
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
		assert.Empty(t, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("unexpected_method_hmac", func(t *testing.T) {
		rootDir := testutil.FindRootDirectory()
		server, _ := prepareAuthenticatorTest(t)
		config := &JWTConfig{
			Expiry: 20,
			RSA: KeyPairConfig{
				Public:  path.Join(rootDir, "resources/rsa/public.pem"),
				Private: path.Join(rootDir, "resources/rsa/private.pem"),
			},
		}
		jwtService := NewJWTService(config, &osfs.FS{})
		mockUserService := &MockUserService[TestUser]{}
		a := NewJWTAuthenticator(jwtService, mockUserService)
		a.SigningMethod = jwt.SigningMethodHS256
		authenticator := Middleware(a)

		token, err := jwtService.GenerateTokenWithClaims(jwt.MapClaims{"sub": "johndoe@example.org"}, jwt.SigningMethodRS256)
		require.NoError(t, err)

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().Header.Set("Authorization", "Bearer "+token)
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Nil(t, request.User)
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.jwt-invalid")}, body)
		assert.Equal(t, `Bearer realm="Authorization required", charset="UTF-8"`, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("unexpected_method_rsa", func(t *testing.T) {
		server, _ := prepareAuthenticatorTest(t)
		config := &JWTConfig{
			Expiry: 20,
			Secret: "secret",
		}
		jwtService := NewJWTService(config, &osfs.FS{})
		mockUserService := &MockUserService[TestUser]{}
		a := NewJWTAuthenticator(jwtService, mockUserService)
		a.SigningMethod = jwt.SigningMethodRS256
		authenticator := Middleware(a)

		token, err := jwtService.GenerateToken("johndoe@example.org")
		require.NoError(t, err)

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().Header.Set("Authorization", "Bearer "+token)
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Nil(t, request.User)
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.jwt-invalid")}, body)
		assert.Equal(t, `Bearer realm="Authorization required", charset="UTF-8"`, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("unexpected_method_ecdsa", func(t *testing.T) {
		server, _ := prepareAuthenticatorTest(t)
		config := &JWTConfig{
			Expiry: 20,
			Secret: "secret",
		}
		jwtService := NewJWTService(config, &osfs.FS{})
		mockUserService := &MockUserService[TestUser]{}
		a := NewJWTAuthenticator(jwtService, mockUserService)
		a.SigningMethod = jwt.SigningMethodES256
		authenticator := Middleware(a)

		token, err := jwtService.GenerateToken("johndoe@example.org")
		require.NoError(t, err)

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().Header.Set("Authorization", "Bearer "+token)
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Nil(t, request.User)
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.jwt-invalid")}, body)
		assert.Equal(t, `Bearer realm="Authorization required", charset="UTF-8"`, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("unsupported_method", func(t *testing.T) {
		server, _ := prepareAuthenticatorTest(t)
		config := &JWTConfig{
			Expiry: 20,
			Secret: "secret",
		}
		jwtService := NewJWTService(config, &osfs.FS{})
		mockUserService := &MockUserService[TestUser]{}
		a := NewJWTAuthenticator(jwtService, mockUserService)
		a.SigningMethod = jwt.SigningMethodPS256
		authenticator := Middleware(a)

		token, err := jwtService.GenerateToken("johndoe@example.org")
		require.NoError(t, err)

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().Header.Set("Authorization", "Bearer "+token)
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		assert.Panics(t, func() {
			_, _ = authenticator.Authenticate(request)
		})
	})

	t.Run("no_auth", func(t *testing.T) {
		server, _ := prepareAuthenticatorTest(t)
		config := &JWTConfig{
			Expiry: 20,
			Secret: "secret",
		}
		jwtService := NewJWTService(config, &osfs.FS{})
		mockUserService := &MockUserService[TestUser]{}
		authenticator := Middleware(NewJWTAuthenticator(jwtService, mockUserService))

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Nil(t, request.User)
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.no-credentials-provided")}, body)
		assert.Equal(t, `Bearer realm="Authorization required", charset="UTF-8"`, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("optional_success", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		config := &JWTConfig{
			Expiry: 20,
			Secret: "secret",
		}
		jwtService := NewJWTService(config, &osfs.FS{})
		mockUserService := &MockUserService[TestUser]{
			user: user,
		}
		a := NewJWTAuthenticator(jwtService, mockUserService)
		a.Optional = true
		authenticator := Middleware(a)

		token, err := jwtService.GenerateToken(user.Email)
		require.NoError(t, err)

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().Header.Set("Authorization", "Bearer "+token)
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Equal(t, user.ID, request.User.(*TestUser).ID)
			assert.Equal(t, user.Name, request.User.(*TestUser).Name)
			assert.Equal(t, user.Email, request.User.(*TestUser).Email)
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
		assert.Empty(t, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("optional_invalid_token", func(t *testing.T) {
		server, _ := prepareAuthenticatorTest(t)
		config := &JWTConfig{
			Expiry: 20,
			Secret: "secret",
		}
		jwtService := NewJWTService(config, &osfs.FS{})
		mockUserService := &MockUserService[TestUser]{}
		a := NewJWTAuthenticator(jwtService, mockUserService)
		a.Optional = true
		authenticator := MiddlewareWithRealm(a, "custom realm")

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().Header.Set("Authorization", "Bearer invalidtoken")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: true}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Nil(t, request.User)
			assert.Fail(t, "middleware passed despite failed authentication")
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		assert.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"error": server.Lang.GetDefault().Get("auth.jwt-invalid")}, body)
		assert.Equal(t, `Bearer realm="custom realm", charset="UTF-8"`, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("optional_no_auth", func(t *testing.T) {
		server, _ := prepareAuthenticatorTest(t)
		config := &JWTConfig{
			Expiry: 20,
			Secret: "secret",
		}
		jwtService := NewJWTService(config, &osfs.FS{})
		mockUserService := &MockUserService[TestUser]{}
		a := NewJWTAuthenticator(jwtService, mockUserService)
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
		assert.Empty(t, resp.Header.Get("WWW-Authenticate"))
	})
}
