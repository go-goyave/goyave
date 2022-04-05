package auth

import (
	"encoding/base64"
	"net/http/httptest"
	"testing"

	"goyave.dev/goyave/v4/config"

	"goyave.dev/goyave/v4"
	"goyave.dev/goyave/v4/database"

	_ "goyave.dev/goyave/v4/database/dialect/mysql"
)

type BasicAuthenticatorTestSuite struct {
	goyave.TestSuite
}

func (suite *BasicAuthenticatorTestSuite) SetupSuite() {
	config.Set("database.connection", "mysql")
	database.ClearRegisteredModels()
	database.RegisterModel(&TestUser{})

	database.Migrate()

}

func (suite *BasicAuthenticatorTestSuite) SetupTest() {
	user := &TestUser{
		Name:     "Admin",
		Password: "$2y$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
		Email:    "johndoe@example.org",
	}
	database.GetConnection().Create(user)
}

func (suite *BasicAuthenticatorTestSuite) createRequest(username, password string) *goyave.Request {
	request := suite.CreateTestRequest(httptest.NewRequest("GET", "/", nil))
	request.Header().Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
	return request
}

func (suite *BasicAuthenticatorTestSuite) TestAuthenticate() {
	user := &TestUser{}
	basicAuthenticator := &BasicAuthenticator{}
	suite.Nil(basicAuthenticator.Authenticate(suite.createRequest("johndoe@example.org", "password"), user))
	suite.Equal("Admin", user.Name)

	user = &TestUser{}
	suite.Equal("Invalid credentials.", basicAuthenticator.Authenticate(suite.createRequest("johndoe@example.org", "wrong password"), user).Error())
	user = &TestUser{}
	suite.Equal("Invalid credentials.", basicAuthenticator.Authenticate(suite.createRequest("wrongemail@example.org", "password"), user).Error())

	user = &TestUser{}
	request := suite.CreateTestRequest(httptest.NewRequest("GET", "/", nil))
	request.Header().Set("Authorization", "Basic")
	suite.Equal("Invalid or missing authentication header.", basicAuthenticator.Authenticate(request, user).Error())

	suite.Panics(func() {
		userNoTable := &TestUserPromoted{}
		if err := basicAuthenticator.Authenticate(suite.createRequest("johndoe@example.org", "password"), userNoTable); err != nil {
			suite.Fail(err.Error())
		}
	})
}

func (suite *BasicAuthenticatorTestSuite) TestOptional() {
	basicAuthenticator := &BasicAuthenticator{Optional: true}
	suite.Nil(basicAuthenticator.Authenticate(suite.CreateTestRequest(httptest.NewRequest("GET", "/", nil)), nil))
}

func (suite *BasicAuthenticatorTestSuite) TestAuthenticateViaConfig() {
	config.Set("auth.basic.username", "admin")
	config.Set("auth.basic.password", "secret")

	authenticator := ConfigBasicAuth()
	request := suite.CreateTestRequest(httptest.NewRequest("GET", "/", nil))
	request.Header().Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:wrong_password")))
	result := suite.Middleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
		suite.Fail("Auth middleware passed")
	})
	result.Body.Close()
	suite.Equal(401, result.StatusCode)

	request = suite.CreateTestRequest(httptest.NewRequest("GET", "/", nil))
	request.Header().Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:secret")))
	result = suite.Middleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
		suite.IsType(&BasicUser{}, request.User)
		suite.Equal("admin", request.User.(*BasicUser).Name)
		response.Status(200)
	})
	result.Body.Close()
	suite.Equal(200, result.StatusCode)
}

func (suite *BasicAuthenticatorTestSuite) TearDownTest() {
	suite.ClearDatabase()
}

func (suite *BasicAuthenticatorTestSuite) TearDownSuite() {
	database.Conn().Migrator().DropTable(&TestUser{})
	database.ClearRegisteredModels()
}

func TestBasicAuthenticatorSuite(t *testing.T) {
	goyave.RunTest(t, new(BasicAuthenticatorTestSuite))
}
