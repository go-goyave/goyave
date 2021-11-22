package validation

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
	"goyave.dev/goyave/v4/lang"
	"goyave.dev/goyave/v4/util/fsutil"
	"goyave.dev/goyave/v4/util/sliceutil"
	"goyave.dev/goyave/v4/util/walk"
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

	rule = parseRule("min:5")
	suite.Equal("min", rule.Name)
	suite.Equal(1, len(rule.Params))
	suite.Equal("5", rule.Params[0])

	suite.Panics(func() {
		parseRule("invalid,rule")
	})
}

func (suite *ValidatorTestSuite) TestGetMessage() {
	suite.Equal("The :field is required.", getMessage(&Field{Rules: []*Rule{}, Path: &walk.Path{}}, &Rule{Name: "required"}, reflect.ValueOf("test"), "en-US"))
	suite.Equal("The :field must be at least :min.", getMessage(&Field{Rules: []*Rule{{Name: "numeric"}}, Path: &walk.Path{}}, &Rule{Name: "min"}, reflect.ValueOf(42), "en-US"))

	field := &Field{
		Rules: []*Rule{{Name: "numeric"}},
		Path: &walk.Path{
			Type: walk.PathTypeArray,
			Next: &walk.Path{Type: walk.PathTypeElement},
		},
	}
	suite.Equal("The :field values must be at least :min.", getMessage(field, &Rule{Name: "min"}, reflect.ValueOf(42), "en-US"))

	field = &Field{
		Rules: []*Rule{
			{Name: "numeric"},
			{Name: "min", Params: []string{"5"}},
		},
		Path: &walk.Path{
			Type: walk.PathTypeArray,
			Next: &walk.Path{Type: walk.PathTypeElement},
		},
	}
	suite.Equal("The :field values must be at least :min.", getMessage(field, field.Rules[1], reflect.ValueOf(42), "en-US"))

	// Test type fallback if no type rule is found
	suite.Equal("The :field must be at least :min.", getMessage(&Field{Rules: []*Rule{}, Path: &walk.Path{}}, &Rule{Name: "min"}, reflect.ValueOf(42), "en-US"))
	suite.Equal("The :field must be at least :min characters.", getMessage(&Field{Rules: []*Rule{}, Path: &walk.Path{}}, &Rule{Name: "min"}, reflect.ValueOf("test"), "en-US"))

	// Integer share message with numeric
	suite.Equal("The :field must be at least :min.", getMessage(&Field{Rules: []*Rule{{"integer", nil}}, Path: &walk.Path{}}, &Rule{Name: "min"}, reflect.Value{}, "en-US"))
}

func (suite *ValidatorTestSuite) TestAddRule() {
	suite.Panics(func() {
		AddRule("required", &RuleDefinition{
			Function: func(ctx *Context) bool {
				return false
			},
		})
	})

	AddRule("new_rule", &RuleDefinition{
		Function: func(ctx *Context) bool {
			return true
		},
	})
	_, ok := validationRules["new_rule"]
	suite.True(ok)
}

func (suite *ValidatorTestSuite) TestValidate() {
	errors := Validate(nil, &Rules{}, false, "en-US")
	suite.Equal(1, len(errors))
	suite.Equal("Malformed request", errors["[data]"].Errors[0])

	errors = Validate(nil, RuleSet{}, false, "en-US")
	suite.Equal(1, len(errors))
	suite.Equal("Malformed request", errors["[data]"].Errors[0])

	errors = Validate(nil, &Rules{}, true, "en-US")
	suite.Equal(1, len(errors))
	suite.Equal("Malformed JSON", errors["[data]"].Errors[0])

	errors = Validate(nil, RuleSet{}, true, "en-US")
	suite.Equal(1, len(errors))
	suite.Equal("Malformed JSON", errors["[data]"].Errors[0])

	errors = Validate(map[string]interface{}{
		"string": "hello world",
		"number": 42,
	}, RuleSet{
		"string": List{"required", "string"},
		"number": List{"required", "numeric", "min:10"},
	}, true, "en-US")
	suite.Equal(0, len(errors))

	errors = Validate(map[string]interface{}{
		"string": "hello world",
		"number": 42,
	}, &Rules{
		Fields: FieldMap{
			"string": &Field{
				Rules: []*Rule{
					{Name: "required"},
					{Name: "string"},
				},
			},
			"number": &Field{
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
		"nullField": List{"numeric"},
	}, true, "en-US")
	_, exists := data["nullField"]
	suite.False(exists)
	suite.Equal(0, len(errors))

	data = map[string]interface{}{
		"nullField": nil,
	}
	errors = Validate(data, RuleSet{
		"nullField": List{"required", "nullable", "numeric"},
	}, true, "en-US")
	val, exists := data["nullField"]
	suite.True(exists)
	suite.Nil(val)
	suite.Equal(0, len(errors))

	data = map[string]interface{}{
		"nullField": "test",
	}
	errors = Validate(data, RuleSet{
		"nullField": List{"required", "nullable", "numeric"},
	}, true, "en-US")
	val, exists = data["nullField"]
	suite.True(exists)
	suite.Equal("test", val)
	suite.Equal(1, len(errors))

	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"nullField": &Field{
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
			"text": &Field{
				Rules: []*Rule{
					{Name: "required"},
					{Name: "string"},
				},
			},
		},
	}, true, "en-US")
	suite.Equal(1, len(errors))
	suite.Equal("The text is required.", errors["text"].Errors[0])
}

func (suite *ValidatorTestSuite) TestValidateWithArray() {
	data := map[string]interface{}{
		"string": "hello",
	}
	errors := Validate(data, RuleSet{
		"string": List{"required", "array"},
	}, false, "en-US")
	suite.Equal("array", GetFieldType(data["string"]))
	suite.Equal("hello", data["string"].([]string)[0])
	suite.Equal(0, len(errors))

	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"string": &Field{
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
		"string":   List{"required", "array"},
		"string[]": List{"min:3"},
	}, false, "en-US")
	suite.Len(errors, 0)

	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"string": &Field{
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
				},
			},
			"string[]": &Field{
				Rules: []*Rule{
					{Name: "min", Params: []string{"3"}},
				},
			},
		},
	}, false, "en-US")
	suite.Len(errors, 0)

	data = map[string]interface{}{
		"string": []string{"hi", ",", "there"},
	}
	errors = Validate(data, RuleSet{
		"string":   List{"required", "array"},
		"string[]": List{"min:3"},
	}, false, "en-US")
	suite.Len(errors, 1)

	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"string": &Field{
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
				},
			},
			"string[]": &Field{
				Rules: []*Rule{
					{Name: "min", Params: []string{"3"}},
				},
			},
		},
	}, false, "en-US")
	suite.Len(errors, 1)

	data = map[string]interface{}{
		"string": []string{"johndoe@example.org", "foobar@example.org"},
	}
	errors = Validate(data, RuleSet{
		"string":   List{"required", "array:string"},
		"string[]": List{"email"},
	}, true, "en-US")
	suite.Len(errors, 0)

	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"string": &Field{
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array", Params: []string{"string"}},
				},
			},
			"string[]": &Field{
				Rules: []*Rule{
					{Name: "email"},
				},
			},
		},
	}, true, "en-US")
	suite.Len(errors, 0)

	// Empty array
	data = map[string]interface{}{
		"string": []string{},
	}
	errors = Validate(data, RuleSet{
		"string":   List{"array"},
		"string[]": List{"uuid:5"},
	}, true, "en-US")
	suite.Len(errors, 0)

	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"string": &Field{
				Rules: []*Rule{
					{Name: "array"},
				},
			},
			"string[]": &Field{
				Rules: []*Rule{
					{Name: "uuid", Params: []string{"5"}},
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
		"values":   List{"required", "array"},
		"values[]": List{"array:numeric"},
	}, false, "en-US")
	suite.Len(errors, 0)

	arr, ok := data["values"].([][]float64)
	if suite.True(ok) {
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
			"values": &Field{
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
				},
			},
			"values[]": &Field{
				Rules: []*Rule{
					{Name: "array", Params: []string{"numeric"}},
				},
			},
		},
	}, false, "en-US")
	suite.Len(errors, 0)

	arr, ok = data["values"].([][]float64)
	if suite.True(ok) {
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
		"values":   List{"required", "array"},
		"values[]": List{"array:numeric", "min:3"},
	}, true, "en-US")
	suite.Len(errors, 1)

	_, ok = data["values"].([][]float64)
	suite.True(ok)

	data = map[string]interface{}{
		"values": [][]float64{{5, 8}, {0.6, 7}},
	}
	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"values": &Field{
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
				},
			},
			"values[]": &Field{
				Rules: []*Rule{
					{Name: "array", Params: []string{"numeric"}},
					{Name: "min", Params: []string{"3"}},
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
		"values":   List{"required", "array"},
		"values[]": List{"array:numeric", "min:3"},
	}, true, "en-US")
	suite.Len(errors, 0)

	data = map[string]interface{}{
		"values": [][]float64{{5, 8, 6}, {0.6, 7, 9}},
	}
	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"values": &Field{
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
				},
			},
			"values[]": &Field{
				Rules: []*Rule{
					{Name: "array", Params: []string{"numeric"}},
					{Name: "min", Params: []string{"3"}},
				},
			},
		},
	}, true, "en-US")
	suite.Len(errors, 0)

	data = map[string]interface{}{
		"values": [][]float64{{5, 8}, {3, 7}},
	}
	errors = Validate(data, RuleSet{
		"values":     List{"required", "array"},
		"values[]":   List{"array:numeric"},
		"values[][]": List{"numeric", "min:3"},
	}, true, "en-US")
	suite.Len(errors, 0)

	data = map[string]interface{}{
		"values": [][]float64{{5, 8}, {3, 7}},
	}
	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"values": &Field{
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
				},
			},
			"values[]": &Field{
				Rules: []*Rule{
					{Name: "array", Params: []string{"numeric"}},
				},
			},
			"values[][]": &Field{
				Rules: []*Rule{
					{Name: "min", Params: []string{"3"}},
				},
			},
		},
	}, true, "en-US")
	suite.Len(errors, 0)

	data = map[string]interface{}{
		"values": [][]float64{{5, 8}, {0.6, 7}},
	}
	errors = Validate(data, RuleSet{
		"values":     List{"required", "array"},
		"values[]":   List{"array:numeric"},
		"values[][]": List{"min:3"},
	}, true, "en-US")
	suite.Len(errors, 1)

	data = map[string]interface{}{
		"values": [][]float64{{5, 8}, {0.6, 7}},
	}
	errors = Validate(data, &Rules{
		Fields: FieldMap{
			"values": &Field{
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
				},
			},
			"values[]": &Field{
				Rules: []*Rule{
					{Name: "array", Params: []string{"numeric"}},
				},
			},
			"values[][]": &Field{
				Rules: []*Rule{
					{Name: "min", Params: []string{"3"}},
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
		"values":       List{"required", "array"},
		"values[]":     List{"array", "max:3"},
		"values[][]":   List{"array:numeric"},
		"values[][][]": List{"numeric", "max:4"},
	}, true, "en-US")
	suite.Len(errors, 0)

	arr, ok := data["values"].([][][]float64)
	suite.True(ok)
	if ok {
		suite.Equal(2, len(arr))
		suite.Equal(2, len(arr[0]))
		suite.Equal(3, len(arr[1]))
		suite.Equal(0, len(arr[1][1]))
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
			"values": &Field{
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
				},
			},
			"values[]": &Field{
				Rules: []*Rule{
					{Name: "array"},
					{Name: "max", Params: []string{"3"}},
				},
			},
			"values[][]": &Field{
				Rules: []*Rule{
					{Name: "array", Params: []string{"numeric"}},
				},
			},
			"values[][][]": &Field{
				Rules: []*Rule{
					{Name: "max", Params: []string{"4"}},
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
		"values":       List{"required", "array"},
		"values[]":     List{"array", "max:3"},
		"values[][]":   List{"array:numeric"},
		"values[][][]": List{"max:4"},
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
			"values": &Field{
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
				},
			},
			"values[]": &Field{
				Rules: []*Rule{
					{Name: "array"},
					{Name: "max", Params: []string{"3"}},
				},
			},
			"values[][]": &Field{
				Rules: []*Rule{
					{Name: "array", Params: []string{"numeric"}},
				},
			},
			"values[][][]": &Field{
				Rules: []*Rule{
					{Name: "max", Params: []string{"4"}},
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
		"values":       List{"required", "array"},
		"values[]":     List{"array", "max:3"},
		"values[][]":   List{"array:numeric"},
		"values[][][]": List{"max:4"},
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
			"values": &Field{
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
				},
			},
			"values[]": &Field{
				Rules: []*Rule{
					{Name: "array"},
					{Name: "max", Params: []string{"3"}},
				},
			},
			"values[][]": &Field{
				Rules: []*Rule{
					{Name: "array", Params: []string{"numeric"}},
				},
			},
			"values[][][]": &Field{
				Rules: []*Rule{
					{Name: "max", Params: []string{"4"}},
				},
			},
		},
	}, true, "en-US")
	suite.Len(errors, 1)
}

func (suite *ValidatorTestSuite) TestValidateNDimensionalArrayParentSkipped() {
	// array[][] but without array[]
	data := map[string]interface{}{
		"values": [][]interface{}{{"0.5", 1.42}, {0.6, 7}},
	}
	errors := Validate(data, RuleSet{
		"values[][]": List{"numeric"},
	}, false, "en-US")
	suite.Len(errors, 0)

	// Should still be generic slice
	arr, ok := data["values"].([][]interface{})
	if suite.True(ok) {
		suite.Equal(2, len(arr))
		suite.Equal(2, len(arr[0]))
		suite.Equal(2, len(arr[1]))
		suite.Equal(0.5, arr[0][0])
		suite.Equal(1.42, arr[0][1])
		suite.Equal(0.6, arr[1][0])
		suite.Equal(7.0, arr[1][1])
		suite.IsType([]interface{}{}, arr[0])
		suite.IsType([]interface{}{}, arr[1])
	}

	data = map[string]interface{}{
		"values": [][]interface{}{{"0.5", 1.42}, {0.6, 7}},
	}
	errors = Validate(data, RuleSet{
		"values[]": List{"array:numeric"},
	}, false, "en-US")
	suite.Len(errors, 0)

	arr2, ok := data["values"].([][]interface{})
	if suite.True(ok) {
		suite.Equal(2, len(arr2))
		suite.Equal(2, len(arr2[0]))
		suite.Equal(2, len(arr2[1]))
		suite.Equal(0.5, arr2[0][0])
		suite.Equal(1.42, arr2[0][1])
		suite.Equal(0.6, arr2[1][0])
		suite.Equal(7.0, arr2[1][1])
		suite.IsType([]interface{}{}, arr2[0])
		suite.IsType([]interface{}{}, arr2[1])
	}

}

func (suite *ValidatorTestSuite) TestFieldCheck() {
	suite.NotPanics(func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "required"},
				{Name: "numeric"},
			},
		}

		field.Check()

		suite.True(field.isRequired)
		suite.False(field.isArray)
		suite.False(field.isObject)
		suite.False(field.isNullable)
		suite.False(field.IsArray())
		suite.False(field.IsObject())
		suite.False(field.IsNullable())
		suite.True(field.IsRequired())
	})

	suite.NotPanics(func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "nullable"},
				{Name: "array"},
			},
		}

		field.Check()

		suite.False(field.isRequired)
		suite.True(field.isArray)
		suite.False(field.isObject)
		suite.True(field.isNullable)
		suite.True(field.IsArray())
		suite.False(field.IsObject())
		suite.True(field.IsNullable())
		suite.False(field.IsRequired())
	})

	suite.NotPanics(func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "nullable"},
				{Name: "object"},
			},
		}

		field.Check()

		suite.False(field.isRequired)
		suite.False(field.isArray)
		suite.True(field.isObject)
		suite.True(field.isNullable)
		suite.False(field.IsArray())
		suite.True(field.IsObject())
		suite.True(field.IsNullable())
		suite.False(field.IsRequired())
	})

	suite.Panics(func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "required"},
				{Name: "not a rule"},
			},
		}

		field.Check()
	})
}

func (suite *ValidatorTestSuite) TestFieldCheckArrayProhibitedRules() {
	prohibitedRules := []string{
		"file", "mime", "image", "extension", "count",
		"count_min", "count_max", "count_between",
	}
	for _, v := range prohibitedRules {
		suite.Panics(func() {
			field := &Field{
				Rules: []*Rule{
					{Name: v},
				},
				Path: &walk.Path{Type: walk.PathTypeArray},
			}
			field.Check()
		})
	}
}

func (suite *ValidatorTestSuite) TestParseRuleSet() {
	set := RuleSet{
		"string":   List{"required", "array:string"},
		"string[]": List{"min:3"},
		"number":   List{"numeric"},
	}

	rules := set.parse()
	suite.Len(rules.Fields, 2)
	suite.Len(rules.Fields["string"].(*Field).Rules, 2)
	suite.Equal(&Rule{Name: "required", Params: []string{}}, rules.Fields["string"].(*Field).Rules[0])
	suite.Equal(&Rule{Name: "array", Params: []string{"string"}}, rules.Fields["string"].(*Field).Rules[1])
	suite.Equal(&Rule{Name: "min", Params: []string{"3"}}, rules.Fields["string"].(*Field).Elements.Rules[0])

	expectedPath := &walk.Path{
		Type: walk.PathTypeArray,
		Next: &walk.Path{
			Type: walk.PathTypeElement,
		},
	}
	suite.Equal(expectedPath, rules.Fields["string"].(*Field).Elements.Path)
	suite.Len(rules.Fields["number"].(*Field).Rules, 1)
	suite.Equal(&Rule{Name: "numeric", Params: []string{}}, rules.Fields["number"].(*Field).Rules[0])

	parsed := set.AsRules()
	suite.Equal(rules.Fields, parsed.Fields)
	suite.Equal(rules.checked, parsed.checked)

	suite.True(rules.checked)
	// Resulting Rules should be checked after parsing
	suite.Panics(func() {
		set := RuleSet{
			"string":   List{"required", "not a rule"},
			"string[]": List{"min:3"},
		}
		set.parse()
	})
}

func (suite *ValidatorTestSuite) TestAsRules() {
	rules := &Rules{
		Fields: FieldMap{
			"values": &Field{
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
				},
			},
			"values[]": &Field{
				Rules: []*Rule{
					{Name: "array"},
					{Name: "max", Params: []string{"3"}},
				},
			},
			"values[][]": &Field{
				Rules: []*Rule{
					{Name: "array", Params: []string{"numeric"}},
				},
			},
			"values[][][]": &Field{
				Rules: []*Rule{
					{Name: "max", Params: []string{"4"}},
				},
			},
		},
	}
	suite.Equal(rules, rules.AsRules())

	suite.Panics(func() {
		rules := &Rules{
			Fields: FieldMap{
				"values": &Field{
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
			"values": &Field{
				Rules: []*Rule{
					{Name: "required"},
					{Name: "array"},
				},
			},
			"values[]": &Field{
				Rules: []*Rule{
					{Name: "array"},
					{Name: "max", Params: []string{"3"}},
				},
			},
			"values[][]": &Field{
				Rules: []*Rule{
					{Name: "array", Params: []string{"numeric"}},
				},
			},
			"values[][][]": &Field{
				Rules: []*Rule{
					{Name: "max", Params: []string{"4"}},
				},
			},
		},
	}
	suite.False(rules.checked)
	rules.Check()
	suite.True(rules.checked)

	// Check should not be executed multiple times
	rules.Fields["values"].(*Field).Rules[0].Name = "not a rule"
	suite.NotPanics(func() {
		rules.Check()
	})
}

func (suite *ValidatorTestSuite) TestGetFieldType() {
	suite.Equal("numeric", getFieldType(reflect.ValueOf(1)))
	suite.Equal("numeric", getFieldType(reflect.ValueOf(1.1)))
	suite.Equal("numeric", getFieldType(reflect.ValueOf(uint(1))))
	suite.Equal("numeric", getFieldType(reflect.ValueOf(float32(1))))
	suite.Equal("string", getFieldType(reflect.ValueOf("hello")))
	suite.Equal("array", getFieldType(reflect.ValueOf([]string{"hello", "world"})))
	suite.Equal("file", getFieldType(reflect.ValueOf([]fsutil.File{})))
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
		"int": List{"integer", "min:2"},
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

func (suite *ValidatorTestSuite) TestRuleComparisonNonGuaranteedOrder() { // https://github.com/go-goyave/goyave/issues/144
	rules := &Rules{
		Fields: FieldMap{
			"start": &Field{Rules: []*Rule{
				{Name: "date", Params: []string{"02-01-2006"}}, // Use another date format to prevent auto-conversion
			}},
			"end": &Field{Rules: []*Rule{
				{Name: "date", Params: []string{"02-01-2006"}},
				{Name: "after", Params: []string{"start"}},
			}},
		},
	}

	// Test several times to check if even if map iteration order is not guaranteed,
	// validation still behaves as expected.
	for i := 0; i < 100; i++ {
		data := map[string]interface{}{
			"start": "05-06-2008",
			"end":   "05-06-2009",
		}
		err := Validate(data, rules, true, "en-US")
		suite.Empty(err)
	}
}

func (suite *ValidatorTestSuite) TestSortKeys() {
	rules := &Rules{
		Fields: FieldMap{
			CurrentElement: &Field{Rules: []*Rule{}},
			"text": &Field{Rules: []*Rule{
				{Name: "string"},
			}},
			"mid": &Field{Rules: []*Rule{
				{Name: "date"},
				{Name: "after", Params: []string{"start"}},
				{Name: "before", Params: []string{"end"}},
			}},
			"end": &Field{Rules: []*Rule{
				{Name: "date"},
				{Name: "after", Params: []string{"start"}},
			}},
			"start": &Field{Rules: []*Rule{
				{Name: "date"},
			}},
		},
	}
	rules.sortKeys()

	// Expect [text start end mid]
	// Use relative indexes because order is not guaranteed (text may be anywhere)
	indexStart := sliceutil.IndexOfStr(rules.sortedKeys, "start")
	indexEnd := sliceutil.IndexOfStr(rules.sortedKeys, "end")
	indexMid := sliceutil.IndexOfStr(rules.sortedKeys, "mid")
	suite.Greater(indexEnd, indexStart)
	suite.Greater(indexMid, indexStart)
	suite.Greater(indexMid, indexEnd)
	suite.Contains(rules.sortedKeys, "start")
	suite.Contains(rules.sortedKeys, "mid")
	suite.Contains(rules.sortedKeys, "end")
	suite.Contains(rules.sortedKeys, "text")
	suite.NotContains(rules.sortedKeys, CurrentElement)
}

func (suite *ValidatorTestSuite) TestSortKeysIncoherent() {
	rules := &Rules{
		Fields: FieldMap{
			"end": &Field{Rules: []*Rule{
				{Name: "date"},
				{Name: "after", Params: []string{"start"}},
			}},
			"start": &Field{Rules: []*Rule{
				{Name: "date"},
				{Name: "after", Params: []string{"end"}},
			}},
		},
	}
	rules.sortKeys()

	// In that case, whatever order can be used but consistency not ensured
	// In any case, this shouldn't crash
	suite.Contains(rules.sortedKeys, "start")
	suite.Contains(rules.sortedKeys, "end")
}

func (suite *ValidatorTestSuite) TestSortKeysMultipleComparedFields() {
	rules := &Rules{
		Fields: FieldMap{
			"text": &Field{Rules: []*Rule{
				{Name: "string"},
			}},
			"mid": &Field{Rules: []*Rule{
				{Name: "date"},
				{Name: "after", Params: []string{"start"}},
				{Name: "before", Params: []string{"end"}},
			}},
			"end": &Field{Rules: []*Rule{
				{Name: "date"},
				{Name: "date_between", Params: []string{"start", "end"}},
			}},
			"start": &Field{Rules: []*Rule{
				{Name: "date"},
			}},
		},
	}
	rules.sortKeys()
	indexStart := sliceutil.IndexOfStr(rules.sortedKeys, "start")
	indexEnd := sliceutil.IndexOfStr(rules.sortedKeys, "end")
	indexMid := sliceutil.IndexOfStr(rules.sortedKeys, "mid")
	suite.Greater(indexEnd, indexStart)
	suite.Greater(indexMid, indexStart)
	suite.Greater(indexMid, indexEnd)
	suite.Contains(rules.sortedKeys, "start")
	suite.Contains(rules.sortedKeys, "mid")
	suite.Contains(rules.sortedKeys, "end")
	suite.Contains(rules.sortedKeys, "text")
}

func (suite *ValidatorTestSuite) TestSortKeysBuiltinRules() {
	// Tests that all rules that are supposed to be comparing fields
	// are sorted correctly.
	suite.testSortKeysWithRule("greater_than")
	suite.testSortKeysWithRule("greater_than_equal")
	suite.testSortKeysWithRule("lower_than")
	suite.testSortKeysWithRule("lower_than_equal")
	suite.testSortKeysWithRule("in_array")
	suite.testSortKeysWithRule("not_in_array")
	suite.testSortKeysWithRule("same")
	suite.testSortKeysWithRule("different")
	suite.testSortKeysWithRule("before")
	suite.testSortKeysWithRule("before_equal")
	suite.testSortKeysWithRule("after")
	suite.testSortKeysWithRule("after_equal")
	suite.testSortKeysWithRule("date_equals")
	suite.testSortKeysWithRule("date_between")
}

func (suite *ValidatorTestSuite) TestSortKeysWithNullable() {
	rules := RuleSet{
		"field1": List{"nullable", "string"},
		"field2": List{"required", "string"},
	}
	rules.AsRules().sortKeys()
}

func (suite *ValidatorTestSuite) testSortKeysWithRule(rule string) {
	rules := &Rules{
		Fields: FieldMap{
			"one": &Field{Rules: []*Rule{
				{Name: "string"},
				{Name: rule, Params: []string{"two"}},
			}},
			"two": &Field{Rules: []*Rule{
				{Name: "string"},
			}},
		},
	}
	rules.sortKeys()
	suite.Equal([]string{"two", "one"}, rules.sortedKeys)
}

func (suite *ValidatorTestSuite) TestValidateObjectInArray() {
	// array[].field
	data := map[string]interface{}{
		"array": []interface{}{
			map[string]interface{}{"field": "1"},
		},
	}
	errors := Validate(data, RuleSet{
		"array":         List{"required", "array:object"},
		"array[].field": List{"numeric", "max:3"},
	}, true, "en-US")
	suite.Len(errors, 0)

	arr, ok := data["array"].([]map[string]interface{})
	if suite.True(ok) {
		suite.Len(arr, 1)
		suite.Equal(1.0, arr[0]["field"])
	}

	// array[][].field
	data = map[string]interface{}{
		"array": [][]interface{}{
			{
				map[string]interface{}{"field": "1"},
			},
		},
	}
	errors = Validate(data, RuleSet{
		"array":           List{"required", "array"},
		"array[]":         List{"required", "array:object"},
		"array[][].field": List{"numeric", "max:3"},
	}, true, "en-US")
	suite.Len(errors, 0)

	arr2, ok := data["array"].([][]map[string]interface{})
	if suite.True(ok) {
		suite.Len(arr2, 1)
		suite.Equal(1.0, arr2[0][0]["field"])
	}

	// array[].subarray[]
	data = map[string]interface{}{
		"array": []interface{}{
			map[string]interface{}{"subarray": []interface{}{"5"}},
		},
	}
	errors = Validate(data, RuleSet{
		"array":            List{"required", "array"},
		"array[].subarray": List{"array:numeric"},
	}, true, "en-US")
	suite.Len(errors, 0)

	arr3, ok := data["array"].([]map[string]interface{})
	if suite.True(ok) {
		suite.Len(arr3, 1)
		suite.IsType([]float64{}, arr3[0]["subarray"])
		suite.Equal(5.0, arr3[0]["subarray"].([]float64)[0])
	}

	// array[].subarray[].field
	data = map[string]interface{}{
		"array": []interface{}{
			map[string]interface{}{"subarray": []map[string]interface{}{
				{"field": "5"},
			}},
		},
	}
	errors = Validate(data, RuleSet{
		"array":                    List{"required", "array"},
		"array[].subarray":         List{"array:object"},
		"array[].subarray[].field": List{"numeric"},
	}, true, "en-US")
	suite.Len(errors, 0)

	arr4, ok := data["array"].([]map[string]interface{})
	if suite.True(ok) {
		suite.Len(arr4, 1)
		suite.Equal(5.0, arr4[0]["subarray"].([]map[string]interface{})[0]["field"])
	}

	// array[].subarray[].field[]
	data = map[string]interface{}{
		"array": []interface{}{
			map[string]interface{}{"subarray": []map[string]interface{}{
				{"field": []interface{}{"5"}},
			}},
		},
	}
	errors = Validate(data, RuleSet{
		"array":                    List{"required", "array"},
		"array[].subarray":         List{"array:object"},
		"array[].subarray[].field": List{"array:numeric"},
	}, true, "en-US")
	suite.Len(errors, 0)

	arr5, ok := data["array"].([]map[string]interface{})
	if suite.True(ok) {
		suite.Len(arr5, 1)
		field := arr5[0]["subarray"].([]map[string]interface{})[0]["field"]
		suite.IsType([]float64{}, field)
		suite.Equal(5.0, field.([]float64)[0])
	}

	// Same but without other fields validation
	data = map[string]interface{}{
		"array": []interface{}{
			map[string]interface{}{"subarray": []map[string]interface{}{
				{"field": []interface{}{"5"}},
			}},
		},
	}
	errors = Validate(data, RuleSet{
		"array[].subarray[].field": List{"array:numeric"},
	}, true, "en-US")
	suite.Len(errors, 0)

	arr6, ok := data["array"].([]interface{})
	if suite.True(ok) {
		suite.Len(arr6, 1)
		field := arr6[0].(map[string]interface{})["subarray"].([]map[string]interface{})[0]["field"]
		suite.IsType([]float64{}, field)
		suite.Equal(5.0, field.([]float64)[0])
	}
}

func (suite *ValidatorTestSuite) TestValidateObjectInArrayErrors() {
	// array[].field
	data := map[string]interface{}{
		"array": []interface{}{
			map[string]interface{}{"field": "1"},
			map[string]interface{}{"field": "2"},
			map[string]interface{}{"field": "5"},
		},
	}
	errors := Validate(data, RuleSet{
		"array":         List{"required", "array:object"},
		"array[].field": List{"numeric", "min:3"},
	}, true, "en-US")

	expected := Errors{
		"array": &FieldErrors{
			Elements: ArrayErrors{
				0: &FieldErrors{
					Fields: Errors{
						"field": &FieldErrors{Errors: []string{"The field must be at least 3."}},
					},
				},
				1: &FieldErrors{
					Fields: Errors{
						"field": &FieldErrors{Errors: []string{"The field must be at least 3."}},
					},
				},
			},
		},
	}
	suite.Equal(expected, errors)

	// array[].subarray[].field
	data = map[string]interface{}{
		"array": []interface{}{
			map[string]interface{}{"subarray": []map[string]interface{}{
				{"field": "1"},
				{"field": "6"},
			}},
			map[string]interface{}{"subarray": []map[string]interface{}{
				{"field": "2"},
			}},
		},
	}
	errors = Validate(data, RuleSet{
		"array":                    List{"required", "array"},
		"array[].subarray":         List{"array:object"},
		"array[].subarray[].field": List{"numeric", "min:3"},
	}, true, "en-US")
	expected = Errors{
		"array": &FieldErrors{
			Elements: ArrayErrors{
				0: &FieldErrors{
					Fields: Errors{
						"subarray": &FieldErrors{
							Elements: ArrayErrors{
								0: &FieldErrors{
									Fields: Errors{
										"field": &FieldErrors{Errors: []string{"The field must be at least 3."}},
									},
								},
							},
						},
					},
				},
				1: &FieldErrors{
					Fields: Errors{
						"subarray": &FieldErrors{
							Elements: ArrayErrors{
								0: &FieldErrors{
									Fields: Errors{
										"field": &FieldErrors{Errors: []string{"The field must be at least 3."}},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	suite.Equal(expected, errors)
}

func (suite *ValidatorTestSuite) TestValidateRequiredInObjectInArray() {
	data := map[string]interface{}{
		"array": []interface{}{
			map[string]interface{}{"field": "1"},
			map[string]interface{}{},
			map[string]interface{}{"field": "5"},
		},
	}
	errors := Validate(data, RuleSet{
		"array":         List{"required", "array:object"},
		"array[].field": List{"required", "numeric", "min:3"},
	}, true, "en-US")
	expected := Errors{
		"array": &FieldErrors{
			Elements: ArrayErrors{
				0: &FieldErrors{
					Fields: Errors{
						"field": &FieldErrors{Errors: []string{"The field must be at least 3."}},
					},
				},
				1: &FieldErrors{
					Fields: Errors{
						"field": &FieldErrors{Errors: []string{"The field is required.", "The field must be numeric."}},
					},
				},
			},
		},
	}
	suite.Equal(expected, errors)

	data = map[string]interface{}{
		"array": [][]interface{}{
			{
				map[string]interface{}{},
				map[string]interface{}{"field": 3.0},
			},
		},
	}
	errors = Validate(data, RuleSet{
		"array":           List{"required", "array"},
		"array[]":         List{"required", "array:object"},
		"array[][].field": List{"required", "numeric", "max:3"},
	}, true, "en-US")
	expected = Errors{
		"array": &FieldErrors{
			Elements: ArrayErrors{
				0: &FieldErrors{
					Elements: ArrayErrors{
						0: &FieldErrors{
							Fields: Errors{
								"field": &FieldErrors{Errors: []string{"The field is required.", "The field must be numeric."}},
							},
						},
					},
				},
			},
		},
	}
	suite.Equal(expected, errors)
}

func (suite *ValidatorTestSuite) TestValidateWrongBody() {
	data := map[string]interface{}{
		"array": []interface{}{
			map[string]interface{}{"field": "1"},
			map[string]interface{}{},
			map[string]interface{}{"field": "5"},
		},
		"narray": [][]interface{}{
			{
				map[string]interface{}{"field": 1},
				"a",
				"b",
			},
		},
		"object": map[string]interface{}{
			"array": []interface{}{
				5,
				[]string{"a", "b"},
				[]string{},
				map[string]interface{}{
					"field": "1",
				},
			},
		},
		"object2": map[string]interface{}{
			"array": []interface{}{
				6,
				map[string]interface{}{
					"field": "1",
				},
			},
		},
		"edgecase": []string{},
	}

	errors := Validate(data, RuleSet{
		"array":                         List{"required", "array:object"},
		"array[]":                       List{"required", "object"},
		"array[].field":                 List{"required", "numeric", "max:3"},
		"narray[][]":                    List{"object"},
		"narray[][].field":              List{"required", "numeric"},
		"object":                        List{"required", "object"},
		"object.array":                  List{"required", "array"},
		"object.array[]":                List{"required", "array:string"},
		"object.array[][]":              List{"required", "string", "min:2"},
		"object2.array":                 List{"required", "array:object"},
		"object2.array[].field":         List{"required", "object"},
		"edgecase[][][][]":              List{"required", "string"},
		"missingobject":                 List{"required", "object"},
		"missingobject.subobject":       List{"required", "object"},
		"missingobject.subobject.field": List{"required", "string"},
		"missingobject.array":           List{"required", "array:object"},
		"missingobject.array[].field":   List{"required", "array:string"},
		"missingobject.array[].field[]": List{"required", "string"},
	}, true, "en-US")

	expected := Errors{
		"array": &FieldErrors{
			Elements: ArrayErrors{
				1: &FieldErrors{
					Fields: Errors{
						"field": &FieldErrors{Errors: []string{"The field is required.", "The field must be numeric."}},
					},
				},
				2: &FieldErrors{
					Fields: Errors{
						"field": &FieldErrors{Errors: []string{"The field may not be greater than 3."}},
					},
				},
			},
		},
		"narray": &FieldErrors{
			Elements: ArrayErrors{
				0: &FieldErrors{
					Elements: ArrayErrors{
						1: &FieldErrors{
							Errors: []string{"The narray[] values must be objects."},
						},
						2: &FieldErrors{
							Errors: []string{"The narray[] values must be objects."},
						},
					},
				},
			},
		},
		"object": &FieldErrors{
			Fields: Errors{
				"array": &FieldErrors{
					Elements: ArrayErrors{
						0: &FieldErrors{
							Errors: []string{"The array values must be arrays."},
						},
						1: &FieldErrors{
							Elements: ArrayErrors{
								0: &FieldErrors{Errors: []string{"The array[] values must be at least 2 characters."}},
								1: &FieldErrors{Errors: []string{"The array[] values must be at least 2 characters."}},
							},
						},
						2: &FieldErrors{
							Elements: ArrayErrors{
								-1: &FieldErrors{Errors: []string{"The array[] values are required.", "The array[] values must be strings."}},
							},
						},
						3: &FieldErrors{
							Errors: []string{"The array values must be arrays."},
						},
					},
				},
			},
		},
		"object2": &FieldErrors{
			Fields: Errors{
				"array": &FieldErrors{
					Errors: []string{"The array must be an array."},
					Elements: ArrayErrors{
						1: &FieldErrors{
							Fields: Errors{
								"field": &FieldErrors{Errors: []string{"The field must be an object."}},
							},
						},
					},
				},
			},
		},
		"missingobject": &FieldErrors{
			Errors: []string{"The missingobject is required.", "The missingobject must be an object."},
		},
	}

	suite.Equal(expected, errors)
}

func (suite *ValidatorTestSuite) TestValidateArrayNoConversionIfAllElementsNotSameType() {
	data := map[string]interface{}{
		"array": [][]interface{}{
			{5, 6.0},
		},
	}

	errors := Validate(data, RuleSet{
		"array":   List{"required", "array"},
		"array[]": List{"required", "array"},
	}, true, "en-US")
	suite.Empty(errors)

	a, ok := data["array"].([][]interface{})
	if suite.True(ok) {
		suite.Equal(5, a[0][0])
		suite.Equal(6.0, a[0][1])
	}
}

func (suite *ValidatorTestSuite) TestInvalidPath() {
	suite.Panics(func() {
		RuleSet{
			"invalid path.": List{"required", "string"},
		}.AsRules()
	})
}

func (suite *ValidatorTestSuite) TestValidateContext() {
	data := map[string]interface{}{
		"a": "b",
	}

	validationRules["test_rule"] = &RuleDefinition{}
	defer func() {
		delete(validationRules, "test_rule")
	}()
	rules := RuleSet{
		"a": List{"required", "test_rule"},
	}.AsRules()
	validationRules["test_rule"].Function = func(c *Context) bool {
		suite.Equal(data, c.Data)
		suite.Equal("b", c.Value)
		suite.Equal(data, c.Parent)
		suite.Equal("a", c.Name)
		suite.Equal(rules.Fields["a"], c.Field)
		suite.Equal(&Rule{Name: "test_rule", Params: []string{}}, c.Rule)
		suite.NotNil(c.Now)
		return true
	}

	Validate(data, rules, true, "en-US")
}

func (suite *ValidatorTestSuite) TestRuleSetComposition() {
	user := RuleSet{
		CurrentElement: List{"required", "object"},
		"name":         List{"required", "string"},
		"email":        List{"required", "email"},
	}

	rules := RuleSet{
		CurrentElement: List{"required", "object"},
		"user":         user,
		"field":        List{"nullable", "numeric"},
	}
	rules.processComposition()

	expected := RuleSet{
		"user":       List{"required", "object"},
		"user.name":  List{"required", "string"},
		"user.email": List{"required", "email"},
		"field":      List{"nullable", "numeric"},
	}

	suite.Equal(expected, rules)

	// Original is not altered
	expectedUser := RuleSet{
		CurrentElement: List{"required", "object"},
		"name":         List{"required", "string"},
		"email":        List{"required", "email"},
	}
	suite.Equal(expectedUser, user)
}

func (suite *ValidatorTestSuite) TestRulesComposition() {
	user := &Rules{
		Fields: FieldMap{
			CurrentElement: &Field{
				Rules: []*Rule{{Name: "required"}, {Name: "object"}},
			},
			"name": &Field{
				Rules: []*Rule{{Name: "required"}, {Name: "object"}},
			},
			"email": &Field{
				Rules: []*Rule{{Name: "required"}, {Name: "email"}},
			},
		},
	}

	rules := &Rules{
		Fields: FieldMap{
			CurrentElement: &Field{
				Rules: []*Rule{{Name: "required"}, {Name: "object"}},
			},
			"user": user,
			"field": &Field{
				Rules: []*Rule{{Name: "nullable"}, {Name: "numeric"}},
			},
		},
	}
	rules.processComposition()

	expected := &Rules{
		Fields: FieldMap{
			"user": &Field{
				Rules: []*Rule{{Name: "required"}, {Name: "object"}},
			},
			"user.name": &Field{
				Rules: []*Rule{{Name: "required"}, {Name: "object"}},
			},
			"user.email": &Field{
				Rules: []*Rule{{Name: "required"}, {Name: "email"}},
			},
			"field": &Field{
				Rules: []*Rule{{Name: "nullable"}, {Name: "numeric"}},
			},
		},
	}

	suite.Equal(expected, rules)

	// Original is not altered
	expectedUser := &Rules{
		Fields: FieldMap{
			CurrentElement: &Field{
				Rules: []*Rule{{Name: "required"}, {Name: "object"}},
			},
			"name": &Field{
				Rules: []*Rule{{Name: "required"}, {Name: "object"}},
			},
			"email": &Field{
				Rules: []*Rule{{Name: "required"}, {Name: "email"}},
			},
		},
	}
	suite.Equal(expectedUser, user)
}

func (suite *ValidatorTestSuite) TestValidateNonNullableInQuery() {
	data := map[string]interface{}{}

	rules := RuleSet{
		"page":   List{"integer"},
		"search": List{"string"},
	}

	errors := Validate(data, rules, false, "en-US")
	suite.Nil(errors)

	suite.Empty(data)
}

func (suite *ValidatorTestSuite) TestValidateRequiredFieldInNonRequiredObject() {
	// Object is not required so if object is not given, "object.field" should pass validation.
	data := map[string]interface{}{}
	rules := RuleSet{
		"object":       List{"object"},
		"object.field": List{"required", "string"},
	}

	errors := Validate(data, rules, false, "en-US")
	suite.Nil(errors)
}

func (suite *ValidatorTestSuite) TestValidateRequiredFieldInNonRequiredArray() {
	// Array is not required so if array is not given, elements should pass validation.
	// Array elements should not fail validation if the array is empty.
	data := map[string]interface{}{}
	rules := RuleSet{
		"array":         List{"array:object"},
		"array[].field": List{"required", "string"},
	}

	errors := Validate(data, rules, false, "en-US")
	suite.Nil(errors)

	data["array"] = []map[string]interface{}{}
	errors = Validate(data, rules, false, "en-US")
	suite.Nil(errors)
}

func (suite *ValidatorTestSuite) TestValidateEmptyStringNotRequired() {
	rules := RuleSet{
		"name": List{"between:5,50"},
	}
	data := map[string]interface{}{
		"name": "",
	}

	errors := Validate(data, rules, false, "en-US")
	suite.NotNil(errors)
}

func TestValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(ValidatorTestSuite))
}
