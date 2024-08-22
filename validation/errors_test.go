package validation

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/util/walk"
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
		p.Next.Index = lo.ToPtr(3)
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
		p.Next.Index = lo.ToPtr(4)
		p.Next.Next.Index = lo.ToPtr(5)
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
		p.Index = lo.ToPtr(3)
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

	t.Run("Merge", func(t *testing.T) {
		mergeErrs := func() *Errors {
			return &Errors{
				Fields: FieldsErrors{
					"mergeField": &Errors{
						Fields: FieldsErrors{
							"nested": &Errors{Errors: []string{"nested err"}},
						},
						Errors: []string{"merge mergeField err"},
					},
					"field": &Errors{
						Errors: []string{"merge field err"},
					},
				},
				Elements: ArrayErrors{
					1: &Errors{
						Errors: []string{"element err"},
					},
					3: &Errors{
						Fields: FieldsErrors{
							"elementField": &Errors{Errors: []string{"element merge err"}},
						},
					},
				},
				Errors: []string{"merge err 1", "merge err 2"},
			}
		}

		cases := []struct {
			base      *Errors
			mergeErrs *Errors
			path      *walk.Path
			want      *Errors
			desc      string
		}{
			{
				desc: "root",
				base: &Errors{
					Fields: FieldsErrors{
						"field": &Errors{
							Errors: []string{"field err"},
						},
					},
					Elements: ArrayErrors{
						3: &Errors{},
					},
					Errors: []string{"error 1"},
				},
				mergeErrs: mergeErrs(),
				path:      walk.MustParse(""),
				want: &Errors{
					Fields: FieldsErrors{
						"mergeField": &Errors{
							Fields: FieldsErrors{
								"nested": &Errors{Errors: []string{"nested err"}},
							},
							Errors: []string{"merge mergeField err"},
						},
						"field": &Errors{
							Errors: []string{"field err", "merge field err"},
						},
					},
					Elements: ArrayErrors{
						1: &Errors{
							Errors: []string{"element err"},
						},
						3: &Errors{
							Fields: FieldsErrors{
								"elementField": &Errors{Errors: []string{"element merge err"}},
							},
						},
					},
					Errors: []string{"error 1", "merge err 1", "merge err 2"},
				},
			},
			{
				desc: "in_array_element",
				base: &Errors{
					Fields: FieldsErrors{
						"field": &Errors{
							Errors: []string{"field err"},
						},
					},
					Elements: ArrayErrors{
						3: &Errors{},
					},
					Errors: []string{"error 1"},
				},
				mergeErrs: mergeErrs(),
				path: &walk.Path{
					Type:  walk.PathTypeArray,
					Index: lo.ToPtr(3),
					Next:  &walk.Path{Type: walk.PathTypeElement},
				},
				want: &Errors{
					Fields: FieldsErrors{
						"field": &Errors{
							Errors: []string{"field err"},
						},
					},
					Elements: ArrayErrors{
						3: mergeErrs(),
					},
					Errors: []string{"error 1"},
				},
			},
			{
				desc: "in_new_array_element",
				base: &Errors{
					Fields: FieldsErrors{
						"field": &Errors{
							Errors: []string{"field err"},
						},
					},
					Elements: ArrayErrors{
						3: &Errors{},
					},
					Errors: []string{"error 1"},
				},
				mergeErrs: mergeErrs(),
				path: &walk.Path{
					Type:  walk.PathTypeArray,
					Index: lo.ToPtr(2),
					Next:  &walk.Path{Type: walk.PathTypeObject, Next: &walk.Path{Type: walk.PathTypeElement, Name: lo.ToPtr("property")}},
				},
				want: &Errors{
					Fields: FieldsErrors{
						"field": &Errors{
							Errors: []string{"field err"},
						},
					},
					Elements: ArrayErrors{
						3: &Errors{},
						2: &Errors{
							Fields: FieldsErrors{
								"property": mergeErrs(),
							},
						},
					},
					Errors: []string{"error 1"},
				},
			},
			{
				desc: "in_field",
				base: &Errors{
					Fields: FieldsErrors{
						"field": &Errors{
							Errors: []string{"field err"},
						},
					},
					Elements: ArrayErrors{
						3: &Errors{},
					},
					Errors: []string{"error 1"},
				},
				mergeErrs: mergeErrs(),
				path: &walk.Path{
					Type: walk.PathTypeObject,
					Next: &walk.Path{
						Type: walk.PathTypeElement,
						Name: lo.ToPtr("field"),
					},
				},
				want: &Errors{
					Fields: FieldsErrors{
						"field": &Errors{
							Fields: FieldsErrors{
								"mergeField": &Errors{
									Fields: FieldsErrors{
										"nested": &Errors{Errors: []string{"nested err"}},
									},
									Errors: []string{"merge mergeField err"},
								},
								"field": &Errors{
									Errors: []string{"merge field err"},
								},
							},
							Elements: ArrayErrors{
								1: &Errors{
									Errors: []string{"element err"},
								},
								3: &Errors{
									Fields: FieldsErrors{
										"elementField": &Errors{Errors: []string{"element merge err"}},
									},
								},
							},
							Errors: []string{"field err", "merge err 1", "merge err 2"},
						},
					},
					Elements: ArrayErrors{
						3: &Errors{},
					},
					Errors: []string{"error 1"},
				},
			},
			{
				desc: "in_new_field",
				base: &Errors{
					Fields: FieldsErrors{
						"field": &Errors{
							Errors: []string{"field err"},
						},
					},
					Elements: ArrayErrors{
						3: &Errors{},
					},
					Errors: []string{"error 1"},
				},
				mergeErrs: mergeErrs(),
				path: &walk.Path{
					Type: walk.PathTypeObject,
					Next: &walk.Path{
						Type: walk.PathTypeObject,
						Name: lo.ToPtr("mergeObject"),
						Next: &walk.Path{
							Type: walk.PathTypeElement,
							Name: lo.ToPtr("mergeProp"),
						},
					},
				},
				want: &Errors{
					Fields: FieldsErrors{
						"field": &Errors{
							Errors: []string{"field err"},
						},
						"mergeObject": &Errors{
							Fields: FieldsErrors{
								"mergeProp": mergeErrs(),
							},
						},
					},
					Elements: ArrayErrors{
						3: &Errors{},
					},
					Errors: []string{"error 1"},
				},
			},
			{
				desc: "in_new_field_elements",
				base: &Errors{
					Fields: FieldsErrors{
						"field": &Errors{
							Errors: []string{"field err"},
						},
					},
					Elements: ArrayErrors{
						3: &Errors{},
					},
					Errors: []string{"error 1"},
				},
				mergeErrs: mergeErrs(),
				path: &walk.Path{
					Type: walk.PathTypeObject,
					Next: &walk.Path{
						Type:  walk.PathTypeArray,
						Name:  lo.ToPtr("mergeArray"),
						Index: lo.ToPtr(4),
						Next:  &walk.Path{Type: walk.PathTypeElement},
					},
				},
				want: &Errors{
					Fields: FieldsErrors{
						"field": &Errors{
							Errors: []string{"field err"},
						},
						"mergeArray": &Errors{
							Elements: ArrayErrors{
								4: mergeErrs(),
							},
						},
					},
					Elements: ArrayErrors{
						3: &Errors{},
					},
					Errors: []string{"error 1"},
				},
			},
		}

		for _, c := range cases {
			t.Run(c.desc, func(t *testing.T) {
				errs := c.base
				errs.Merge(c.path, c.mergeErrs)
				assert.Equal(t, c.want, errs)
			})
		}
	})
}
