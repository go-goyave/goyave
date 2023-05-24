package auth

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"
	"goyave.dev/goyave/v4"
	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/util/testutil"
	"goyave.dev/goyave/v4/util/typeutil"

	_ "goyave.dev/goyave/v4/database/dialect/sqlite"
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

func (a *TestBasicUnauthorizer) OnUnauthorized(response *goyave.ResponseV5, _ *goyave.RequestV5, err error) {
	response.JSON(http.StatusUnauthorized, map[string]string{"custom error key": err.Error()})
}

func prepareAuthenticatorTest() (*testutil.TestServer, *TestUser) {
	cfg := config.LoadDefault()
	cfg.Set("database.connection", "sqlite3")
	cfg.Set("database.name", "testauthenticator.db")
	cfg.Set("database.options", "mode=memory")
	cfg.Set("app.debug", false)
	server, err := testutil.NewTestServerWithConfig(cfg, nil)
	if err != nil {
		panic(err)
	}
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
		server, user := prepareAuthenticatorTest()

		authenticator := Middleware[*TestUser](&BasicAuthenticator{})

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

		request = server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "incorrect password")
		resp = server.TestMiddleware(authenticator, request, func(response *goyave.ResponseV5, _ *goyave.RequestV5) {
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

	t.Run("MiddlewareUnauthorizer", func(t *testing.T) {
		server, user := prepareAuthenticatorTest()

		authenticator := Middleware[*TestUser](&TestBasicUnauthorizer{})

		request := server.NewTestRequest(http.MethodGet, "/protected", nil)
		request.Request().SetBasicAuth(user.Email, "incorrect password")
		resp := server.TestMiddleware(authenticator, request, func(response *goyave.ResponseV5, request *goyave.RequestV5) {
			response.Status(http.StatusOK)
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		body, err := testutil.ReadJSONBody[map[string]string](resp.Body)
		_ = resp.Body.Close()
		if !assert.NoError(t, err) {
			return
		}
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
		{desc: "TestUser", model: &TestUser{}, input: []string{"username", "password"}, expected: []*string{typeutil.Ptr("email"), typeutil.Ptr("password")}},
		{desc: "TestUser_invalid_tag", model: &TestUser{}, input: []string{"username", "notatag", "password"}, expected: []*string{typeutil.Ptr("email"), nil, typeutil.Ptr("password")}},
		{desc: "TestUserOverride", model: &TestUserOverride{}, input: []string{"password"}, expected: []*string{typeutil.Ptr("password_override")}},
		{desc: "TestUserInvalidOverride", model: &TestUserInvalidOverride{}, input: []string{"password"}, expected: []*string{typeutil.Ptr("password")}},
		{desc: "TestUserPromoted", model: &TestUserPromoted{}, input: []string{"username", "password"}, expected: []*string{typeutil.Ptr("email"), typeutil.Ptr("password")}},
		{desc: "TestUserPromotedPtr", model: &TestUserPromotedPtr{}, input: []string{"username", "password"}, expected: []*string{typeutil.Ptr("email"), typeutil.Ptr("password")}},
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
