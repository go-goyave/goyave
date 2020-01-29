package auth

import (
	"encoding/base64"
	"net/http/httptest"
	"testing"

	"github.com/System-Glitch/goyave/v2/config"

	"github.com/System-Glitch/goyave/v2"
	"github.com/System-Glitch/goyave/v2/database"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type BasicAuthenticatorTestSuite struct {
	goyave.TestSuite
}

func (suite *BasicAuthenticatorTestSuite) SetupSuite() {
	config.Set("dbConnection", "mysql")
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
	suite.Equal("These credentials don't match our records.", basicAuthenticator.Authenticate(suite.createRequest("johndoe@example.org", "wrong password"), user).Error())
	user = &TestUser{}
	suite.Equal("These credentials don't match our records.", basicAuthenticator.Authenticate(suite.createRequest("wrongemail@example.org", "password"), user).Error())

	user = &TestUser{}
	request := suite.CreateTestRequest(httptest.NewRequest("GET", "/", nil))
	request.Header().Set("Authorization", "Basic")
	suite.Equal("Invalid or missing authentication header.", basicAuthenticator.Authenticate(request, user).Error())

	suite.Panics(func() {
		userNoTable := &TestUserPromoted{}
		basicAuthenticator.Authenticate(suite.createRequest("johndoe@example.org", "password"), userNoTable)
	})
}

func (suite *BasicAuthenticatorTestSuite) TestAuthenticateViaConfig() {
	config.Set("authUsername", "admin")
	config.Set("authPassword", "secret")

	authenticator := ConfigBasicAuth()
	request := suite.CreateTestRequest(httptest.NewRequest("GET", "/", nil))
	request.Header().Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:wrong_password")))
	result := suite.Middleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
		suite.Fail("Auth middleware passed")
	})
	suite.Equal(401, result.StatusCode)

	request = suite.CreateTestRequest(httptest.NewRequest("GET", "/", nil))
	request.Header().Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:secret")))
	result = suite.Middleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
		suite.IsType(&BasicUser{}, request.User)
		suite.Equal("admin", request.User.(*BasicUser).Name)
		response.Status(200)
	})
	suite.Equal(200, result.StatusCode)
}

func (suite *BasicAuthenticatorTestSuite) TearDownTest() {
	suite.ClearDatabase()
}

func (suite *BasicAuthenticatorTestSuite) TearDownSuite() {
	database.GetConnection().Exec("DROP TABLE test_users;")
	database.ClearRegisteredModels()
}

func TestBasicAuthenticatorSuite(t *testing.T) {
	goyave.RunTest(t, new(BasicAuthenticatorTestSuite))
}
