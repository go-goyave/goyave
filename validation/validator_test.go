package validation

import (
	"reflect"
	"testing"

	"github.com/System-Glitch/goyave/v2/lang"
	"github.com/stretchr/testify/suite"
)

type ValidatorTestSuite struct {
	suite.Suite
}

func (suite *ValidatorTestSuite) SetupSuite() {
	lang.LoadDefault()
}

func (suite *ValidatorTestSuite) TestIsTypeDependant() {
	// suite.True(isTypeDependent("min"))
	// suite.False(isTypeDependent("required"))
	// TODO test isTypeDependant
}

func (suite *ValidatorTestSuite) TestIsRequired() {
	// suite.True(isRequired([]string{"string", "required", "min:5"}))
	// suite.False(isRequired([]string{"string", "min:5"}))
	// TODO test isRequired
}

func (suite *ValidatorTestSuite) TestIsNullable() {
	// suite.True(isNullable([]string{"string", "required", "nullable", "min:5"}))
	// suite.False(isNullable([]string{"string", "min:5", "required"}))
	// TODO test isNullable
}

func (suite *ValidatorTestSuite) TestIsArray() {
	// suite.True(isArray([]string{"array", "required", "nullable", "min:5"}))
	// suite.False(isArray([]string{"string", "min:5", "required"}))
	// TODO test isArray
}

// func (suite *ValidatorTestSuite) TestArrayType() {
// 	suite.True(isArrayType("integer"))
// 	suite.False(isArrayType("file"))
// }
// TODO test IsType

func (suite *ValidatorTestSuite) TestParseRule() {
	rule := parseRule("required")
	suite.Equal("required", rule.Name)
	suite.Equal(0, len(rule.Params))

	rule = parseRule("min:5")
	suite.Equal("min", rule.Name)
	suite.Equal(1, len(rule.Params))
	suite.Equal("5", rule.Params[0])

	suite.Panics(func() {
		parseRule("invalid,rule")
	})

	suite.Panics(func() {
		parseRule("invalidrule")
	})

	rule = parseRule(">min:3")
	suite.Equal("min", rule.Name)
	suite.Equal(1, len(rule.Params))
	suite.Equal("3", rule.Params[0])
	suite.Equal(uint8(1), rule.ArrayDimension)

	suite.Panics(func() {
		parseRule(">file")
	})

	rule = parseRule(">>max:5")
	suite.Equal("max", rule.Name)
	suite.Equal(1, len(rule.Params))
	suite.Equal("5", rule.Params[0])
	suite.Equal(uint8(2), rule.ArrayDimension)
}

func (suite *ValidatorTestSuite) TestGetMessage() {
	suite.Equal("The :field is required.", getMessage(&Rule{Name: "required"}, reflect.ValueOf("test"), "en-US"))
	suite.Equal("The :field must be at least :min.", getMessage(&Rule{Name: "min"}, reflect.ValueOf(42), "en-US"))
	suite.Equal("The :field values must be at least :min.", getMessage(&Rule{Name: "min", ArrayDimension: 1}, reflect.ValueOf(42), "en-US"))
}

func (suite *ValidatorTestSuite) TestAddRule() {
	suite.Panics(func() {
		AddRule("required", &RuleDefinition{
			Function: func(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
				return false
			},
		})
	})

	AddRule("new_rule", &RuleDefinition{
		Function: func(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
			return true
		},
	})
	_, ok := validationRules["new_rule"]
	suite.True(ok)
	// No need to test that anymore
	// suite.False(isTypeDependent("new_rule"))

	AddRule("new_rule_type_dependent", &RuleDefinition{
		Function: func(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
			return true
		},
	})
	_, ok = validationRules["new_rule_type_dependent"]
	suite.True(ok)
	// No need to test that anymore
	// suite.True(isTypeDependent("new_rule_type_dependent"))
}

func (suite *ValidatorTestSuite) TestValidate() {
	errors := Validate(nil, Rules{}, false, "en-US")
	suite.Equal(1, len(errors))
	suite.Equal("Malformed request", errors["error"][0])

	errors = Validate(nil, Rules{}, true, "en-US")
	suite.Equal(1, len(errors))
	suite.Equal("Malformed JSON", errors["error"][0])

	errors = Validate(map[string]interface{}{
		"string": "hello world",
		"number": 42,
	}, ParseRuleSet(RuleSet{
		"string": {"required", "string"},
		"number": {"required", "numeric", "min:10"},
	}), true, "en-US")
	suite.Equal(0, len(errors))

	data := map[string]interface{}{
		"nullField": nil,
	}
	errors = Validate(data, ParseRuleSet(RuleSet{
		"nullField": {"numeric"},
	}), true, "en-US")
	_, exists := data["nullField"]
	suite.False(exists)
	suite.Equal(0, len(errors))

	data = map[string]interface{}{
		"nullField": nil,
	}
	errors = Validate(data, ParseRuleSet(RuleSet{
		"nullField": {"required", "nullable", "numeric"},
	}), true, "en-US")
	_, exists = data["nullField"]
	suite.True(exists)
	suite.Equal(0, len(errors))

	data = map[string]interface{}{
		"nullField": "test",
	}
	errors = Validate(data, ParseRuleSet(RuleSet{
		"nullField": {"required", "nullable", "numeric"},
	}), true, "en-US")
	_, exists = data["nullField"]
	suite.True(exists)
	suite.Equal(1, len(errors))
}

func (suite *ValidatorTestSuite) TestValidateWithArray() {
	data := map[string]interface{}{
		"string": "hello",
	}
	errors := Validate(data, ParseRuleSet(RuleSet{
		"string": {"required", "array"},
	}), false, "en-US")
	suite.Equal("array", GetFieldType(data["string"]))
	suite.Equal("hello", data["string"].([]string)[0])
	suite.Equal(0, len(errors))
}

func (suite *ValidatorTestSuite) TestValidateArrayValues() {
	data := map[string]interface{}{
		"string": []string{"hello", "world"},
	}
	errors := Validate(data, ParseRuleSet(RuleSet{
		"string": {"required", "array", ">min:3"},
	}), false, "en-US")
	suite.Equal(0, len(errors))

	data = map[string]interface{}{
		"string": []string{"hi", ",", "there"},
	}
	errors = Validate(data, ParseRuleSet(RuleSet{
		"string": {"required", "array", ">min:3"},
	}), false, "en-US")
	suite.Equal(1, len(errors))

	data = map[string]interface{}{
		"string": []string{"johndoe@example.org", "foobar@example.org"},
	}
	errors = Validate(data, ParseRuleSet(RuleSet{
		"string": {"required", "array:string", ">email"},
	}), true, "en-US")
	suite.Equal(0, len(errors))

	suite.Panics(func() {
		rule := &Rule{Name: "required", ArrayDimension: 1}
		validateRuleInArray(rule, "string", rule.ArrayDimension, map[string]interface{}{"string": "hi"})
	})

	// Empty array
	data = map[string]interface{}{
		"string": []string{},
	}
	errors = Validate(data, ParseRuleSet(RuleSet{
		"string": {"array", ">uuid:5"},
	}), true, "en-US")
	suite.Equal(0, len(errors))

	// TODO test Validate with verbose declaration
}

func (suite *ValidatorTestSuite) TestValidateTwoDimensionalArray() {
	data := map[string]interface{}{
		"values": [][]interface{}{{"0.5", 1.42}, {0.6, 7}},
	}
	errors := Validate(data, ParseRuleSet(RuleSet{
		"values": {"required", "array", ">array:numeric"},
	}), false, "en-US")
	suite.Equal(0, len(errors))

	arr, ok := data["values"].([][]float64)
	suite.True(ok)
	if ok {
		suite.Equal(2, len(arr))
		suite.Equal(0.5, arr[0][0])
		suite.Equal(1.42, arr[0][1])
		suite.Equal(0.6, arr[1][0])
		suite.Equal(7.0, arr[1][1])
	}

	data = map[string]interface{}{
		"values": [][]float64{{5, 8}, {0.6, 7}},
	}
	errors = Validate(data, ParseRuleSet(RuleSet{
		"values": {"required", "array", ">array:numeric", ">min:3"},
	}), true, "en-US")
	suite.Equal(1, len(errors))

	_, ok = data["values"].([][]float64)
	suite.True(ok)

	data = map[string]interface{}{
		"values": [][]float64{{5, 8, 6}, {0.6, 7, 9}},
	}
	errors = Validate(data, ParseRuleSet(RuleSet{
		"values": {"required", "array", ">array:numeric", ">min:3"},
	}), true, "en-US")
	suite.Equal(0, len(errors))

	data = map[string]interface{}{
		"values": [][]float64{{5, 8}, {3, 7}},
	}
	errors = Validate(data, ParseRuleSet(RuleSet{
		"values": {"required", "array", ">array:numeric", ">>min:3"},
	}), true, "en-US")
	suite.Equal(0, len(errors))

	data = map[string]interface{}{
		"values": [][]float64{{5, 8}, {0.6, 7}},
	}
	errors = Validate(data, ParseRuleSet(RuleSet{
		"values": {"required", "array", ">array:numeric", ">>min:3"},
	}), true, "en-US")
	suite.Equal(1, len(errors))

	// TODO test Validate with verbose declaration
}

func (suite *ValidatorTestSuite) TestValidateNDimensionalArray() {
	data := map[string]interface{}{
		"values": [][][]interface{}{
			{{"0.5", 1.42}, {0.6, 4, 3}},
			{{"0.6", "1.43"}, {}, {2}},
		},
	}
	errors := Validate(data, ParseRuleSet(RuleSet{
		"values": {"required", "array", ">array", ">>array:numeric", ">max:3", ">>>max:4"},
	}), true, "en-US")
	suite.Equal(0, len(errors))

	arr, ok := data["values"].([][][]float64)
	suite.True(ok)
	if ok {
		suite.Equal(2, len(arr))
		suite.Equal(2, len(arr[0]))
		suite.Equal(3, len(arr[1]))
		suite.Equal(0.5, arr[0][0][0])
		suite.Equal(1.42, arr[0][0][1])
		suite.Equal(2.0, arr[1][2][0])
	}

	data = map[string]interface{}{
		"values": [][][]interface{}{
			{{"0.5", 1.42}, {0.6, 4, 3}},
			{{"0.6", "1.43"}, {}, {2}, {4}},
		},
	}
	errors = Validate(data, ParseRuleSet(RuleSet{
		"values": {"required", "array", ">array", ">>array:numeric", ">max:3", ">>>max:4"},
	}), true, "en-US")
	suite.Equal(1, len(errors))

	data = map[string]interface{}{
		"values": [][][]interface{}{
			{{"0.5", 1.42}, {0.6, 9, 3}},
			{{"0.6", "1.43"}, {}, {2}},
		},
	}
	errors = Validate(data, ParseRuleSet(RuleSet{
		"values": {"required", "array", ">array", ">>array:numeric", ">max:3", ">>>max:4"},
	}), true, "en-US")
	suite.Equal(1, len(errors))

	// TODO test Validate with verbose declaration
}

func TestValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(ValidatorTestSuite))
}
