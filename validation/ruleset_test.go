package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/lang"
	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/util/walk"
)

func requiredIfTestFunction(_ *Context) bool { return true }

func BenchmarkRuleSet(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		ruleset := RuleSet{
			{Path: CurrentElement, Rules: List{
				Object(),
			}},
			{Path: "property", Rules: List{
				String(),
			}},
			{Path: "nullable", Rules: List{
				Nullable(),
			}},
			{Path: "object", Rules: List{
				Object(),
			}},
			{Path: "object.property", Rules: List{
				String(),
			}},
			{Path: "anonymous.property", Rules: List{
				String(),
			}},
			{Path: "composition", Rules: RuleSet{
				{Path: "composed_prop", Rules: List{Int()}},
				{Path: "nested_composition", Rules: RuleSet{
					{Path: CurrentElement, Rules: List{String()}},
				}},
				{Path: "composition_on_current_element", Rules: RuleSet{
					{Path: CurrentElement, Rules: RuleSet{
						{Path: "nested_prop", Rules: List{String()}},
					}},
				}},
			}},
			{Path: "two_dim_array", Rules: List{Array()}},
			{Path: "two_dim_array[]", Rules: List{Array()}},
			{Path: "two_dim_array[][]", Rules: List{Int()}},

			// Parent arrays should be injected
			{Path: "array[]", Rules: List{Int()}},
			{Path: "deep_array[][][]", Rules: List{Int()}},

			{Path: "array_composition", Rules: RuleSet{
				{Path: CurrentElement, Rules: List{Array()}},
				{Path: "[]", Rules: List{Int()}},
			}},

			{Path: "array_element_composition", Rules: RuleSet{
				{Path: CurrentElement, Rules: List{Array()}},
				{Path: "[]", Rules: RuleSet{
					{Path: CurrentElement, Rules: List{Int()}},
				}},
			}},
		}
		ruleset.AsRules()
	}
}

func TestRuleset(t *testing.T) {
	ruleset := RuleSet{
		{Path: CurrentElement, Rules: List{
			Object(),
		}},
		{Path: "property", Rules: List{
			String(),
		}},
		{Path: "nullable", Rules: List{
			Nullable(),
		}},
		{Path: "object", Rules: List{
			Object(),
		}},
		{Path: "object.property", Rules: List{
			String(),
		}},
		{Path: `object.\*`, Rules: List{
			String(),
		}},
		{Path: "anonymous.property", Rules: List{
			String(),
		}},
		{Path: "composition", Rules: RuleSet{
			{Path: "composed_prop", Rules: List{Int()}},
			{Path: "nested_composition", Rules: RuleSet{
				{Path: CurrentElement, Rules: List{String()}},
			}},
			{Path: "composition_on_current_element", Rules: RuleSet{
				{Path: CurrentElement, Rules: RuleSet{
					{Path: "nested_prop", Rules: List{String()}},
				}},
			}},
		}},
		{Path: "two_dim_array", Rules: List{Array()}},
		{Path: "two_dim_array[]", Rules: List{Array()}},
		{Path: "two_dim_array[][]", Rules: List{Int64()}},

		// Parent arrays should be injected
		{Path: "array[]", Rules: List{Int()}},
		{Path: "deep_array[][][]", Rules: List{Int16()}},

		// Should not be injected because bracket is escaped
		{Path: `escaped_array\[]`, Rules: List{Int8()}},
		{Path: `object.array\[]`, Rules: List{Int8()}},

		{Path: "array_composition", Rules: RuleSet{
			{Path: CurrentElement, Rules: List{Array()}},
			{Path: "[]", Rules: List{Int()}},
		}},

		{Path: "array_of_objects_composition", Rules: RuleSet{
			{Path: CurrentElement, Rules: List{Array()}},
			{Path: "[]", Rules: List{Object()}},
			{Path: "[].field", Rules: List{String()}},
			{Path: `[].fi\[\]eld`, Rules: List{String()}},
			{Path: `[].\*`, Rules: List{String()}},
			{Path: `[].object.prop\.with\.a\.dot`, Rules: List{String()}},
		}},

		{Path: "array_element_composition", Rules: RuleSet{
			{Path: CurrentElement, Rules: List{Array()}},
			{Path: "[]", Rules: RuleSet{
				{Path: CurrentElement, Rules: List{Int8()}},
			}},
		}},

		{Path: "array_object_element_composition", Rules: RuleSet{
			{Path: CurrentElement, Rules: List{Array()}},
			{Path: "[]", Rules: RuleSet{
				{Path: CurrentElement, Rules: List{Object()}},
				{Path: "field", Rules: List{String()}},
			}},
		}},

		{Path: "deep_array_element_composition", Rules: RuleSet{
			{Path: CurrentElement, Rules: List{Array()}},
			{Path: "[]", Rules: RuleSet{
				{Path: "[]", Rules: RuleSet{
					{Path: CurrentElement, Rules: List{Int32()}},
				}},
			}},
		}},
	}

	expected := Rules{
		{
			Path:       walk.MustParse(""),
			Validators: []Validator{Object()},
			isObject:   true,
		},
		{
			Path:       walk.MustParse("property"),
			Validators: []Validator{String()},
		},
		{

			Path:       walk.MustParse("nullable"),
			Validators: []Validator{Nullable()},
			isNullable: true,
		},
		{
			Path:       walk.MustParse("object"),
			Validators: []Validator{Object()},
			isObject:   true,
		},
		{
			Path:       walk.MustParse("object.property"),
			Validators: []Validator{String()},
		},
		{
			Path:       walk.MustParse(`object.\*`),
			Validators: []Validator{String()},
		},
		{
			Path:       walk.MustParse("anonymous.property"),
			Validators: []Validator{String()},
		},
		{
			Path:        walk.MustParse("composition.composed_prop"),
			Validators:  []Validator{Int()},
			prefixDepth: 1,
		},
		{
			Path:        walk.MustParse("composition.nested_composition"),
			Validators:  []Validator{String()},
			prefixDepth: 2,
		},
		{
			Path:        walk.MustParse("composition.composition_on_current_element.nested_prop"),
			Validators:  []Validator{String()},
			prefixDepth: 2,
		},
		{
			Path:       walk.MustParse("two_dim_array"),
			Validators: []Validator{Array()},
			Elements: &Field{
				Path:       walk.MustParse("[]"),
				Validators: []Validator{Array()},
				Elements: &Field{
					Path:       walk.MustParse("[]"),
					Validators: []Validator{Int64()},
				},
				isArray: true,
			},
			isArray: true,
		},
		{
			Path:       walk.MustParse("array"), // Injected
			Validators: []Validator{Array()},
			Elements: &Field{
				Path:       walk.MustParse("[]"),
				Validators: []Validator{Int()},
			},
			isArray: true,
		},
		{
			Path:       walk.MustParse("deep_array"), // Injected
			Validators: []Validator{Array()},
			Elements: &Field{
				Path:       walk.MustParse("[]"), // Injected
				Validators: []Validator{Array()},
				Elements: &Field{
					Path:       walk.MustParse("[]"), // Injected
					Validators: []Validator{Array()},
					Elements: &Field{
						Path:       walk.MustParse("[]"),
						Validators: []Validator{Int16()},
					},
					isArray: true,
				},
				isArray: true,
			},
			isArray: true,
		},
		{
			Path:       walk.MustParse(`escaped_array\[]`),
			Validators: []Validator{Int8()},
		},
		{
			Path:       walk.MustParse(`object.array\[]`),
			Validators: []Validator{Int8()},
		},
		{
			Path:       walk.MustParse("array_composition"),
			Validators: []Validator{Array()},
			Elements: &Field{
				Path:        walk.MustParse("[]"),
				Validators:  []Validator{Int()},
				prefixDepth: 1,
			},
			prefixDepth: 1,
			isArray:     true,
		},
		{
			Path:       walk.MustParse("array_of_objects_composition"),
			Validators: []Validator{Array()},
			Elements: &Field{
				Path:        walk.MustParse("[]"),
				Validators:  []Validator{Object()},
				prefixDepth: 1,
				isObject:    true,
			},
			prefixDepth: 1,
			isArray:     true,
		},
		{
			Path:        walk.MustParse("array_of_objects_composition[].field"),
			Validators:  []Validator{String()},
			prefixDepth: 1,
		},
		{
			Path:        walk.MustParse(`array_of_objects_composition[].fi\[\]eld`),
			Validators:  []Validator{String()},
			prefixDepth: 1,
		},
		{
			Path:        walk.MustParse(`array_of_objects_composition[].\*`),
			Validators:  []Validator{String()},
			prefixDepth: 1,
		},
		{
			Path:        walk.MustParse(`array_of_objects_composition[].object.prop\.with\.a\.dot`),
			Validators:  []Validator{String()},
			prefixDepth: 1,
		},
		{
			Path:       walk.MustParse("array_element_composition"),
			Validators: []Validator{Array()},
			Elements: &Field{
				Path:        walk.MustParse("[]"),
				Validators:  []Validator{Int8()},
				prefixDepth: 2,
			},
			prefixDepth: 1,
			isArray:     true,
		},
		{
			Path:       walk.MustParse("array_object_element_composition"),
			Validators: []Validator{Array()},
			Elements: &Field{
				Path:        walk.MustParse("[]"),
				Validators:  []Validator{Object()},
				prefixDepth: 2,
				isObject:    true,
			},
			prefixDepth: 1,
			isArray:     true,
		},
		{
			Path:        walk.MustParse("array_object_element_composition[].field"),
			Validators:  []Validator{String()},
			prefixDepth: 2,
		},
		{
			Path:       walk.MustParse("deep_array_element_composition"),
			Validators: []Validator{Array()},
			Elements: &Field{
				Path:       walk.MustParse("[]"), // Injected
				Validators: []Validator{Array()},
				Elements: &Field{
					Path:        walk.MustParse("[]"),
					Validators:  []Validator{Int32()},
					prefixDepth: 3,
				},
				prefixDepth: 2,
				isArray:     true,
			},
			prefixDepth: 1,
			isArray:     true,
		},
	}

	assert.Equal(t, expected, ruleset.AsRules())
}

func TestRuleSetRequired(t *testing.T) {
	ruleset := RuleSet{
		{Path: "required", Rules: List{
			Required(),
		}},
		{Path: "required_if", Rules: List{
			RequiredIf(requiredIfTestFunction),
		}},
	}

	rules := ruleset.AsRules()

	if !assert.Len(t, rules, 2) {
		return
	}

	for _, r := range rules {
		assert.True(t, r.isRequired(nil))
	}
}

func TestRules(t *testing.T) {
	rules := Rules{{}, {}}
	assert.Equal(t, rules, rules.AsRules())
}

func TestBaseValidator(t *testing.T) {
	v := &BaseValidator{}

	opts := &Options{
		DB:       &gorm.DB{},
		Config:   &config.Config{},
		Language: lang.Default,
		Logger:   &slog.Logger{},
	}
	v.Init(opts)
	assert.Same(t, opts.DB, v.db)
	assert.Same(t, opts.Config, v.config)
	assert.Same(t, opts.Language, v.lang)
	assert.Same(t, opts.Logger, v.logger)

	assert.False(t, v.IsTypeDependent())
	assert.False(t, v.IsType())
	assert.Equal(t, []string{}, v.MessagePlaceholders(nil))
}

// https://github.com/go-goyave/goyave/issues/248
// The order of validation must be preserved when using composition.
func TestRuleSetIssue248(t *testing.T) {
	ruleset := RuleSet{
		{Path: CurrentElement, Rules: List{Object()}},
		{Path: "field1", Rules: List{String()}},
		{Path: "array", Rules: List{Array()}},
		{Path: "array[]", Rules: RuleSet{
			{Path: CurrentElement, Rules: List{Object()}},
			{Path: "elementField", Rules: List{Int64()}},
		}},
	}

	expected := Rules{
		{
			Path:       walk.MustParse(""),
			Validators: []Validator{Object()},
			isObject:   true,
		},
		{
			Path:       walk.MustParse("field1"),
			Validators: []Validator{String()},
		},
		{
			Path:       walk.MustParse("array"),
			Validators: []Validator{Array()},
			Elements: &Field{
				Path:        walk.MustParse("[]"),
				Validators:  []Validator{Object()},
				prefixDepth: 2,
				isObject:    true,
			},
			isArray: true,
		},
		{
			Path:        walk.MustParse("array[].elementField"),
			Validators:  []Validator{Int64()},
			prefixDepth: 2,
		},
	}

	assert.Equal(t, expected, ruleset.AsRules())

	// Let's try again but simply swapping the order of "array[]" and "array"
	// "elementField" should now be validated before "array" and "array[]".
	ruleset = RuleSet{
		{Path: CurrentElement, Rules: List{Object()}},
		{Path: "field1", Rules: List{String()}},
		{Path: "array[]", Rules: RuleSet{
			{Path: CurrentElement, Rules: List{Object()}},
			{Path: "elementField", Rules: List{Int64()}},
		}},
		{Path: "array", Rules: List{Array()}},
	}

	expected = Rules{
		{
			Path:       walk.MustParse(""),
			Validators: []Validator{Object()},
			isObject:   true,
		},
		{
			Path:       walk.MustParse("field1"),
			Validators: []Validator{String()},
		},
		{
			Path:        walk.MustParse("array[].elementField"),
			Validators:  []Validator{Int64()},
			prefixDepth: 2,
		},
		{
			Path:       walk.MustParse("array"),
			Validators: []Validator{Array()},
			Elements: &Field{
				Path:        walk.MustParse("[]"),
				Validators:  []Validator{Object()},
				prefixDepth: 2,
				isObject:    true,
			},
			isArray: true,
		},
	}

	assert.Equal(t, expected, ruleset.AsRules())
}

// https://github.com/go-goyave/goyave/issues/249
// Repeated / duplicate paths are forbidden (should panic)
// Use of both wildcard and object properties is forbidden
func TestRuleSetRepeatedPath(t *testing.T) {
	cases := []struct {
		desc    string
		wantErr string
		ruleset RuleSet
	}{
		{
			desc: "root_element",
			ruleset: RuleSet{
				{Path: CurrentElement, Rules: List{JSON()}},
				{Path: CurrentElement, Rules: List{Object()}},
			},
			wantErr: "validation.RuleSet: duplicate path \"\" in rule set",
		},
		{
			desc: "root_element_wildcard",
			ruleset: RuleSet{
				{Path: "*", Rules: List{JSON()}},
				{Path: "*", Rules: List{Object()}},
			},
			wantErr: "validation.RuleSet: duplicate path \"*\" in rule set",
		},
		{
			desc: "root_element_escaped_wildcard",
			ruleset: RuleSet{
				{Path: "*", Rules: List{JSON()}},
				{Path: `\*`, Rules: List{Object()}},
			},
			wantErr: "validation.RuleSet: cannot validate an object property with both the wildcard (*) and specific property paths (at \"\\*\")",
		},
		{
			desc: "field",
			ruleset: RuleSet{
				{Path: CurrentElement, Rules: List{Object()}},
				{Path: "field", Rules: List{String()}},
				{Path: "field", Rules: List{Min(1)}},
			},
			wantErr: "validation.RuleSet: duplicate path \"field\" in rule set",
		},
		{
			desc: "array",
			ruleset: RuleSet{
				{Path: CurrentElement, Rules: List{Object()}},
				{Path: "array", Rules: List{Array()}},
				{Path: "array[]", Rules: List{Int()}},
				{Path: "array[]", Rules: List{Min(1)}},
			},
			wantErr: "validation.RuleSet: duplicate path \"array[]\" in rule set",
		},
		{
			desc: "escaped_wildcard",
			ruleset: RuleSet{
				{Path: "object.*", Rules: List{JSON()}},
				{Path: `object.\*`, Rules: List{Object()}},
			},
			wantErr: "validation.RuleSet: cannot validate an object property with both the wildcard (*) and specific property paths (at \"object.\\*\")",
		},
		{
			desc: "deep_escaped_wildcard",
			ruleset: RuleSet{
				{Path: "objects", Rules: List{Object()}},
				{Path: "objects.*", Rules: List{Object()}},
				{Path: "objects.*.sub", Rules: List{Object()}},
				{Path: "objects.*.sub.*", Rules: List{JSON()}},
				{Path: `objects.*.sub.\*`, Rules: List{Object()}},
			},
			wantErr: "validation.RuleSet: cannot validate an object property with both the wildcard (*) and specific property paths (at \"objects.*.sub.\\*\")",
		},
		{
			desc: "wildcard_with_other_props",
			ruleset: RuleSet{
				{Path: "object.*", Rules: List{JSON()}},
				{Path: `object.a`, Rules: List{Object()}},
			},
			wantErr: "validation.RuleSet: cannot validate an object property with both the wildcard (*) and specific property paths (at \"object.a\")",
		},
		{
			desc: "escaped_array",
			ruleset: RuleSet{
				{Path: `array\[]`, Rules: List{Array()}},
				{Path: `array\[\]`, Rules: List{Array()}},
			},
			wantErr: "validation.RuleSet: duplicate path \"array\\[\\]\" in rule set",
		},
		{
			desc: "composition",
			ruleset: RuleSet{
				{Path: CurrentElement, Rules: List{Object()}},
				{Path: "object", Rules: RuleSet{
					{Path: CurrentElement, Rules: List{Object()}},
					{Path: "field", Rules: List{String()}},
				}},
				{Path: "object.field", Rules: List{Min(1)}},
			},
			wantErr: "validation.RuleSet: duplicate path \"object.field\" in rule set",
		},
		{
			desc: "composition_current_element",
			ruleset: RuleSet{
				{Path: CurrentElement, Rules: List{Object()}},
				{Path: "object", Rules: RuleSet{
					{Path: CurrentElement, Rules: List{Object()}},
					{Path: "field", Rules: List{String()}},
				}},
				{Path: "object", Rules: List{Min(1)}},
			},
			wantErr: "validation.RuleSet: duplicate path \"object\" in rule set",
		},
		{
			desc: "composition_array",
			ruleset: RuleSet{
				{Path: CurrentElement, Rules: List{Object()}},
				{Path: "array", Rules: RuleSet{
					{Path: CurrentElement, Rules: List{Array()}},
					{Path: "[]", Rules: List{String()}},
				}},
				{Path: "array[]", Rules: List{Min(1)}},
			},
			wantErr: "validation.RuleSet: duplicate path \"array[]\" in rule set",
		},
		{
			desc: "deep_injected_array",
			ruleset: RuleSet{
				{Path: "[][][]", Rules: List{Int()}},
				{Path: "[][][]", Rules: List{Min(1)}},
			},
			wantErr: "validation.RuleSet: duplicate path \"[][][]\" in rule set",
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			assert.PanicsWithError(t, c.wantErr, func() {
				c.ruleset.AsRules()
			})
		})
	}
}
