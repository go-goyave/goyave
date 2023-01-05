package validation

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/lang"
)

type PlaceholderTestSuite struct {
	suite.Suite
}

func (suite *PlaceholderTestSuite) SetupSuite() {
	lang.LoadDefault()
	if err := config.LoadFrom("../config.test.json"); err != nil {
		suite.FailNow(err.Error())
	}
	config.Set("app.defaultLanguage", "en-US")
}

func (suite *PlaceholderTestSuite) TestPlaceholders() {
	suite.Equal("fieldName", placeholders[":field"]("fieldName", "en-US", &Context{Rule: &Rule{Name: "required", Params: []string{}}}))
	suite.Equal("email address", placeholders[":field"]("email", "en-US", &Context{Rule: &Rule{Name: "required", Params: []string{}}}))
	suite.Equal("5", placeholders[":min"]("field", "en-US", &Context{Rule: &Rule{Name: "min", Params: []string{"5"}}}))
	suite.Equal("5", placeholders[":max"]("field", "en-US", &Context{Rule: &Rule{Name: "max", Params: []string{"5"}}}))
	suite.Equal("10", placeholders[":max"]("field", "en-US", &Context{Rule: &Rule{Name: "between", Params: []string{"5", "10"}}}))

	suite.Equal("email address", placeholders[":other"]("field", "en-US", &Context{Rule: &Rule{Name: "greater_than", Params: []string{"email"}}}))
	suite.Equal("otherField", placeholders[":other"]("field", "en-US", &Context{Rule: &Rule{Name: "greater_than", Params: []string{"otherField"}}}))

	suite.Equal("a, b, c", placeholders[":values"]("field", "en-US", &Context{Rule: &Rule{Name: "in", Params: []string{"a", "b", "c"}}}))
	suite.Equal("", placeholders[":version"]("field", "en-US", &Context{Rule: &Rule{Name: "uuid", Params: []string{}}}))
	suite.Equal("v5", placeholders[":version"]("field", "en-US", &Context{Rule: &Rule{Name: "uuid", Params: []string{"5"}}}))

	suite.Equal("email address", placeholders[":date"]("field", "en-US", &Context{Rule: &Rule{Name: "date", Params: []string{"email"}}}))
	suite.Equal("2019-11-02T17:00:00", placeholders[":date"]("field", "en-US", &Context{Rule: &Rule{Name: "date", Params: []string{"2019-11-02T17:00:00"}}}))
	suite.Equal("2019-11-03T17:00:00", placeholders[":max_date"]("field", "en-US", &Context{Rule: &Rule{Name: "date", Params: []string{"2019-11-02T17:00:00", "2019-11-03T17:00:00"}}}))
}

func (suite *PlaceholderTestSuite) TestProcessPlaceholders() {
	suite.Equal("The email address is required.", processPlaceholders("email", "The :field is required.", "en-US", &Context{Rule: &Rule{Name: "required", Params: []string{}}}))
	suite.Equal("The email address is required.", processPlaceholders("user.email", "The :field is required.", "en-US", &Context{Rule: &Rule{Name: "required", Params: []string{}}}))
	suite.Equal("The image must be a file with one of the following extensions: ppm.", processPlaceholders("image", "The :field must be a file with one of the following extensions: :values.", "en-US", &Context{Rule: &Rule{Name: "extension", Params: []string{"ppm"}}}))
	suite.Equal("The image must be a file with one of the following extensions: ppm, png.", processPlaceholders("image", "The :field must be a file with one of the following extensions: :values.", "en-US", &Context{Rule: &Rule{Name: "extension", Params: []string{"ppm", "png"}}}))
	suite.Equal("The image must have exactly 2 file(s).", processPlaceholders("image", "The :field must have exactly :value file(s).", "en-US", &Context{Rule: &Rule{Name: "count", Params: []string{"2"}}}))
}

func (suite *PlaceholderTestSuite) TestPlaceholdersInjectionPrevention() {
	suite.Equal("The :date is required.", processPlaceholders(":date", "The :field is required.", "en-US", &Context{Rule: &Rule{Name: "required", Params: []string{}}}))
}

func TestPlaceholderTestSuite(t *testing.T) {
	suite.Run(t, new(PlaceholderTestSuite))
}
