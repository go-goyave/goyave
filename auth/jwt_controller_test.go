package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"goyave.dev/goyave/v3"
	"goyave.dev/goyave/v3/config"
	"goyave.dev/goyave/v3/database"
	"goyave.dev/goyave/v3/validation"
)

const testUserPassword = "secret"

type JWTControllerTestSuite struct {
	user *TestUser
	goyave.TestSuite
}

func (suite *JWTControllerTestSuite) SetupSuite() {
	config.Set("database.connection", "mysql")
	database.ClearRegisteredModels()
	database.RegisterModel(&TestUser{})

	database.Migrate()
}

func (suite *JWTControllerTestSuite) SetupTest() {
	password, err := bcrypt.GenerateFromPassword([]byte(testUserPassword), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	suite.user = &TestUser{
		Name:     "Admin",
		Email:    "johndoe@example.org",
		Password: string(password),
	}
	database.GetConnection().Create(suite.user)
}

func (suite *JWTControllerTestSuite) TestLogin() {
	controller := NewJWTController(&TestUser{})
	suite.NotNil(controller)

	request := suite.CreateTestRequest(nil)
	request.Data = map[string]interface{}{
		"username": "johndoe@example.org",
		"password": testUserPassword,
	}
	writer := httptest.NewRecorder()
	response := suite.CreateTestResponse(writer)

	controller.Login(response, request)
	result := writer.Result()
	suite.Equal(http.StatusOK, result.StatusCode)

	json := map[string]string{}
	err := suite.GetJSONBody(result, &json)
	suite.Nil(err)

	if err == nil {
		token, ok := json["token"]
		suite.True(ok)
		suite.NotEmpty(token)
	}
	result.Body.Close()

	request.Data = map[string]interface{}{
		"username": "johndoe@example.org",
		"password": "wrongpassword",
	}
	writer = httptest.NewRecorder()
	response = suite.CreateTestResponse(writer)

	controller.Login(response, request)
	result = writer.Result()
	suite.Equal(http.StatusUnauthorized, result.StatusCode)

	json = map[string]string{}
	err = suite.GetJSONBody(result, &json)
	suite.Nil(err)

	if err == nil {
		errMessage, ok := json["validationError"]
		suite.True(ok)
		suite.Equal("These credentials don't match our records.", errMessage)
	}
	result.Body.Close()
}

func (suite *JWTControllerTestSuite) TestLoginWithCustomTokenFunc() {
	controller := NewJWTController(&TestUser{})
	suite.NotNil(controller)
	controller.TokenFunc = func(r *goyave.Request, user interface{}) (string, error) {
		return GenerateTokenWithClaims(jwt.MapClaims{
			"userid": (user.(*TestUser)).Email,
			"sub":    (user.(*TestUser)).ID,
		}, jwt.SigningMethodHS256)
	}

	request := suite.CreateTestRequest(nil)
	request.Data = map[string]interface{}{
		"username": "johndoe@example.org",
		"password": testUserPassword,
	}
	writer := httptest.NewRecorder()
	response := suite.CreateTestResponse(writer)

	controller.Login(response, request)
	result := writer.Result()
	suite.Equal(http.StatusOK, result.StatusCode)

	json := map[string]string{}
	err := suite.GetJSONBody(result, &json)
	suite.Nil(err)

	if err == nil {
		token, ok := json["token"]
		suite.True(ok)
		suite.NotEmpty(token)
		claims := jwt.MapClaims{}
		_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.GetString("auth.jwt.secret")), nil
		})
		suite.Nil(err)

		userID, okID := claims["userid"]
		suite.True(okID)
		suite.Equal("johndoe@example.org", userID)
		sub, okSub := claims["sub"]
		suite.True(okSub)
		suite.Equal(suite.user.ID, uint(sub.(float64)))
	}
	result.Body.Close()

	request.Data = map[string]interface{}{
		"username": "johndoe@example.org",
		"password": "wrongpassword",
	}
	writer = httptest.NewRecorder()
	response = suite.CreateTestResponse(writer)

	controller.Login(response, request)
	result = writer.Result()
	suite.Equal(http.StatusUnauthorized, result.StatusCode)

	json = map[string]string{}
	err = suite.GetJSONBody(result, &json)
	suite.Nil(err)

	if err == nil {
		errMessage, ok := json["validationError"]
		suite.True(ok)
		suite.Equal("These credentials don't match our records.", errMessage)
	}
	result.Body.Close()
}

func (suite *JWTControllerTestSuite) TestLoginTokenFuncError() {
	controller := NewJWTController(&TestUser{})
	suite.NotNil(controller)
	controller.TokenFunc = func(r *goyave.Request, user interface{}) (string, error) {
		return "", fmt.Errorf("test error")
	}
	request := suite.CreateTestRequest(nil)
	request.Data = map[string]interface{}{
		"username": "johndoe@example.org",
		"password": testUserPassword,
	}
	writer := httptest.NewRecorder()
	response := suite.CreateTestResponse(writer)
	suite.Panics(func() {
		controller.Login(response, request)
	})
}

func (suite *JWTControllerTestSuite) TestLoginWithFieldOverride() {
	controller := NewJWTController(&TestUser{})
	controller.UsernameField = "email"
	controller.PasswordField = "pass"
	suite.NotNil(controller)

	request := suite.CreateTestRequest(nil)
	request.Data = map[string]interface{}{
		"email": "johndoe@example.org",
		"pass":  testUserPassword,
	}
	writer := httptest.NewRecorder()
	response := suite.CreateTestResponse(writer)

	controller.Login(response, request)
	result := writer.Result()
	suite.Equal(http.StatusOK, result.StatusCode)
	result.Body.Close()
}

func (suite *JWTControllerTestSuite) TestValidation() {
	suite.RunServer(func(router *goyave.Router) {
		JWTRoutes(router, &TestUser{})
	}, func() {
		headers := map[string]string{"Content-Type": "application/json"}
		data := map[string]interface{}{}
		body, _ := json.Marshal(data)
		resp, err := suite.Post("/auth/login", headers, bytes.NewReader(body))
		suite.Nil(err)
		if err == nil {
			defer resp.Body.Close()
			json := map[string]validation.Errors{}
			err := suite.GetJSONBody(resp, &json)
			suite.Nil(err)
			if err == nil {
				suite.Len(json["validationError"]["username"], 2)
				suite.Len(json["validationError"]["password"], 2)
			}
		}
	})
}

func (suite *JWTControllerTestSuite) TestLoginPanic() {
	suite.Panics(func() {
		request := suite.CreateTestRequest(nil)
		request.Data = map[string]interface{}{
			"username": "johndoe@example.org",
			"password": testUserPassword,
		}
		controller := NewJWTController(&TestUserOverride{})
		controller.Login(suite.CreateTestResponse(httptest.NewRecorder()), request)
	})
}

func (suite *JWTControllerTestSuite) TestRoutes() {
	suite.RunServer(func(router *goyave.Router) {
		suite.NotNil(JWTRoutes(router, &TestUser{}))
	}, func() {
		json, err := json.Marshal(map[string]string{
			"username": "johndoe@example.org",
			"password": testUserPassword,
		})
		if err != nil {
			panic(err)
		}

		headers := map[string]string{"Content-Type": "application/json"}
		resp, err := suite.Post("/auth/login", headers, strings.NewReader(string(json)))
		suite.Nil(err)
		suite.NotNil(resp)
		if resp != nil {
			suite.Equal(200, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

func (suite *JWTControllerTestSuite) TearDownTest() {
	suite.ClearDatabase()
}

func (suite *JWTControllerTestSuite) TearDownSuite() {
	database.Conn().Migrator().DropTable(&TestUser{})
	database.ClearRegisteredModels()
}

func TestJWTControllerSuite(t *testing.T) {
	goyave.RunTest(t, new(JWTControllerTestSuite))
}
