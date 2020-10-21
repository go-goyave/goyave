package database

import (
	"os"
	"testing"

	"github.com/System-Glitch/goyave/v3/config"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
)

type ValidationTestSuite struct {
	suite.Suite
	previousEnv string
}

func (suite *ValidationTestSuite) SetupSuite() {
	if _, ok := dialects["mysql"]; !ok {
		RegisterDialect("mysql", "{username}:{password}@({host}:{port})/{name}?{options}", mysql.Open)
	}
	suite.previousEnv = os.Getenv("GOYAVE_ENV")
	os.Setenv("GOYAVE_ENV", "test")
	if err := config.Load(); err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *ValidationTestSuite) TestValidateUnique() {
	ClearRegisteredModels()
	RegisterModel(&User{})
	Migrate()
	defer ClearRegisteredModels()

	user := &User{
		Name:  "Hugh",
		Email: "hugh@example.org",
	}
	db := Conn()
	db.Create(user)
	defer db.Migrator().DropTable(user)

	suite.False(validateUnique("email", "hugh@example.org", []string{"users"}, map[string]interface{}{}))
	suite.False(validateUnique("email", "hugh@example.org", []string{"users", "email"}, map[string]interface{}{}))
	suite.True(validateUnique("email", "hugh2@example.org", []string{"users"}, map[string]interface{}{}))
	suite.True(validateUnique("email", "hugh2@example.org", []string{"users", "email"}, map[string]interface{}{}))
	suite.True(validateUnique("email", "hugh@example.org", []string{"users", "name"}, map[string]interface{}{}))

	// model not found
	suite.Panics(func() {
		validateUnique("email", "hugh@example.org", []string{"not a model", "email"}, map[string]interface{}{})
	})
}

func (suite *ValidationTestSuite) TearDownAllSuite() {
	os.Setenv("GOYAVE_ENV", suite.previousEnv)
}

func TestValidationTestSuite(t *testing.T) {
	// Ensure this test is running with a working database service running
	// in the background. Running "run_test.sh" runs a mariadb container.
	suite.Run(t, new(ValidationTestSuite))
}
