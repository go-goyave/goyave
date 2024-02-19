package auth

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"
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

type TestUserPromoted struct {
	TestUser
}

type TestUserPromotedPtr struct {
	*TestUser
}

type TestUserOverride struct {
	gorm.Model
	Name     string `gorm:"type:varchar(100)"`
	Password string `gorm:"type:varchar(100);column:password_override" auth:"password"`
	Email    string `gorm:"type:varchar(100);uniqueIndex" auth:"username"`
}

type TestUserInvalidOverride struct {
	gorm.Model
	Name     string `gorm:"type:varchar(100)"`
	Password string `gorm:"type:varchar(100);column:" auth:"password"`
	Email    string `gorm:"type:varchar(100);uniqueIndex" auth:"username"`
}

type TestBasicUnauthorizer struct {
	BasicAuthenticator
}

func (a *TestBasicUnauthorizer) OnUnauthorized(response *goyave.Response, _ *goyave.Request, err error) {
	response.JSON(http.StatusUnauthorized, map[string]string{"custom error key": err.Error()})
}

func prepareAuthenticatorTest(t *testing.T) (*testutil.TestServer, *TestUser) {
	cfg := config.LoadDefault()
	cfg.Set("database.connection", "sqlite3")
	cfg.Set("database.name", "testauthenticator.db")
	cfg.Set("database.options", "mode=memory")
	cfg.Set("app.debug", false)
	server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: cfg})
	db := server.DB()
	if err := db.AutoMigrate(&TestUser{}); err != nil {
		panic(err)
	}
	password, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	user := &TestUser{
		Name:     "johndoe",
		Email:    "johndoe@example.org",
		Password: string(password),
	}
	if err := db.Create(user).Error; err != nil {
		panic(err)
	}

	return server, user
}

func TestAuthenticator(t *testing.T) {

	t.Run("Middleware", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		t.Cleanup(func() { server.CloseDB() })

		authenticator := Middleware[*TestUser](&BasicAuthenticator{})

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
	})

	t.Run("NoAuth", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		t.Cleanup(func() { server.CloseDB() })

		authenticator := Middleware[*TestUser](&BasicAuthenticator{})

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "secret")
		request.Route = &goyave.Route{Meta: map[string]any{MetaAuth: false}}
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Nil(t, request.User)
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())

		request.Route = &goyave.Route{Meta: map[string]any{}}
		resp = server.TestMiddleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
			assert.Nil(t, request.User)
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())
	})

	t.Run("MiddlewareUnauthorizer", func(t *testing.T) {
		server, user := prepareAuthenticatorTest(t)
		t.Cleanup(func() { server.CloseDB() })

		authenticator := Middleware[*TestUser](&TestBasicUnauthorizer{})

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
	})
}

func TestFindColumns(t *testing.T) {
	db, _ := gorm.Open(tests.DummyDialector{})

	cases := []struct {
		desc     string
		model    any
		input    []string
		expected []*string
	}{
		{desc: "TestUser", model: &TestUser{}, input: []string{"username", "password"}, expected: []*string{lo.ToPtr("email"), lo.ToPtr("password")}},
		{desc: "TestUser_invalid_tag", model: &TestUser{}, input: []string{"username", "notatag", "password"}, expected: []*string{lo.ToPtr("email"), nil, lo.ToPtr("password")}},
		{desc: "TestUserOverride", model: &TestUserOverride{}, input: []string{"password"}, expected: []*string{lo.ToPtr("password_override")}},
		{desc: "TestUserInvalidOverride", model: &TestUserInvalidOverride{}, input: []string{"password"}, expected: []*string{lo.ToPtr("password")}},
		{desc: "TestUserPromoted", model: &TestUserPromoted{}, input: []string{"username", "password"}, expected: []*string{lo.ToPtr("email"), lo.ToPtr("password")}},
		{desc: "TestUserPromotedPtr", model: &TestUserPromotedPtr{}, input: []string{"username", "password"}, expected: []*string{lo.ToPtr("email"), lo.ToPtr("password")}},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			fields := FindColumns(db, c.model, c.input...)
			if assert.Len(t, fields, len(c.input)) {
				for i, f := range fields {
					expected := c.expected[i]
					message := fmt.Sprintf("index %d", i)
					if expected == nil {
						assert.Nil(t, f, message)
					} else {
						assert.NotNil(t, f.Field, message)
						assert.Equal(t, *expected, f.Name, message)
					}
				}
			}
		})
	}
}
