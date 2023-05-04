package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v4/util/typeutil"
	"goyave.dev/goyave/v4/util/walk"
)

func TestErrors(t *testing.T) {

	t.Run("Error_on_root_element", func(t *testing.T) {
		errs := &ErrorsV5{}

		p := walk.MustParse("root")
		errs.Add(p, "message")

		expected := &ErrorsV5{
			Errors: []string{"message"},
		}
		assert.Equal(t, expected, errs)

		p = walk.MustParse("root")
		errs.Add(p, "second message")

		expected = &ErrorsV5{
			Errors: []string{"message", "second message"},
		}
		assert.Equal(t, expected, errs)
	})

	t.Run("Create_or_append_object", func(t *testing.T) {
		errs := &ErrorsV5{}

		p := walk.MustParse("root.object.property")
		errs.Add(p, "message")

		expected := &ErrorsV5{
			Fields: FieldsErrors{
				"object": &ErrorsV5{
					Fields: FieldsErrors{
						"property": &ErrorsV5{
							Errors: []string{"message"},
						},
					},
				},
			},
		}
		assert.Equal(t, expected, errs)

		p = walk.MustParse("root.object.secondProperty")
		errs.Add(p, "second message")

		expected = &ErrorsV5{
			Fields: FieldsErrors{
				"object": &ErrorsV5{
					Fields: FieldsErrors{
						"property": &ErrorsV5{
							Errors: []string{"message"},
						},
						"secondProperty": &ErrorsV5{
							Errors: []string{"second message"},
						},
					},
				},
			},
		}
		assert.Equal(t, expected, errs)
	})

	t.Run("Create_or_append_array", func(t *testing.T) {
		errs := &ErrorsV5{}

		p := walk.MustParse("root.array[]")
		p.Next.Index = typeutil.Ptr(3)
		errs.Add(p, "message")

		expected := &ErrorsV5{
			Fields: FieldsErrors{
				"array": &ErrorsV5{
					Elements: ArrayErrorsV5{
						3: &ErrorsV5{
							Errors: []string{"message"},
						},
					},
				},
			},
		}
		assert.Equal(t, expected, errs)

		errs.Add(p, "second message")

		expected = &ErrorsV5{
			Fields: FieldsErrors{
				"array": &ErrorsV5{
					Elements: ArrayErrorsV5{
						3: &ErrorsV5{
							Errors: []string{"message", "second message"},
						},
					},
				},
			},
		}
		assert.Equal(t, expected, errs)

		p = walk.MustParse("root.array[][]")
		p.Next.Index = typeutil.Ptr(4)
		p.Next.Next.Index = typeutil.Ptr(5)
		errs.Add(p, "third message")

		expected = &ErrorsV5{
			Fields: FieldsErrors{
				"array": &ErrorsV5{
					Elements: ArrayErrorsV5{
						3: &ErrorsV5{
							Errors: []string{"message", "second message"},
						},
						4: &ErrorsV5{
							Elements: ArrayErrorsV5{
								5: &ErrorsV5{
									Errors: []string{"third message"},
								},
							},
						},
					},
				},
			},
		}
		assert.Equal(t, expected, errs)
	})

	t.Run("Create_or_append_root_array", func(t *testing.T) {
		errs := &ErrorsV5{}

		p := walk.MustParse("root[]")
		p.Index = typeutil.Ptr(3)
		errs.Add(p, "message")
		errs.Add(p, "second message")

		expected := &ErrorsV5{
			Elements: ArrayErrorsV5{
				3: &ErrorsV5{
					Errors: []string{"message", "second message"},
				},
			},
		}
		assert.Equal(t, expected, errs)
	})

}
