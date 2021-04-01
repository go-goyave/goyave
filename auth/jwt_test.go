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
	tokenAuthenticator *JWTAuthenticator
	user               *TestUser
}

func (suite *JWTAuthenticatorTestSuite) SetupSuite() {
	config.Set("database.connection", "mysql")
	database.ClearRegisteredModels()
	database.RegisterModel(&TestUser{})

	database.Migrate()
	suite.tokenAuthenticator = &JWTAuthenticator{}
}

func (suite *JWTAuthenticatorTestSuite) SetupTest() {
	suite.user = &TestUser{
		Name:  "Admin",
		Email: "johndoe@example.org",
	}
	database.GetConnection().Create(suite.user)
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

func (suite *JWTAuthenticatorTestSuite) TestGenerateToken() {
	token, err := GenerateToken("johndoe@example.org")
	suite.Nil(err)
	suite.Nil(suite.tokenAuthenticator.Authenticate(suite.createRequest(token), suite.user))
	suite.Equal("Admin", suite.user.Name)
}

func (suite *JWTAuthenticatorTestSuite) TestGenerateTokenValidates() {
	token, err := GenerateToken("johndoe@example.org")
	suite.Nil(err)
	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetString("auth.jwt.secret")), nil
	})
	suite.Nil(err)
}

func (suite *JWTAuthenticatorTestSuite) TestGenerateTokenHasClaims() {
	token, err := GenerateToken("johndoe@example.org")
	suite.Nil(err)
	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetString("auth.jwt.secret")), nil
	})
	suite.Nil(err)

	useridPresent := false
	for key, val := range claims {
		if key == "userid" {
			suite.Equal("johndoe@example.org", val)
			useridPresent = true
		}
	}

	suite.True(useridPresent)
}

func (suite *JWTAuthenticatorTestSuite) TestGenerateTokenWithClaimsHasClaims() {
	token, err := GenerateTokenWithClaims("johndoe@example.org", jwt.MapClaims{
		"sub": suite.user.ID,
	})
	suite.Nil(err)
	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetString("auth.jwt.secret")), nil
	})
	suite.Nil(err)

	useridPresent := false
	subPresent := false
	for key, val := range claims {
		if key == "userid" {
			suite.Equal("johndoe@example.org", val)
			useridPresent = true
		}
		if key == "sub" {
			suite.Equal(suite.user.ID, uint(val.(float64)))
			subPresent = true
		}
	}
	suite.True(useridPresent)
	suite.True(subPresent)
}

func (suite *JWTAuthenticatorTestSuite) TestGenerateTokenWithClaimsValidates() {
	token, err := GenerateToken("johndoe@example.org")
	suite.Nil(err)
	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetString("auth.jwt.secret")), nil
	})
	suite.Nil(err)
}

func (suite *JWTAuthenticatorTestSuite) TestGenerateTokenWithClaims() {
	token, err := GenerateTokenWithClaims("johndoe@example.org", jwt.MapClaims{
		"sub": suite.user.ID,
	})
	suite.Nil(err)
	suite.Nil(suite.tokenAuthenticator.Authenticate(suite.createRequest(token), suite.user))
	suite.Equal("Admin", suite.user.Name)
}

func (suite *JWTAuthenticatorTestSuite) TestGenerateTokenInvalidCredentials() {
	token, err := GenerateToken("wrongemail@example.org")
	suite.Nil(err)
	suite.Equal("These credentials don't match our records.", suite.tokenAuthenticator.Authenticate(suite.createRequest(token), suite.user).Error())
	suite.Nil(err)
}
func (suite *JWTAuthenticatorTestSuite) TestGenerateTokenWithClaimsInvalidCredentials() {
	token, err := GenerateTokenWithClaims("wrongemail@example.org", jwt.MapClaims{
		"sub": suite.user.ID,
	})
	suite.Nil(err)
	suite.Equal("These credentials don't match our records.", suite.tokenAuthenticator.Authenticate(suite.createRequest(token), suite.user).Error())
}

func (suite *JWTAuthenticatorTestSuite) TestAuthenticateInvalidToken() {
	request := suite.CreateTestRequest(nil)
	request.Header().Set("Authorization", "Basic userauthtoken")
	suite.Equal("Invalid or missing authentication header.", suite.tokenAuthenticator.Authenticate(request, suite.user).Error())

	userNoTable := &TestUserPromoted{}
	suite.Equal("Your authentication token is invalid.", suite.tokenAuthenticator.Authenticate(suite.createRequest("userauthtoken"), userNoTable).Error())

	suite.Panics(func() {
		userNoTable := &TestUserPromoted{}
		token, err := GenerateToken("wrongemail@example.org")
		suite.Nil(err)
		if err := suite.tokenAuthenticator.Authenticate(suite.createRequest(token), userNoTable); err != nil {
			suite.Fail(err.Error())
		}
	})
}

func (suite *JWTAuthenticatorTestSuite) TestAuthenticateTokenInFuture() {
	nbf := time.Now().Add(5 * time.Minute)
	token, err := suite.createWrongToken(jwt.SigningMethodHS256, "johndoe@example.org", nbf, nbf)
	suite.Nil(err)
	suite.Equal("Your authentication token is not valid yet.", suite.tokenAuthenticator.Authenticate(suite.createRequest(token), suite.user).Error())
}

func (suite *JWTAuthenticatorTestSuite) TestAuthenticateTokenExpired() {
	nbf := time.Now()
	exp := nbf.Add(-5 * time.Minute)
	token, err := suite.createWrongToken(jwt.SigningMethodHS256, "johndoe@example.org", nbf, exp)
	suite.Nil(err)
	suite.Equal("Your authentication token is expired.", suite.tokenAuthenticator.Authenticate(suite.createRequest(token), suite.user).Error())
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
