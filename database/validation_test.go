package database

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/validation"
)

type ValidationTestSuite struct {
	suite.Suite
	previousEnv string
}

func newTestContext(field string, value interface{}, parameters []string, form map[string]interface{}) *validation.Context {
	ctx := validation.NewContext()
	ctx.Data = form
	ctx.Value = value
	ctx.Name = field
	ctx.Rule = &validation.Rule{
		Params: parameters,
	}
	return ctx
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

	suite.False(validateUnique(newTestContext("email", "hugh@example.org", []string{"users"}, map[string]interface{}{})))
	suite.False(validateUnique(newTestContext("email", "hugh@example.org", []string{"users", "email"}, map[string]interface{}{})))
	suite.True(validateUnique(newTestContext("email", "hugh2@example.org", []string{"users"}, map[string]interface{}{})))
	suite.True(validateUnique(newTestContext("email", "hugh2@example.org", []string{"users", "email"}, map[string]interface{}{})))
	suite.True(validateUnique(newTestContext("email", "hugh@example.org", []string{"users", "name"}, map[string]interface{}{})))

	// model not found
	suite.Panics(func() {
		validateUnique(newTestContext("email", "hugh@example.org", []string{"not a model", "email"}, map[string]interface{}{}))
	})
}

func (suite *ValidationTestSuite) TestValidateExists() {
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

	suite.True(validateExists(newTestContext("email", "hugh@example.org", []string{"users"}, map[string]interface{}{})))
	suite.True(validateExists(newTestContext("email", "hugh@example.org", []string{"users", "email"}, map[string]interface{}{})))
	suite.False(validateExists(newTestContext("email", "hugh2@example.org", []string{"users"}, map[string]interface{}{})))
	suite.False(validateExists(newTestContext("email", "hugh2@example.org", []string{"users", "email"}, map[string]interface{}{})))
	suite.False(validateExists(newTestContext("email", "hugh@example.org", []string{"users", "name"}, map[string]interface{}{})))

	// model not found
	suite.Panics(func() {
		validateExists(newTestContext("email", "hugh@example.org", []string{"not a model", "email"}, map[string]interface{}{}))
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
