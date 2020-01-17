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

type BasicAuthenticatableTestSuite struct {
	goyave.TestSuite
}

func (suite *BasicAuthenticatableTestSuite) SetupSuite() {
	config.Set("dbConnection", "mysql")
	database.ClearRegisteredModels()
	database.RegisterModel(&TestUser{})

	database.Migrate()

}

func (suite *BasicAuthenticatableTestSuite) SetupTest() {
	user := &TestUser{
		Name:     "Admin",
		Password: "$2y$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
		Email:    "johndoe@example.org",
	}
	database.GetConnection().Create(user)
}

func (suite *BasicAuthenticatableTestSuite) createRequest(username, password string) *goyave.Request {
	request := suite.CreateTestRequest(httptest.NewRequest("GET", "/", nil))
	request.Header().Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
	return request
}

func (suite *BasicAuthenticatableTestSuite) TestAuthenticate() {
	user := &TestUser{}
	suite.True(user.Authenticate(suite.createRequest("johndoe@example.org", "password"), user))
	suite.Equal("Admin", user.Name)

	user = &TestUser{}
	suite.False(user.Authenticate(suite.createRequest("johndoe@example.org", "wrong password"), user))
	user = &TestUser{}
	suite.False(user.Authenticate(suite.createRequest("wrongemail@example.org", "password"), user))

	user = &TestUser{}
	request := suite.CreateTestRequest(httptest.NewRequest("GET", "/", nil))
	request.Header().Set("Authorization", "Basic")
	suite.False(user.Authenticate(request, user))

	suite.Panics(func() {
		userNoTable := &TestUserPromoted{}
		userNoTable.Authenticate(suite.createRequest("johndoe@example.org", "password"), userNoTable)
	})
}

func (suite *BasicAuthenticatableTestSuite) TearDownTest() {
	suite.ClearDatabase()
}

func (suite *BasicAuthenticatableTestSuite) TearDownSuite() {
	database.GetConnection().Exec("DROP TABLE test_users;")
	database.ClearRegisteredModels()
}

func TestBasicAuthenticatableSuite(t *testing.T) {
	goyave.RunTest(t, new(BasicAuthenticatableTestSuite))
}
