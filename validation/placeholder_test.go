package validation

import (
	"testing"

	"github.com/System-Glitch/goyave/v3/config"
	"github.com/System-Glitch/goyave/v3/lang"
	"github.com/stretchr/testify/suite"
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
	suite.Equal("fieldName", placeholders[":field"]("fieldName", "required", []string{}, "en-US"))
	suite.Equal("email address", placeholders[":field"]("email", "required", []string{}, "en-US"))
	suite.Equal("5", placeholders[":min"]("field", "min", []string{"5"}, "en-US"))
	suite.Equal("5", placeholders[":max"]("field", "max", []string{"5"}, "en-US"))
	suite.Equal("10", placeholders[":max"]("field", "between", []string{"5", "10"}, "en-US"))

	suite.Equal("email address", placeholders[":other"]("field", "greater_than", []string{"email"}, "en-US"))
	suite.Equal("otherField", placeholders[":other"]("field", "greater_than", []string{"otherField"}, "en-US"))

	suite.Equal("a, b, c", placeholders[":values"]("field", "in", []string{"a", "b", "c"}, "en-US"))
	suite.Equal("", placeholders[":version"]("field", "uuid", []string{}, "en-US"))
	suite.Equal("v5", placeholders[":version"]("field", "uuid", []string{"5"}, "en-US"))

	suite.Equal("email address", placeholders[":date"]("field", "date", []string{"email"}, "en-US"))
	suite.Equal("2019-11-02T17:00:00", placeholders[":date"]("field", "date", []string{"2019-11-02T17:00:00"}, "en-US"))
	suite.Equal("2019-11-03T17:00:00", placeholders[":max_date"]("field", "date", []string{"2019-11-02T17:00:00", "2019-11-03T17:00:00"}, "en-US"))
}

func (suite *PlaceholderTestSuite) TestProcessPlaceholders() {
	suite.Equal("The email address is required.", processPlaceholders("email", "required", []string{}, "The :field is required.", "en-US"))
	suite.Equal("The email address is required.", processPlaceholders("user.email", "required", []string{}, "The :field is required.", "en-US"))
	suite.Equal("The image must be a file with one of the following extensions: ppm.", processPlaceholders("image", "extension", []string{"ppm"}, "The :field must be a file with one of the following extensions: :values.", "en-US"))
	suite.Equal("The image must be a file with one of the following extensions: ppm, png.", processPlaceholders("image", "extension", []string{"ppm", "png"}, "The :field must be a file with one of the following extensions: :values.", "en-US"))
	suite.Equal("The image must have exactly 2 file(s).", processPlaceholders("image", "count", []string{"2"}, "The :field must have exactly :value file(s).", "en-US"))
}

func TestPlaceholderTestSuite(t *testing.T) {
	suite.Run(t, new(PlaceholderTestSuite))
}
