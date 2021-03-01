package auth

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"goyave.dev/goyave/v3"
	"goyave.dev/goyave/v3/config"
	"goyave.dev/goyave/v3/database"
)

type JWTAuthenticatorTestSuite struct {
	goyave.TestSuite
}

func (suite *JWTAuthenticatorTestSuite) SetupSuite() {
	config.Set("database.connection", "mysql")
	database.ClearRegisteredModels()
	database.RegisterModel(&TestUser{})

	database.Migrate()
}

func (suite *JWTAuthenticatorTestSuite) SetupTest() {
	user := &TestUser{
		Name:  "Admin",
		Email: "johndoe@example.org",
	}
	database.GetConnection().Create(user)
}

func (suite *JWTAuthenticatorTestSuite) createRequest(token string) *goyave.Request {
	request := suite.CreateTestRequest(httptest.NewRequest("GET", "/", nil))
	request.Header().Set("Authorization", "Bearer "+token)
	return request
}

func (suite *JWTAuthenticatorTestSuite) createWrongToken(method jwt.SigningMethod, userid string, nbf time.Time, exp time.Time) (string, error) {
	token := jwt.NewWithClaims(method, jwt.MapClaims{
		"userid": userid,
		"nbf":    nbf.Unix(), // Not Before
		"exp":    exp.Unix(), // Expiry
	})

	return token.SignedString([]byte(config.GetString("auth.jwt.secret")))
}

func (suite *JWTAuthenticatorTestSuite) TestAuthenticate() {
	user := &TestUser{}
	tokenAuthenticator := &JWTAuthenticator{}
	token, err := GenerateToken("johndoe@example.org")
	suite.Nil(err)
	suite.Nil(tokenAuthenticator.Authenticate(suite.createRequest(token), user))
	suite.Equal("Admin", user.Name)

	user = &TestUser{}
	token, err = GenerateToken("wrongemail@example.org")
	suite.Nil(err)
	suite.Equal("These credentials don't match our records.", tokenAuthenticator.Authenticate(suite.createRequest(token), user).Error())

	user = &TestUser{}
	request := suite.CreateTestRequest(nil)
	request.Header().Set("Authorization", "Basic userauthtoken")
	suite.Equal("Invalid or missing authentication header.", tokenAuthenticator.Authenticate(request, user).Error())

	userNoTable := &TestUserPromoted{}
	suite.Equal("Your authentication token is invalid.", tokenAuthenticator.Authenticate(suite.createRequest("userauthtoken"), userNoTable).Error())

	suite.Panics(func() {
		userNoTable := &TestUserPromoted{}
		token, err = GenerateToken("wrongemail@example.org")
		suite.Nil(err)
		if err := tokenAuthenticator.Authenticate(suite.createRequest(token), userNoTable); err != nil {
			suite.Fail(err.Error())
		}
	})

	user = &TestUser{}
	nbf := time.Now().Add(5 * time.Minute)
	token, err = suite.createWrongToken(jwt.SigningMethodHS256, "johndoe@example.org", nbf, nbf)
	suite.Nil(err)
	suite.Equal("Your authentication token is not valid yet.", tokenAuthenticator.Authenticate(suite.createRequest(token), user).Error())

	user = &TestUser{}
	nbf = time.Now()
	exp := nbf.Add(-5 * time.Minute)
	token, err = suite.createWrongToken(jwt.SigningMethodHS256, "johndoe@example.org", nbf, exp)
	suite.Nil(err)
	suite.Equal("Your authentication token is expired.", tokenAuthenticator.Authenticate(suite.createRequest(token), user).Error())
}

func (suite *JWTAuthenticatorTestSuite) TestOptional() {
	tokenAuthenticator := &JWTAuthenticator{Optional: true}
	suite.Nil(tokenAuthenticator.Authenticate(suite.CreateTestRequest(httptest.NewRequest("GET", "/", nil)), nil))
}

func (suite *JWTAuthenticatorTestSuite) TearDownTest() {
	suite.ClearDatabase()
}

func (suite *JWTAuthenticatorTestSuite) TearDownSuite() {
	database.Conn().Migrator().DropTable(&TestUser{})
	database.ClearRegisteredModels()
}

func TestJWTAuthenticatorSuite(t *testing.T) {
	goyave.RunTest(t, new(JWTAuthenticatorTestSuite))
}
