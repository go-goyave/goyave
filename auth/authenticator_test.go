package auth

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"gorm.io/gorm"
	"goyave.dev/goyave/v4"
	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/database"

	_ "goyave.dev/goyave/v4/database/dialect/mysql"
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

type AuthenticationTestSuite struct {
	goyave.TestSuite
}

func (suite *AuthenticationTestSuite) SetupSuite() {
	config.Set("database.connection", "mysql")
	database.ClearRegisteredModels()
	database.RegisterModel(&TestUser{})

	database.Migrate()
}

func (suite *AuthenticationTestSuite) SetupTest() {
	user := &TestUser{
		Name:     "Admin",
		Password: "$2y$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
		Email:    "johndoe@example.org",
	}
	database.GetConnection().Create(user)
}

func (suite *AuthenticationTestSuite) TestFindColumns() {
	user := &TestUser{}
	fields := FindColumns(user, "username", "password")
	suite.Len(fields, 2)
	suite.Equal("email", fields[0].Name)
	suite.Equal("password", fields[1].Name)

	fields = FindColumns(user, "username", "notatag", "password")
	suite.Len(fields, 3)
	suite.Equal("email", fields[0].Name)
	suite.Nil(fields[1])
	suite.Equal("password", fields[2].Name)

	userOverride := &TestUserOverride{}
	fields = FindColumns(userOverride, "password")
	suite.Len(fields, 1)
	suite.Equal("password_override", fields[0].Name)

	userInvalidOverride := &TestUserInvalidOverride{}
	fields = FindColumns(userInvalidOverride, "password")
	suite.Len(fields, 1)
	suite.Equal("password", fields[0].Name)
}

func (suite *AuthenticationTestSuite) TestFindColumnsPromoted() {
	user := &TestUserPromoted{}
	fields := FindColumns(user, "username", "password")
	suite.Len(fields, 2)
	suite.Equal("email", fields[0].Name)
	suite.Equal("password", fields[1].Name)

	fields = FindColumns(user, "username", "notatag", "password")
	suite.Len(fields, 3)
	suite.Equal("email", fields[0].Name)
	suite.Nil(fields[1])
	suite.Equal("password", fields[2].Name)

	userPtr := &TestUserPromotedPtr{}
	fields = FindColumns(userPtr, "username", "password")
	suite.Len(fields, 2)
	suite.Equal("email", fields[0].Name)
	suite.Equal("password", fields[1].Name)
}

func (suite *AuthenticationTestSuite) TestAuthMiddleware() {
	// Test middleware with BasicAuth
	authenticator := Middleware(&TestUser{}, &BasicAuthenticator{})

	request := suite.CreateTestRequest(httptest.NewRequest("GET", "/", nil))
	request.Header().Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("johndoe@example.org:wrong_password")))
	result := suite.Middleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
		suite.Fail("Auth middleware passed")
	})
	result.Body.Close()
	suite.Equal(http.StatusUnauthorized, result.StatusCode)

	request = suite.CreateTestRequest(httptest.NewRequest("GET", "/", nil))
	request.Header().Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("johndoe@example.org:password")))
	result = suite.Middleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
		suite.IsType(&TestUser{}, request.User)
		suite.Equal("Admin", request.User.(*TestUser).Name)
		response.Status(200)
	})
	result.Body.Close()
	suite.Equal(200, result.StatusCode)
}

func (suite *AuthenticationTestSuite) TestAuthMiddlewareUnauthorizer() {
	authenticator := Middleware(&TestUser{}, &TestBasicUnauthorizer{})

	request := suite.CreateTestRequest(httptest.NewRequest("GET", "/", nil))
	request.Header().Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("johndoe@example.org:wrong_password")))
	result := suite.Middleware(authenticator, request, func(response *goyave.Response, request *goyave.Request) {
		suite.Fail("Auth middleware passed")
	})
	defer result.Body.Close()

	data := map[string]interface{}{}
	if suite.Nil(suite.GetJSONBody(result, &data)) {
		suite.Contains(data, "custom error key")
	}
	suite.Equal(http.StatusUnauthorized, result.StatusCode)
}

func (suite *AuthenticationTestSuite) TearDownTest() {
	suite.ClearDatabase()
}

func (suite *AuthenticationTestSuite) TearDownSuite() {
	database.Conn().Migrator().DropTable(&TestUser{})
	database.ClearRegisteredModels()
}

func TestAuthenticationSuite(t *testing.T) {
	goyave.RunTest(t, new(AuthenticationTestSuite))
}
