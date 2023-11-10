package validation

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/util/walk"
)

type isRequiredKey struct{}

func TestField(t *testing.T) {

	t.Run("New", func(t *testing.T) {
		validators := []Validator{
			Required(),
			Nullable(),
			Array(),
			Object(),
		}
		f := newField("object.array[].property", validators, 2)
		assert.True(t, f.IsArray())
		assert.True(t, f.IsObject())
		assert.True(t, f.IsNullable())
		assert.True(t, f.isRequired(nil))
		assert.Equal(t, uint(2), f.PrefixDepth())
		assert.Equal(t, walk.MustParse("object.array[].property"), f.Path)
		assert.Equal(t, validators, f.Validators)
	})

	t.Run("New_not_required", func(t *testing.T) {
		validators := []Validator{String()}
		f := newField("object.array[].property", validators, 0)
		assert.False(t, f.IsArray())
		assert.False(t, f.IsObject())
		assert.False(t, f.IsNullable())
		assert.Nil(t, f.isRequired)
		assert.False(t, f.IsRequired(nil))
		assert.Equal(t, uint(0), f.PrefixDepth())
	})

	t.Run("New_required_if", func(t *testing.T) {
		validators := []Validator{
			RequiredIf(func(c *Context) bool {
				return c.Extra[isRequiredKey{}].(bool)
			}),
			String(),
		}
		f := newField("object.array[].property", validators, 0)
		assert.False(t, f.isRequired(&Context{Extra: map[any]any{isRequiredKey{}: false}}))
		assert.True(t, f.isRequired(&Context{Extra: map[any]any{isRequiredKey{}: true}}))
	})

	t.Run("Get_error_path", func(t *testing.T) {
		expected := walk.MustParse("object.array[]")

		f := newField("object.array[]", []Validator{String()}, 0)
		assert.Equal(t, expected, f.getErrorPath(nil, &walk.Context{Path: expected}))

		expected = walk.MustParse("object.array[][]")
		expected.Next.Next.Index = lo.ToPtr(5)
		assert.Equal(t, expected, f.getErrorPath(walk.MustParse("object.array[]"), &walk.Context{Index: 5}))
	})
}
