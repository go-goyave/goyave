package database

import (
	"os"
	"testing"

	"github.com/System-Glitch/goyave/v3/config"
	"github.com/stretchr/testify/suite"
)

func userGenerator() interface{} {
	return &User{
		Name:  "John Doe",
		Email: "johndoe@example.org",
	}
}

type NoTable struct {
	Name string
}

func noTableGenerator() interface{} {
	return &NoTable{
		Name: "John Doe",
	}
}

type FactoryTestSuite struct {
	suite.Suite
	previousEnv string
}

func (suite *FactoryTestSuite) SetupSuite() {
	suite.previousEnv = os.Getenv("GOYAVE_ENV")
	os.Setenv("GOYAVE_ENV", "test")
	if err := config.Load(); err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FactoryTestSuite) TestGenerate() {
	factory := NewFactory(userGenerator)
	records := factory.Generate(2).([]*User)
	suite.Equal(2, len(records))
	for _, user := range records {
		suite.Equal("John Doe", user.Name)
		suite.Equal("johndoe@example.org", user.Email)
	}

	override := &User{
		Name:  "name override",
		Email: "email override",
	}
	records = factory.Override(override).Generate(2).([]*User)
	suite.Equal(2, len(records))
	for _, user := range records {
		suite.Equal("name override", user.Name)
		suite.Equal("email override", user.Email)
	}

	override = &User{
		Name: "name override",
	}
	records = factory.Override(override).Generate(2).([]*User)
	suite.Equal(2, len(records))
	for _, user := range records {
		suite.Equal("name override", user.Name)
		suite.Equal("johndoe@example.org", user.Email)
	}

	suite.Panics(func() {
		override := &struct{ NotThere int }{
			NotThere: 2,
		}
		factory.Override(override).Generate(2)
	})
}

func (suite *FactoryTestSuite) TestSave() {

	db := GetConnection()
	db.AutoMigrate(&User{})
	defer db.Migrator().DropTable(&User{})

	records := NewFactory(userGenerator).Save(2).([]*User)
	suite.Equal(2, len(records))
	for i := uint(0); i < 2; i++ {
		suite.Equal(i+1, records[i].ID)
	}

	users := make([]*User, 0, 2)
	db.Find(&users)

	for _, user := range users {
		suite.Equal("John Doe", user.Name)
		suite.Equal("johndoe@example.org", user.Email)
	}

	suite.Panics(func() {
		NewFactory(noTableGenerator).Save(2)
	})
}

func (suite *FactoryTestSuite) TearDownAllSuite() {
	os.Setenv("GOYAVE_ENV", suite.previousEnv)
}

func TestFactoryTestSuite(t *testing.T) {
	// Ensure this test is running with a working database service running
	// in the background. Running "run_test.sh" runs a mariadb container.
	suite.Run(t, new(FactoryTestSuite))
}
