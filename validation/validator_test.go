package validation

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
	"goyave.dev/goyave/v3/helper/filesystem"
	"goyave.dev/goyave/v3/lang"
)

type ValidatorTestSuite struct {
	suite.Suite
}

func (suite *ValidatorTestSuite) SetupSuite() {
	lang.LoadDefault()
}

func (suite *ValidatorTestSuite) TestParseRule() {
	rule := parseRule("required")
	suite.Equal("required", rule.Name)
	suite.Equal(0, len(rule.Params))
	suite.Equal(uint8(0), rule.ArrayDimension)

	rule = parseRule("min:5")
	suite.Equal("min", rule.Name)
	suite.Equal(1, len(rule.Params))
	suite.Equal("5", rule.Params[0])
	suite.Equal(uint8(0), rule.ArrayDimension)

	suite.Panics(func() {
		parseRule("invalid,rule")
	})

	rule = parseRule(">min:3")
	suite.Equal("min", rule.Name)
	suite.Equal(1, len(rule.Params))
	suite.Equal("3", rule.Params[0])
	suite.Equal(uint8(1), rule.ArrayDimension)

	rule = parseRule(">>max:5")
	suite.Equal("max", rule.Name)
	suite.Equal(1, len(rule.Params))
	suite.Equal("5", rule.Params[0])
	suite.Equal(uint8(2), rule.ArrayDimension)
}

func (suite *ValidatorTestSuite) TestGetMessage() {
	suite.Equal("The :field is required.", getMessage([]*Rule{}, &Rule{Name: "required"}, reflect.ValueOf("test"), "en-US"))
	suite.Equal("The :field must be at least :min.", getMessage([]*Rule{{Name: "numeric"}}, &Rule{Name: "min"}, reflect.ValueOf(42), "en-US"))
	suite.Equal("The :field values must be at least :min.", getMessage([]*Rule{{Name: "numeric", ArrayDimension: 1}}, &Rule{Name: "min", ArrayDimension: 1}, reflect.ValueOf(42), "en-US"))

	rules := []*Rule{
		{Name: "array", Params: []string{"numeric"}},
		{Name: "min", ArrayDimension: 1},
	}
	suite.Equal("The :field values must be at least :min.", getMessage(rules, rules[1], reflect.ValueOf(42), "en-US"))

	// Test type fallback if no type rule is found
	suite.Equal("The :field must be at least :min.", getMessage([]*Rule{}, &Rule{Name: "min"}, reflect.ValueOf(42), "en-US"))
	suite.Equal("The :field must be at least :min characters.", getMessage([]*Rule{}, &Rule{Name: "min"}, reflect.ValueOf("test"), "en-US"))

	// Integer share message with numeric
	suite.Equal("The :field must be at least :min.", getMessage([]*Rule{{"integer", nil, 0}}, &Rule{Name: "min"}, reflect.Value{}, "en-US"))
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
}

func (suite *ValidatorTestSuite) TestValidate() {
	errors := Validate(nil, &Rules{}, false, "en-US")
	suite.Equal(1, len(errors))
	suite.Equal("Malformed request", errors["error"][0])

	errors = Validate(nil, RuleSet{}, false, "en-US")
	suite.Equal(1, len(errors))
	suite.Equal("Malformed request", errors["error"][0])

	errors = Validate(nil, &Rules{}, true, "en-US")
	suite.Equal(1, len(errors))
	suite.Equal("Malformed JSON", errors["error"][0])

	errors = Validate(nil, RuleSet{}, true, "en-US")
	suite.Equal(1, len(errors))
	suite.Equal("Malformed JSON", errors["error"][0])

	errors = Validate(map[string]interface{}{
		"string": "hello world",
		"number": 42,
	}, RuleSet{
		"string": {"required", "string"},
		"number": {"required", "numeric", "min:10"},
	}, true, "en-US")
	suite.Equal(0, len(errors))

	errors = Validate(map[string]interface{}{
		"string": "hello world",
		"number": 42,
	}, &Rules{
		Fields: FieldMap{
			"string": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "string"},
				},
			},
			"number": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "numeric"},
					{Name: "min", Params: []string{"10"}},
				},
			},
		},
	}, true, "en-US")
	suite.Equal(0, len(errors))

	data := map[string]interface{}{
		"nullField": nil,
	}
	errors = Validate(data, RuleSet{
		"nullField": {"numeric"},
	}, true, "en-US")
	_, exists := data["nullField"]
	suite.False(exists)
	suite.Equal(0, len(errors))

	data = map[string]interface{}{
		"nullField": nil,
	}
	errors = Validate(data, RuleSet{
		"nullField": {"required", "nullable", "numeric"},
	}, true, "en-US")
	val, exists := data["nullField"]
	suite.True(exists)
	suite.Nil(val)
	suite.Equal(0, len(errors))

	data = map[string]interface{}{
		"nullField": "test",
	}
	errors = Validate(data, RuleSet{
		"nullField": {"required", "nullable", "numeric"},
	}, true, "en-US")
	val, exists = data["nullField"]
	suite.True(exists)
	suite.Equal("test", val)
	suite.Equal(1, len(errors))

	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"nullField": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "nullable"},
					{Name: "numeric"},
				},
			},
		},
	}, true, "en-US")
	val, exists = data["nullField"]
	suite.True(exists)
	suite.Equal("test", val)
	suite.Equal(1, len(errors))

	data = map[string]interface{}{}
	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"text": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "string"},
				},
			},
		},
	}, true, "en-US")
	suite.Equal(1, len(errors))
	suite.Equal("The text is required.", errors["text"][0])
}

func (suite *ValidatorTestSuite) TestValidateWithArray() {
	data := map[string]interface{}{
		"string": "hello",
	}
	errors := Validate(data, RuleSet{
		"string": {"required", "array"},
	}, false, "en-US")
	suite.Equal("array", GetFieldType(data["string"]))
	suite.Equal("hello", data["string"].([]string)[0])
	suite.Equal(0, len(errors))

	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"string": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
				},
			},
		},
	}, false, "en-US")
	suite.Equal("array", GetFieldType(data["string"]))
	suite.Equal("hello", data["string"].([]string)[0])
	suite.Equal(0, len(errors))
}

func (suite *ValidatorTestSuite) TestValidateArrayValues() {
	data := map[string]interface{}{
		"string": []string{"hello", "world"},
	}
	errors := Validate(data, RuleSet{
		"string": {"required", "array", ">min:3"},
	}, false, "en-US")
	suite.Len(errors, 0)

	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"string": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
					{Name: "min", Params: []string{"3"}, ArrayDimension: 1},
				},
			},
		},
	}, false, "en-US")
	suite.Len(errors, 0)

	data = map[string]interface{}{
		"string": []string{"hi", ",", "there"},
	}
	errors = Validate(data, RuleSet{
		"string": {"required", "array", ">min:3"},
	}, false, "en-US")
	suite.Len(errors, 1)

	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"string": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
					{Name: "min", Params: []string{"3"}, ArrayDimension: 1},
				},
			},
		},
	}, false, "en-US")
	suite.Len(errors, 1)

	data = map[string]interface{}{
		"string": []string{"johndoe@example.org", "foobar@example.org"},
	}
	errors = Validate(data, RuleSet{
		"string": {"required", "array:string", ">email"},
	}, true, "en-US")
	suite.Len(errors, 0)

	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"string": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array", Params: []string{"string"}},
					{Name: "email", ArrayDimension: 1},
				},
			},
		},
	}, true, "en-US")
	suite.Len(errors, 0)

	// Cannot validate array values on non-array field string of type string
	rule := &Rule{Name: "required", ArrayDimension: 1}
	suite.False(validateRuleInArray(rule, "string", rule.ArrayDimension, map[string]interface{}{"string": "hi"}))

	// Empty array
	data = map[string]interface{}{
		"string": []string{},
	}
	errors = Validate(data, RuleSet{
		"string": {"array", ">uuid:5"},
	}, true, "en-US")
	suite.Len(errors, 0)

	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"string": {
				Rules: []*Rule{
					{Name: "array"},
					{Name: "uuid", Params: []string{"5"}, ArrayDimension: 1},
				},
			},
		},
	}, true, "en-US")
	suite.Len(errors, 0)
}

func (suite *ValidatorTestSuite) TestValidateTwoDimensionalArray() {
	data := map[string]interface{}{
		"values": [][]interface{}{{"0.5", 1.42}, {0.6, 7}},
	}
	errors := Validate(data, RuleSet{
		"values": {"required", "array", ">array:numeric"},
	}, false, "en-US")
	suite.Len(errors, 0)

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
		"values": [][]interface{}{{"0.5", 1.42}, {0.6, 7}},
	}
	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"values": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
					{Name: "array", Params: []string{"numeric"}, ArrayDimension: 1},
				},
			},
		},
	}, false, "en-US")
	suite.Len(errors, 0)

	arr, ok = data["values"].([][]float64)
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
	errors = Validate(data, RuleSet{
		"values": {"required", "array", ">array:numeric", ">min:3"},
	}, true, "en-US")
	suite.Len(errors, 1)

	_, ok = data["values"].([][]float64)
	suite.True(ok)

	data = map[string]interface{}{
		"values": [][]float64{{5, 8}, {0.6, 7}},
	}
	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"values": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
					{Name: "array", Params: []string{"numeric"}, ArrayDimension: 1},
					{Name: "min", Params: []string{"3"}, ArrayDimension: 1},
				},
			},
		},
	}, true, "en-US")
	suite.Len(errors, 1)

	_, ok = data["values"].([][]float64)
	suite.True(ok)

	data = map[string]interface{}{
		"values": [][]float64{{5, 8, 6}, {0.6, 7, 9}},
	}
	errors = Validate(data, RuleSet{
		"values": {"required", "array", ">array:numeric", ">min:3"},
	}, true, "en-US")
	suite.Len(errors, 0)

	data = map[string]interface{}{
		"values": [][]float64{{5, 8, 6}, {0.6, 7, 9}},
	}
	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"values": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
					{Name: "array", Params: []string{"numeric"}, ArrayDimension: 1},
					{Name: "min", Params: []string{"3"}, ArrayDimension: 1},
				},
			},
		},
	}, true, "en-US")
	suite.Len(errors, 0)

	data = map[string]interface{}{
		"values": [][]float64{{5, 8}, {3, 7}},
	}
	errors = Validate(data, RuleSet{
		"values": {"required", "array", ">array:numeric", ">>min:3"},
	}, true, "en-US")
	suite.Len(errors, 0)

	data = map[string]interface{}{
		"values": [][]float64{{5, 8}, {3, 7}},
	}
	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"values": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
					{Name: "array", Params: []string{"numeric"}, ArrayDimension: 1},
					{Name: "min", Params: []string{"3"}, ArrayDimension: 2},
				},
			},
		},
	}, true, "en-US")
	suite.Len(errors, 0)

	data = map[string]interface{}{
		"values": [][]float64{{5, 8}, {0.6, 7}},
	}
	errors = Validate(data, RuleSet{
		"values": {"required", "array", ">array:numeric", ">>min:3"},
	}, true, "en-US")
	suite.Len(errors, 1)

	data = map[string]interface{}{
		"values": [][]float64{{5, 8}, {0.6, 7}},
	}
	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"values": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
					{Name: "array", Params: []string{"numeric"}, ArrayDimension: 1},
					{Name: "min", Params: []string{"3"}, ArrayDimension: 2},
				},
			},
		},
	}, true, "en-US")
	suite.Len(errors, 1)
}

func (suite *ValidatorTestSuite) TestValidateNDimensionalArray() {
	data := map[string]interface{}{
		"values": [][][]interface{}{
			{{"0.5", 1.42}, {0.6, 4, 3}},
			{{"0.6", "1.43"}, {}, {2}},
		},
	}
	errors := Validate(data, RuleSet{
		"values": {"required", "array", ">array", ">>array:numeric", ">max:3", ">>>max:4"},
	}, true, "en-US")
	suite.Len(errors, 0)

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
			{{"0.6", "1.43"}, {}, {2}},
		},
	}
	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"values": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
					{Name: "array", ArrayDimension: 1},
					{Name: "array", Params: []string{"numeric"}, ArrayDimension: 2},
					{Name: "max", Params: []string{"3"}, ArrayDimension: 1},
					{Name: "max", Params: []string{"4"}, ArrayDimension: 3},
				},
			},
		},
	}, true, "en-US")
	suite.Len(errors, 0)

	arr, ok = data["values"].([][][]float64)
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
	errors = Validate(data, RuleSet{
		"values": {"required", "array", ">array", ">>array:numeric", ">max:3", ">>>max:4"},
	}, true, "en-US")
	suite.Len(errors, 1)

	data = map[string]interface{}{
		"values": [][][]interface{}{
			{{"0.5", 1.42}, {0.6, 4, 3}},
			{{"0.6", "1.43"}, {}, {2}, {4}},
		},
	}
	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"values": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
					{Name: "array", ArrayDimension: 1},
					{Name: "array", Params: []string{"numeric"}, ArrayDimension: 2},
					{Name: "max", Params: []string{"3"}, ArrayDimension: 1},
					{Name: "max", Params: []string{"4"}, ArrayDimension: 3},
				},
			},
		},
	}, true, "en-US")
	suite.Len(errors, 1)

	data = map[string]interface{}{
		"values": [][][]interface{}{
			{{"0.5", 1.42}, {0.6, 9, 3}},
			{{"0.6", "1.43"}, {}, {2}},
		},
	}
	errors = Validate(data, RuleSet{
		"values": {"required", "array", ">array", ">>array:numeric", ">max:3", ">>>max:4"},
	}, true, "en-US")
	suite.Len(errors, 1)

	data = map[string]interface{}{
		"values": [][][]interface{}{
			{{"0.5", 1.42}, {0.6, 9, 3}},
			{{"0.6", "1.43"}, {}, {2}},
		},
	}
	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"values": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
					{Name: "array", ArrayDimension: 1},
					{Name: "array", Params: []string{"numeric"}, ArrayDimension: 2},
					{Name: "max", Params: []string{"3"}, ArrayDimension: 1},
					{Name: "max", Params: []string{"4"}, ArrayDimension: 3},
				},
			},
		},
	}, true, "en-US")
	suite.Len(errors, 1)
}

func (suite *ValidatorTestSuite) TestFieldCheck() {
	suite.NotPanics(func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "required"},
				{Name: "numeric"},
			},
		}

		field.check()

		suite.True(field.isRequired)
		suite.False(field.isArray)
		suite.False(field.isNullable)
	})

	suite.NotPanics(func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "nullable"},
				{Name: "array"},
			},
		}

		field.check()

		suite.False(field.isRequired)
		suite.True(field.isArray)
		suite.True(field.isNullable)
	})

	suite.Panics(func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "required"},
				{Name: "not a rule"},
			},
		}

		field.check()
	})
}

func (suite *ValidatorTestSuite) TestFieldCheckArrayProhibitedRules() {
	prohibitedRules := []string{
		"confirmed", "file", "mime", "image", "extension", "count",
		"count_min", "count_max", "count_between",
	}
	for _, v := range prohibitedRules {
		suite.Panics(func() {
			field := &Field{
				Rules: []*Rule{
					{Name: v, ArrayDimension: 1},
				},
			}
			field.check()
		})
	}
}

func (suite *ValidatorTestSuite) TestParseRuleSet() {
	set := RuleSet{
		"string": {"required", "array:string", ">min:3"},
		"number": {"numeric"},
	}

	rules := set.parse()
	suite.Len(rules.Fields, 2)
	suite.Len(rules.Fields["string"].Rules, 3)
	suite.Equal(&Rule{Name: "required", Params: []string{}, ArrayDimension: 0}, rules.Fields["string"].Rules[0])
	suite.Equal(&Rule{Name: "array", Params: []string{"string"}, ArrayDimension: 0}, rules.Fields["string"].Rules[1])
	suite.Equal(&Rule{Name: "min", Params: []string{"3"}, ArrayDimension: 1}, rules.Fields["string"].Rules[2])
	suite.Len(rules.Fields["number"].Rules, 1)
	suite.Equal(&Rule{Name: "numeric", Params: []string{}, ArrayDimension: 0}, rules.Fields["number"].Rules[0])

	suite.Equal(rules, set.AsRules())

	suite.True(rules.checked)
	// Resulting Rules should be checked after parsing
	suite.Panics(func() {
		set := RuleSet{
			"string": {"required", "not a rule", ">min:3"},
		}
		set.parse()
	})
}

func (suite *ValidatorTestSuite) TestAsRules() {
	rules := &Rules{
		Fields: FieldMap{
			"values": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
					{Name: "array", ArrayDimension: 1},
					{Name: "array", Params: []string{"numeric"}, ArrayDimension: 2},
					{Name: "max", Params: []string{"3"}, ArrayDimension: 1},
					{Name: "max", Params: []string{"4"}, ArrayDimension: 3},
				},
			},
		},
	}
	suite.Equal(rules, rules.AsRules())

	suite.Panics(func() {
		rules := &Rules{
			Fields: FieldMap{
				"values": {
					Rules: []*Rule{
						{Name: "not a rule"},
					},
				},
			},
		}
		suite.False(rules.checked)
		rules.AsRules()
	})
}

func (suite *ValidatorTestSuite) TestRulesCheck() {
	rules := &Rules{
		Fields: FieldMap{
			"values": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
					{Name: "array", ArrayDimension: 1},
					{Name: "array", Params: []string{"numeric"}, ArrayDimension: 2},
					{Name: "max", Params: []string{"3"}, ArrayDimension: 1},
					{Name: "max", Params: []string{"4"}, ArrayDimension: 3},
				},
			},
		},
	}
	suite.False(rules.checked)
	rules.check()
	suite.True(rules.checked)

	// Check should not be executed multiple times
	rules.Fields["values"].Rules[0].Name = "not a rule"
	suite.NotPanics(func() {
		rules.check()
	})
}

func (suite *ValidatorTestSuite) TestGetFieldType() {
	suite.Equal("numeric", getFieldType(reflect.ValueOf(1)))
	suite.Equal("numeric", getFieldType(reflect.ValueOf(1.1)))
	suite.Equal("numeric", getFieldType(reflect.ValueOf(uint(1))))
	suite.Equal("numeric", getFieldType(reflect.ValueOf(float32(1))))
	suite.Equal("string", getFieldType(reflect.ValueOf("hello")))
	suite.Equal("array", getFieldType(reflect.ValueOf([]string{"hello", "world"})))
	suite.Equal("file", getFieldType(reflect.ValueOf([]filesystem.File{})))
	suite.Equal("object", getFieldType(reflect.ValueOf(map[string]interface{}{"hello": 1, "world": "!"})))
	suite.Equal("unsupported", getFieldType(reflect.ValueOf(nil)))
	suite.Equal("unsupported", getFieldType(reflect.ValueOf(map[string]int{"hello": 1, "world": 2})))
}

func (suite *ValidatorTestSuite) TestGetFieldFromName() {
	data := map[string]interface{}{
		"notobject": "test",
		"object": map[string]interface{}{
			"key": 1,
			"child": map[string]interface{}{
				"name": "Michel",
			},
		},
	}

	name, val, parent, ok := GetFieldFromName("notobject", data)
	suite.Equal("notobject", name)
	suite.Equal("test", val)
	suite.Equal(data, parent)
	suite.True(ok)

	name, val, parent, ok = GetFieldFromName("object", data)
	suite.Equal("object", name)
	suite.Equal(data["object"], val)
	suite.Equal(data, parent)
	suite.True(ok)

	name, val, parent, ok = GetFieldFromName("object.key", data)
	suite.Equal("key", name)
	suite.Equal(1, val)
	suite.Equal(data["object"], parent)
	suite.True(ok)

	name, val, parent, ok = GetFieldFromName("object.child", data)
	suite.Equal("child", name)
	suite.Equal(data["object"].(map[string]interface{})["child"], val)
	suite.Equal(data["object"], parent)
	suite.True(ok)

	name, val, parent, ok = GetFieldFromName("object.child.name", data)
	suite.Equal("name", name)
	suite.Equal("Michel", val)
	suite.Equal(data["object"].(map[string]interface{})["child"], parent)
	suite.True(ok)

	name, val, parent, ok = GetFieldFromName("object.child.notafield", data)
	suite.Empty(name)
	suite.Nil(val)
	suite.Nil(parent)
	suite.False(ok)
}

func (suite *ValidatorTestSuite) TestTypeDependentAfterConversion() {
	// Before this bug was fixed, type-dependent rules received the original value
	// instead of the converted one, leading to wrong validation.

	data := map[string]interface{}{
		"int": "0",
	}
	set := RuleSet{
		"int": {"integer", "min:2"},
	}

	errors := Validate(data, set, true, "en-US")
	suite.Equal(1, len(errors))

	data["int"] = "2"
	errors = Validate(data, set, true, "en-US")
	suite.Empty(errors)

}

func (suite *ValidatorTestSuite) TestRuleIsType() {
	rule := &Rule{Name: "numeric"}
	suite.True(rule.IsType())

	rule = &Rule{Name: "greater_than"}
	suite.False(rule.IsType())

	rule = &Rule{Name: "nullable"}
	suite.False(rule.IsType())

	suite.Panics(func() {
		rule := &Rule{Name: "not a rule"}
		rule.IsType()
	})
}

func (suite *ValidatorTestSuite) TestRuleIsTypeDependent() {
	rule := &Rule{Name: "numeric"}
	suite.False(rule.IsTypeDependent())

	rule = &Rule{Name: "greater_than"}
	suite.True(rule.IsTypeDependent())

	rule = &Rule{Name: "nullable"}
	suite.False(rule.IsTypeDependent())

	suite.Panics(func() {
		rule := &Rule{Name: "not a rule"}
		rule.IsTypeDependent()
	})
}

func TestValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(ValidatorTestSuite))
}
