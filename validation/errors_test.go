package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v4/util/typeutil"
	"goyave.dev/goyave/v4/util/walk"
)

func TestErrors(t *testing.T) {

	t.Run("Error_on_root_element", func(t *testing.T) {
		errs := &Errors{}

		p := walk.MustParse("root")
		errs.Add(p, "message")

		expected := &Errors{
			Errors: []string{"message"},
		}
		assert.Equal(t, expected, errs)

		p = walk.MustParse("root")
		errs.Add(p, "second message")

		expected = &Errors{
			Errors: []string{"message", "second message"},
		}
		assert.Equal(t, expected, errs)
	})

	t.Run("Create_or_append_object", func(t *testing.T) {
		errs := &Errors{}

		p := walk.MustParse("root.object.property")
		errs.Add(p, "message")

		expected := &Errors{
			Fields: FieldsErrors{
				"object": &Errors{
					Fields: FieldsErrors{
						"property": &Errors{
							Errors: []string{"message"},
						},
					},
				},
			},
		}
		assert.Equal(t, expected, errs)

		p = walk.MustParse("root.object.secondProperty")
		errs.Add(p, "second message")

		expected = &Errors{
			Fields: FieldsErrors{
				"object": &Errors{
					Fields: FieldsErrors{
						"property": &Errors{
							Errors: []string{"message"},
						},
						"secondProperty": &Errors{
							Errors: []string{"second message"},
						},
					},
				},
			},
		}
		assert.Equal(t, expected, errs)
	})

	t.Run("Create_or_append_array", func(t *testing.T) {
		errs := &Errors{}

		p := walk.MustParse("root.array[]")
		p.Next.Index = typeutil.Ptr(3)
		errs.Add(p, "message")

		expected := &Errors{
			Fields: FieldsErrors{
				"array": &Errors{
					Elements: ArrayErrors{
						3: &Errors{
							Errors: []string{"message"},
						},
					},
				},
			},
		}
		assert.Equal(t, expected, errs)

		errs.Add(p, "second message")

		expected = &Errors{
			Fields: FieldsErrors{
				"array": &Errors{
					Elements: ArrayErrors{
						3: &Errors{
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

		expected = &Errors{
			Fields: FieldsErrors{
				"array": &Errors{
					Elements: ArrayErrors{
						3: &Errors{
							Errors: []string{"message", "second message"},
						},
						4: &Errors{
							Elements: ArrayErrors{
								5: &Errors{
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
		errs := &Errors{}

		p := walk.MustParse("root[]")
		p.Index = typeutil.Ptr(3)
		errs.Add(p, "message")
		errs.Add(p, "second message")

		expected := &Errors{
			Elements: ArrayErrors{
				3: &Errors{
					Errors: []string{"message", "second message"},
				},
			},
		}
		assert.Equal(t, expected, errs)
	})

}
