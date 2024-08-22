package validation

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/lang"
	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/fsutil"
	"goyave.dev/goyave/v5/util/walk"
)

type extraKey struct{}

type testValidator struct {
	BaseValidator
	placeholders    func(ctx *Context) []string
	validateFunc    func(c component, ctx *Context) bool
	isType          bool
	isTypeDependent bool
}

func (v *testValidator) Validate(ctx *Context) bool {
	return v.validateFunc(v.component, ctx)
}

func (v *testValidator) IsTypeDependent() bool {
	return v.isTypeDependent
}

func (v *testValidator) IsType() bool {
	return v.isType
}

func (v *testValidator) MessagePlaceholders(ctx *Context) []string {
	if v.placeholders != nil {
		return v.placeholders(ctx)
	}
	return []string{}
}

func (v *testValidator) Name() string {
	return "test_validator"
}

func TestComponent(t *testing.T) {
	c := &component{
		db:     &gorm.DB{},
		config: config.LoadDefault(),
		lang:   lang.New().GetDefault(),
		logger: slog.New(slog.NewDevModeHandler(bytes.NewBuffer(make([]byte, 0, 10)), nil)),
	}

	assert.Equal(t, c.db, c.DB())
	assert.Equal(t, c.config, c.Config())
	assert.Equal(t, c.lang, c.Lang())
	assert.Equal(t, c.logger, c.Logger())

	t.Run("unset", func(t *testing.T) {
		c := &component{}
		assert.Panics(t, func() { c.DB() })
		assert.Panics(t, func() { c.Config() })
		assert.Panics(t, func() { c.Lang() })
		assert.Panics(t, func() { c.Logger() })
	})
}

func TestContext(t *testing.T) {
	t.Run("Errors", func(t *testing.T) {
		c := &Context{}
		c.AddError(fmt.Errorf("err1"), fmt.Errorf("err2"), fmt.Errorf("err3"))
		assert.Len(t, c.errors, 3)
		for i, e := range c.Errors() {
			err, ok := e.(*errors.Error)
			if assert.True(t, ok) {
				assert.Equal(t, []error{fmt.Errorf("err%d", i+1)}, err.Unwrap())
			}
		}
	})

	t.Run("AddArrayElementValidationErrors", func(t *testing.T) {
		c := &Context{}
		c.AddArrayElementValidationErrors(1, 2, 3)
		assert.Equal(t, []int{1, 2, 3}, c.arrayElementErrors)
	})
}

func TestGetFieldName(t *testing.T) {
	language := lang.New().GetDefault()

	cases := []struct {
		path *walk.Path
		desc string
		want string
	}{
		{desc: "CurrentElement", path: walk.MustParse(""), want: "body"},
		{desc: "untranslated_property", path: walk.MustParse("property"), want: "property"},
		{desc: "untranslated_object.property", path: walk.MustParse("object.property"), want: "property"},
		{desc: "untranslated_array[]", path: walk.MustParse("array[]"), want: "array"},
		{desc: "translated_email", path: walk.MustParse("email"), want: "email address"},
		{desc: "translated_object.email", path: walk.MustParse("object.email"), want: "email address"},
		{desc: "translated_email[]", path: walk.MustParse("email[]"), want: "email address"},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			assert.Equal(t, c.want, GetFieldName(language, c.path))
		})
	}
}

func TestGetFieldType(t *testing.T) {
	cases := []struct {
		desc  string
		value any
		want  string
	}{
		{desc: "numeric_int", value: 1, want: FieldTypeNumeric},
		{desc: "numeric_int64", value: int64(1), want: FieldTypeNumeric},
		{desc: "numeric_uint", value: uint(1), want: FieldTypeNumeric},
		{desc: "numeric_uint64", value: uint64(1), want: FieldTypeNumeric},
		{desc: "numeric_float32", value: float32(1), want: FieldTypeNumeric},
		{desc: "numeric_float64", value: float64(1), want: FieldTypeNumeric},
		{desc: "string", value: "", want: FieldTypeString},
		{desc: "bool", value: true, want: FieldTypeBool},
		{desc: "slice_int", value: []int{}, want: FieldTypeArray},
		{desc: "slice_string", value: []string{}, want: FieldTypeArray},
		{desc: "slice_file", value: []fsutil.File{}, want: FieldTypeFile},
		{desc: "object", value: map[string]any{}, want: FieldTypeObject},
		{desc: "unsupported", value: struct{}{}, want: FieldTypeUnsupported},
		{desc: "unsupported_uintptr", value: uintptr(1), want: FieldTypeUnsupported},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			assert.Equal(t, c.want, GetFieldType(c.value))
		})
	}
}

func TestValidateExtraNotNil(t *testing.T) {
	options := &Options{
		Data:     nil,
		Language: lang.New().GetDefault(),
		Rules: RuleSet{
			{Path: CurrentElement, Rules: List{Required()}},
		},
	}
	Validate(options)
	assert.NotNil(t, options.Extra)
}

func TestValidate(t *testing.T) {
	cases := []struct {
		desc                 string
		wantData             any
		options              *Options
		wantValidationErrors *Errors
		wantErrors           []error
	}{
		{
			desc: "nil_data",
			options: &Options{
				Data:     nil,
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: CurrentElement, Rules: List{Required()}},
					{Path: "property", Rules: List{Required()}},
				},
			},
			wantValidationErrors: &Errors{
				Errors: []string{"The body is required."},
				// No field errors, they are skipped because parent is absent
			},
			wantErrors: nil,
		},
		{
			desc: "context",
			options: &Options{
				Data:     map[string]any{"property": "value"},
				DB:       &gorm.DB{},
				Logger:   slog.New(slog.NewDevModeHandler(bytes.NewBuffer(make([]byte, 0, 10)), nil)),
				Extra:    map[any]any{extraKey{}: "value"},
				Language: lang.New().GetDefault(),
				Config:   config.LoadDefault(),
				Rules: RuleSet{
					{Path: "property", Rules: List{&testValidator{
						validateFunc: func(c component, ctx *Context) bool {
							// Validator init called
							assert.NotNil(t, c.db)
							assert.NotNil(t, c.config)
							assert.NotNil(t, c.lang)
							assert.NotNil(t, c.logger)

							// Context content
							assert.Equal(t, map[any]any{extraKey{}: "value"}, ctx.Extra)
							assert.Equal(t, map[string]any{"property": "value"}, ctx.Data)
							assert.Equal(t, "value", ctx.Value)
							assert.Equal(t, map[string]any{"property": "value"}, ctx.Parent)
							assert.Equal(t, "property", ctx.Name)
							assert.NotNil(t, ctx.Field) // Content of the field is tested by RuleSet
							assert.False(t, ctx.Now.IsZero())
							assert.False(t, ctx.Invalid)
							assert.Equal(t, walk.MustParse("property"), ctx.Path())
							return true
						},
					}}},
				},
			},
		},
		{
			desc: "now_option_set",
			options: &Options{
				Now: lo.Must(time.Parse(time.RFC3339, "2023-06-28T00:00:00Z")),
				Rules: RuleSet{
					{Path: "property", Rules: List{&testValidator{
						validateFunc: func(_ component, ctx *Context) bool {
							assert.Equal(t, lo.Must(time.Parse(time.RFC3339, "2023-06-28T00:00:00Z")), ctx.Now)
							return true
						},
					}}},
				},
			},
		},
		{
			desc: "absent_parent_not_found",
			options: &Options{
				Data:     map[string]any{"property": "value"},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: "property", Rules: List{Required()}},
					{Path: "object", Rules: List{Required()}},
					{Path: "object.property", Rules: List{Required()}},
				},
			},
			wantValidationErrors: &Errors{
				Fields: FieldsErrors{
					"object": &Errors{
						Errors: []string{"The object is required."},
					},
					// object.property is skipped because parent "object" not found
				},
			},
			wantErrors: nil,
		},
		{
			desc: "absent_not_required",
			options: &Options{
				Data:     map[string]any{"property": "value"},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: "property", Rules: List{Required()}},
					{Path: "object", Rules: List{Object()}},
					{Path: "object.property", Rules: List{Required()}},
				},
			},
		},
		{
			desc: "nil_delete_from_parent",
			options: &Options{
				Data:     map[string]any{"property": nil},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: "property", Rules: List{Required()}},
				},
			},
			wantValidationErrors: &Errors{
				Fields: FieldsErrors{
					"property": &Errors{
						Errors: []string{"The property is required."},
					},
				},
			},
			wantData: map[string]any{},
		},
		{
			desc: "nil_nullable_dont_delete_from_parent",
			options: &Options{
				Data:     map[string]any{"property": nil},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: "property", Rules: List{Required(), Nullable()}},
				},
			},
			wantData: map[string]any{"property": nil}},
		{
			desc: "root_array",
			options: &Options{
				Data:     []any{"a", "b", "c"},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: CurrentElement, Rules: List{Required(), Array()}},
					{Path: "[]", Rules: List{String()}},
				},
			},
			wantData: []string{"a", "b", "c"},
		},
		{
			desc: "root_object_array",
			options: &Options{
				Data:     []any{map[string]any{"value": "a"}, map[string]any{"value": "b"}},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: CurrentElement, Rules: List{Required(), Array()}},
					{Path: "[]", Rules: List{Object()}},
					{Path: "[].value", Rules: List{Required(), String()}},
				},
			},
			wantData: []map[string]any{{"value": "a"}, {"value": "b"}},
		},
		{
			desc: "root_object_array_with_error",
			options: &Options{
				Data:     []any{map[string]any{"value": "a"}, map[string]any{"value": "b"}, "c"},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: CurrentElement, Rules: List{Required(), Array()}},
					{Path: "[]", Rules: List{Object()}},
					{Path: "[].value", Rules: List{Required(), String()}},
				},
			},
			wantValidationErrors: &Errors{
				Elements: ArrayErrors{
					2: &Errors{
						Errors: []string{"The body elements must be objects."},
					},
				},
			},
		},
		{
			desc: "root_object_array_with_deep_error",
			options: &Options{
				Data:     []any{map[string]any{"value": "a"}},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: CurrentElement, Rules: List{Required(), Array()}},
					{Path: "[]", Rules: List{Object()}},
					{Path: "[].value", Rules: List{Required(), Object()}},
				},
			},
			wantValidationErrors: &Errors{
				Elements: ArrayErrors{
					0: &Errors{
						Fields: FieldsErrors{
							"value": &Errors{
								Errors: []string{"The value must be an object."},
							},
						},
					},
				},
			},
		},
		{
			desc: "root_n_array",
			options: &Options{
				Data:     [][]any{{"a", "b"}, {"c", ""}},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: CurrentElement, Rules: List{Required(), Array()}},
					{Path: "[]", Rules: List{Array()}},
					{Path: "[][]", Rules: List{String()}},
				},
			},
			wantData: [][]string{{"a", "b"}, {"c", ""}},
		},
		{
			desc: "root_string",
			options: &Options{
				Data:     "foobar",
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: CurrentElement, Rules: List{Required(), String()}},
				},
			},
			wantData: "foobar",
		},
		{
			desc: "root_number",
			options: &Options{
				Data:     "123",
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: CurrentElement, Rules: List{Required(), Int()}},
				},
			},
			wantData: 123,
		},
		{
			desc: "composition_context_data", // We expect the ctx.Data to be the data relative to the composed RuleSet (prefixDepth)
			options: &Options{
				Data:     map[string]any{"object": map[string]any{"property": "value"}},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: "object", Rules: RuleSet{
						{Path: CurrentElement, Rules: List{Required(), Object()}},
						{Path: "property", Rules: List{Required(), String(), &testValidator{
							validateFunc: func(_ component, ctx *Context) bool {
								assert.Equal(t, map[string]any{"property": "value"}, ctx.Data)
								return true
							},
						}}},
					}},
				},
			},
		},
		{
			desc: "composition_context_data_array",
			options: &Options{
				Data:     map[string]any{"composedArray": [][]string{{"a"}}, "array": [][]string{{"b"}}},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: "composedArray", Rules: RuleSet{
						{Path: CurrentElement, Rules: List{Required(), Array(), &testValidator{
							validateFunc: func(_ component, ctx *Context) bool {
								assert.Equal(t, [][]string{{"a"}}, ctx.Data)
								return true
							},
						}}},
						{Path: "[]", Rules: RuleSet{
							{Path: CurrentElement, Rules: List{Required(), Array(), &testValidator{
								validateFunc: func(_ component, ctx *Context) bool {
									assert.Equal(t, []string{"a"}, ctx.Data)
									return true
								},
							}}},
							{Path: "[]", Rules: RuleSet{
								{Path: CurrentElement, Rules: List{Required(), String(), &testValidator{
									validateFunc: func(_ component, ctx *Context) bool {
										assert.Equal(t, "a", ctx.Data)
										return true
									},
								}}},
							}},
						}},
					}},
					{Path: "array", Rules: List{Required(), Array()}},
					{Path: "array[]", Rules: List{Required(), Array()}},
					{Path: "array[][]", Rules: List{Required(), String(), &testValidator{
						validateFunc: func(_ component, ctx *Context) bool {
							assert.Equal(t, map[string]any{"composedArray": [][]string{{"a"}}, "array": [][]string{{"b"}}}, ctx.Data)
							return true
						},
					}}},
				},
			},
			wantData: map[string]any{"composedArray": [][]string{{"a"}}, "array": [][]string{{"b"}}},
		},
		{
			desc: "non-nullable_nil_array_element",
			options: &Options{
				Data:     map[string]any{"array": []any{"a", nil, "b"}},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: "array", Rules: List{Required(), Array()}},
					{Path: "array[]", Rules: List{Required()}},
				},
			},
			wantValidationErrors: &Errors{
				Fields: FieldsErrors{
					"array": &Errors{
						Elements: ArrayErrors{
							1: &Errors{Errors: []string{"The array elements are required."}},
						},
					},
				},
			},
		},
		{
			desc: "nil_array_element",
			options: &Options{
				Data:     map[string]any{"array": []any{"a", nil, "b"}, "nullableArray": []any{"a", nil, "b"}},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: "array", Rules: List{Required(), Array()}},
					{Path: "array[]", Rules: List{Required()}},
					{Path: "nullableArray", Rules: List{Required(), Array()}},
					{Path: "nullableArray[]", Rules: List{Nullable()}},
				},
			},
			wantValidationErrors: &Errors{
				Fields: FieldsErrors{
					"array": &Errors{
						Elements: ArrayErrors{
							1: &Errors{Errors: []string{"The array elements are required."}},
						},
					},
				},
			},
		},
		{
			desc: "single_value_array_conversion",
			options: &Options{
				Data:                     map[string]any{"singleValueArray": "a", "array": []string{"b", "c"}},
				ConvertSingleValueArrays: true,
				Rules: RuleSet{
					{Path: "array", Rules: List{Required(), Array()}},
					{Path: "array[]", Rules: List{String()}},
					{Path: "singleValueArray", Rules: List{Required(), Array()}},
					{Path: "singleValueArray[]", Rules: List{String()}},
				},
			},
			wantData: map[string]any{"singleValueArray": []string{"a"}, "array": []string{"b", "c"}},
		},
		{
			desc: "errors",
			options: &Options{
				Data: map[string]any{"property": "a"},
				Rules: RuleSet{
					{Path: "property", Rules: List{Required(), &testValidator{
						validateFunc: func(_ component, ctx *Context) bool {
							ctx.AddError(fmt.Errorf("test error 1"), fmt.Errorf("test error 2"))
							return true
						},
					}}},
				},
			},
			wantErrors: []error{fmt.Errorf("test error 1"), fmt.Errorf("test error 2")},
		},
		{
			desc: "validation_errors",
			options: &Options{
				Data:     map[string]any{"property": "a", "object": map[string]any{"property": "c"}, "array": []any{"d"}, "narray": []any{[]any{1, "e", 3}}, "number": 0},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: "property", Rules: List{Required(), Int()}},
					{Path: "number", Rules: List{Required(), Int(), Between(1, 4)}},
					{Path: "missing", Rules: List{Required(), String()}},
					{Path: "object", Rules: List{Required(), Object()}},
					{Path: "object.property", Rules: List{Required(), Int()}},
					{Path: "array", Rules: List{Required(), Array()}},
					{Path: "array[]", Rules: List{Int()}},
					{Path: "narray", Rules: List{Required(), Array()}},
					{Path: "narray[]", Rules: List{Required(), Array()}},
					{Path: "narray[][]", Rules: List{Int()}},
				},
			},
			wantValidationErrors: &Errors{
				Fields: FieldsErrors{
					"property": &Errors{Errors: []string{"The property must be an integer."}},
					"number":   &Errors{Errors: []string{"The number must be between 1 and 4."}},
					"missing":  &Errors{Errors: []string{"The missing is required.", "The missing must be a string."}},
					"object": &Errors{
						Fields: FieldsErrors{
							"property": &Errors{Errors: []string{"The property must be an integer."}},
						},
					},
					"array": &Errors{
						Elements: ArrayErrors{
							0: &Errors{Errors: []string{"The array elements must be integers."}},
						},
					},
					"narray": &Errors{
						Elements: ArrayErrors{
							0: &Errors{
								Elements: ArrayErrors{
									1: &Errors{Errors: []string{"The narray elements must be integers."}},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "added_errors",
			options: &Options{
				Data:     map[string]any{},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: CurrentElement, Rules: List{&addErrorValidator{
						addedValidationErrors: []AddedValidationError[string]{
							{Path: walk.MustParse("property"), Error: "added error"},
							{Path: walk.MustParse("object.addedProp"), Error: "added error"},
							{Path: &walk.Path{Type: walk.PathTypeArray, Name: lo.ToPtr("array"), Index: lo.ToPtr(3), Next: &walk.Path{Type: walk.PathTypeElement}}, Error: "added error"},
							{Path: &walk.Path{Type: walk.PathTypeArray, Name: lo.ToPtr("narray"), Index: lo.ToPtr(0), Next: &walk.Path{Type: walk.PathTypeArray, Index: lo.ToPtr(3), Next: &walk.Path{Type: walk.PathTypeElement}}}, Error: "added error"},
						},
					}}},
				},
			},
			wantValidationErrors: &Errors{
				Fields: FieldsErrors{
					"property": &Errors{Errors: []string{"added error"}},
					"object": &Errors{
						Fields: FieldsErrors{
							"addedProp": &Errors{Errors: []string{"added error"}},
						},
					},
					"array": &Errors{
						Elements: ArrayErrors{
							3: &Errors{Errors: []string{"added error"}},
						},
					},
					"narray": &Errors{
						Elements: ArrayErrors{
							0: &Errors{
								Elements: ArrayErrors{
									3: &Errors{Errors: []string{"added error"}},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "merge_errors",
			options: &Options{
				Data:     map[string]any{"property": "a", "object": map[string]any{"property": "c"}, "array": []any{"d"}, "narray": []any{[]any{1, "e", 3}}, "number": 0},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: "property", Rules: List{Required(), Int()}},
					{Path: "number", Rules: List{Required(), Int(), Between(1, 4)}},
					{Path: "missing", Rules: List{Required(), String()}},
					{Path: "object", Rules: List{Required(), Object()}},
					{Path: "object.property", Rules: List{Required(), Int()}},
					{Path: "array", Rules: List{Required(), Array()}},
					{Path: "array[]", Rules: List{Int()}},
					{Path: "narray", Rules: List{Required(), Array()}},
					{Path: "narray[]", Rules: List{Required(), Array()}},
					{Path: "narray[][]", Rules: List{Int()}},
					{Path: CurrentElement, Rules: List{&addErrorsValidator{
						addedValidationErrors: []AddedValidationError[*Errors]{
							{Path: walk.MustParse("object"), Error: &Errors{
								Fields: FieldsErrors{
									"mergeProp": &Errors{
										Errors: []string{"merge err"},
									},
								},
								Elements: ArrayErrors{
									5: &Errors{
										Errors: []string{"merge err"},
									},
								},
								Errors: []string{"merge err"},
							}},
							{
								Path: &walk.Path{
									Type:  walk.PathTypeArray,
									Name:  lo.ToPtr("array"),
									Index: lo.ToPtr(0),
									Next:  &walk.Path{Type: walk.PathTypeElement},
								},
								Error: &Errors{
									Errors: []string{"merge err"},
								},
							},
							{
								Path: &walk.Path{
									Type:  walk.PathTypeArray,
									Name:  lo.ToPtr("narray"),
									Index: lo.ToPtr(0),
									Next: &walk.Path{
										Type:  walk.PathTypeArray,
										Index: lo.ToPtr(2),
										Next:  &walk.Path{Type: walk.PathTypeElement},
									},
								},
								Error: &Errors{
									Errors: []string{"merge err"},
								},
							},
						},
					}}},
				},
			},
			wantValidationErrors: &Errors{
				Fields: FieldsErrors{
					"property": &Errors{Errors: []string{"The property must be an integer."}},
					"number":   &Errors{Errors: []string{"The number must be between 1 and 4."}},
					"missing":  &Errors{Errors: []string{"The missing is required.", "The missing must be a string."}},
					"object": &Errors{
						Fields: FieldsErrors{
							"property":  &Errors{Errors: []string{"The property must be an integer."}},
							"mergeProp": &Errors{Errors: []string{"merge err"}},
						},
						Elements: ArrayErrors{
							5: &Errors{Errors: []string{"merge err"}},
						},
						Errors: []string{"merge err"},
					},
					"array": &Errors{
						Elements: ArrayErrors{
							0: &Errors{Errors: []string{"The array elements must be integers.", "merge err"}},
						},
					},
					"narray": &Errors{
						Elements: ArrayErrors{
							0: &Errors{
								Elements: ArrayErrors{
									1: &Errors{Errors: []string{"The narray elements must be integers."}},
									2: &Errors{Errors: []string{"merge err"}},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "array_elements_validation_errors",
			options: &Options{
				Data:     map[string]any{"array": []any{"d", "e", "f"}},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: "array", Rules: List{Required(), Array(), &testValidator{
						validateFunc: func(_ component, ctx *Context) bool {
							ctx.AddArrayElementValidationErrors(2, 3)
							return true
						},
					}}},
					{Path: "array[]", Rules: List{String()}},
				},
			},
			wantValidationErrors: &Errors{
				Fields: FieldsErrors{
					"array": &Errors{
						Elements: ArrayErrors{
							2: &Errors{Errors: []string{"validation.rules.test_validator.element"}},
							3: &Errors{Errors: []string{"validation.rules.test_validator.element"}},
						},
					},
				},
			},
		},
		{
			desc: "root_array_elements_validation_errors",
			options: &Options{
				Data:     []any{"d", "e", "f"},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: CurrentElement, Rules: List{Required(), Array(), &testValidator{
						validateFunc: func(_ component, ctx *Context) bool {
							ctx.AddArrayElementValidationErrors(2, 3)
							return true
						},
					}}},
					{Path: "[]", Rules: List{String()}},
				},
			},
			wantValidationErrors: &Errors{
				Elements: ArrayErrors{
					2: &Errors{Errors: []string{"validation.rules.test_validator.element"}},
					3: &Errors{Errors: []string{"validation.rules.test_validator.element"}},
				},
			},
		},
		{
			desc: "type_conversion",
			options: &Options{
				Data:     map[string]any{"property": "123", "object": map[string]any{"property": "456"}, "array": []any{"7"}, "narray": []any{[]any{1, "8", 3}}},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: "property", Rules: List{Required(), Int()}},
					{Path: "object", Rules: List{Required(), Object()}},
					{Path: "object.property", Rules: List{Required(), Int()}},
					{Path: "array", Rules: List{Required(), Array()}},
					{Path: "array[]", Rules: List{Int()}},
					{Path: "narray", Rules: List{Required(), Array()}},
					{Path: "narray[]", Rules: List{Required(), Array()}},
					{Path: "narray[][]", Rules: List{Int()}},
				},
			},
			wantData: map[string]any{"property": 123, "object": map[string]any{"property": 456}, "array": []int{7}, "narray": [][]int{{1, 8, 3}}},
		},
		{
			desc: "empty_array",
			options: &Options{
				Data:     map[string]any{"narray": []any{}},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: "narray", Rules: List{Required(), Array()}},
					{Path: "narray[]", Rules: List{Array()}},
					{Path: "narray[][]", Rules: List{Int()}},
				},
			},
			wantData: map[string]any{"narray": []any{}},
		},
		{
			desc: "empty_narray",
			options: &Options{
				Data:     map[string]any{"narray": []any{[]any{}}},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: "narray", Rules: List{Required(), Array()}},
					{Path: "narray[]", Rules: List{Array()}},
					{Path: "narray[][]", Rules: List{Int()}},
				},
			},
			wantData: map[string]any{"narray": [][]any{{}}},
		},
		{
			desc: "nil_array",
			options: &Options{
				Data:     map[string]any{"narray": nil},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: "narray", Rules: List{Required(), Nullable(), Array()}},
					{Path: "narray[]", Rules: List{Required(), Array()}},
					{Path: "narray[][]", Rules: List{Int()}},
				},
			},
			wantData: map[string]any{"narray": nil},
		},
		{
			desc: "type-dependent",
			options: &Options{
				Data:     map[string]any{"guessString": "string", "guessNumeric": 1, "guessArray": []string{}},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: "string", Rules: List{Required(), String(), &testValidator{
						isTypeDependent: true,
						validateFunc: func(_ component, _ *Context) bool {
							return false
						},
					}}},
					{Path: "integer", Rules: List{Required(), Int(), &testValidator{
						isTypeDependent: true,
						validateFunc: func(_ component, _ *Context) bool {
							return false
						},
					}}},
					{Path: "float", Rules: List{Required(), Float64(), &testValidator{
						isTypeDependent: true,
						validateFunc: func(_ component, _ *Context) bool {
							return false
						},
					}}},
					{Path: "array", Rules: List{Required(), Array(), &testValidator{
						isTypeDependent: true,
						validateFunc: func(_ component, _ *Context) bool {
							return false
						},
					}}},
					{Path: "guessString", Rules: List{Required(), &testValidator{
						isTypeDependent: true,
						validateFunc: func(_ component, _ *Context) bool {
							return false
						},
					}}},
					{Path: "guessNumeric", Rules: List{Required(), &testValidator{
						isTypeDependent: true,
						validateFunc: func(_ component, _ *Context) bool {
							return false
						},
					}}},
					{Path: "guessArray", Rules: List{Required(), &testValidator{
						isTypeDependent: true,
						validateFunc: func(_ component, _ *Context) bool {
							return false
						},
					}}},
				},
			},
			wantValidationErrors: &Errors{
				Fields: FieldsErrors{
					"string":       &Errors{Errors: []string{"The string is required.", "The string must be a string.", "validation.rules.test_validator.string"}},
					"integer":      &Errors{Errors: []string{"The integer is required.", "The integer must be an integer.", "validation.rules.test_validator.numeric"}},
					"float":        &Errors{Errors: []string{"The float is required.", "The float must be numeric.", "validation.rules.test_validator.numeric"}},
					"array":        &Errors{Errors: []string{"The array is required.", "The array must be an array.", "validation.rules.test_validator.array"}},
					"guessString":  &Errors{Errors: []string{"validation.rules.test_validator.string"}},
					"guessNumeric": &Errors{Errors: []string{"validation.rules.test_validator.numeric"}},
					"guessArray":   &Errors{Errors: []string{"validation.rules.test_validator.array"}},
				},
			},
		},
		{
			desc: "expect_array_got_map",
			options: &Options{
				Data:     map[string]any{"array": map[string]any{"key": "value"}},
				Language: lang.New().GetDefault(),
				Rules: RuleSet{
					{Path: "array", Rules: List{Required(), Array()}},
					{Path: "array[]", Rules: List{Int()}},
				},
			},
			wantValidationErrors: &Errors{
				Fields: FieldsErrors{
					"array": &Errors{
						Errors: []string{"The array must be an array."},
					},
				},
			},
		},
		{
			desc: "no_language_provided_in_options",
			options: &Options{
				Data: map[string]any{"property": nil},
				Rules: RuleSet{
					{Path: "property", Rules: List{Required()}},
				},
			},
			wantValidationErrors: &Errors{
				Fields: FieldsErrors{
					"property": &Errors{
						Errors: []string{"The property is required."}, // Translated message (using lang.Default)
					},
				},
			},
			wantData: map[string]any{},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			validationErrors, errs := Validate(c.options)
			assert.Equal(t, c.wantValidationErrors, validationErrors)
			assert.Len(t, errs, len(c.wantErrors))
			for i, e := range errs {
				err, ok := e.(*errors.Error)
				if assert.True(t, ok) {
					assert.Equal(t, []error{c.wantErrors[i]}, err.Unwrap())
				}
			}

			if c.wantData != nil {
				assert.Equal(t, c.wantData, c.options.Data)
			}
		})
	}
}

func TestAddedErrors(t *testing.T) {
	ctx := &Context{}
	expectAddedErrors := []AddedValidationError[string]{
		{Path: walk.MustParse("field1"), Error: "error"},
		{Path: walk.MustParse("field2"), Error: "error"},
	}
	expectArrayElementErrors := []int{4, 6}
	expectMergeErrors := []AddedValidationError[*Errors]{
		{Path: walk.MustParse("field1"), Error: &Errors{
			Fields: FieldsErrors{"field1": &Errors{Errors: []string{"error"}}},
		}},
		{Path: walk.MustParse("field2"), Error: &Errors{Errors: []string{"error"}}},
	}

	ctx.addedValidationErrors = expectAddedErrors
	ctx.arrayElementErrors = expectArrayElementErrors
	ctx.mergeErrors = expectMergeErrors

	assert.Equal(t, expectAddedErrors, ctx.AddedValidationError())
	assert.Equal(t, expectArrayElementErrors, ctx.ArrayElementErrors())
	assert.Equal(t, expectMergeErrors, ctx.AddedValidationErrors())
}

type addErrorValidator struct {
	BaseValidator
	addedValidationErrors []AddedValidationError[string]
}

func (addErrorValidator) Name() string {
	return "addErrorValidator"
}

func (v addErrorValidator) Validate(ctx *Context) bool {
	for _, e := range v.addedValidationErrors {
		ctx.AddValidationError(e.Path, e.Error)
	}
	return true
}

type addErrorsValidator struct {
	BaseValidator
	addedValidationErrors []AddedValidationError[*Errors]
}

func (addErrorsValidator) Name() string {
	return "addErrorsValidator"
}

func (v addErrorsValidator) Validate(ctx *Context) bool {
	for _, e := range v.addedValidationErrors {
		ctx.AddValidationErrors(e.Path, e.Error)
	}
	return true
}
