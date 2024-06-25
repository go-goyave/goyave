package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/util/walk"
)

func requiredIfTestFunction(_ *Context) bool { return true }

func BenchmarkRuleSet(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
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
		{Path: "deep_array_element_composition", Rules: RuleSet{
			{Path: CurrentElement, Rules: List{Array()}},
			{Path: "[]", Rules: RuleSet{
				{Path: "[]", Rules: RuleSet{
					{Path: CurrentElement, Rules: List{Int()}},
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
					Validators: []Validator{Int()},
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
						Validators: []Validator{Int()},
					},
					isArray: true,
				},
				isArray: true,
			},
			isArray: true,
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
			Path:       walk.MustParse("array_element_composition"),
			Validators: []Validator{Array()},
			Elements: &Field{
				Path:        walk.MustParse("[]"),
				Validators:  []Validator{Int()},
				prefixDepth: 2,
			},
			prefixDepth: 1,
			isArray:     true,
		},
		{
			Path:       walk.MustParse("deep_array_element_composition"),
			Validators: []Validator{Array()},
			Elements: &Field{
				Path:       walk.MustParse("[]"), // Injected
				Validators: []Validator{Array()},
				Elements: &Field{
					Path:        walk.MustParse("[]"),
					Validators:  []Validator{Int()},
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

func TestRulesetRequired(t *testing.T) {
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
