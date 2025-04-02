package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func alwaysTrue(_ *Context) bool {
	return true
}

func alwaysFalse(_ *Context) bool {
	return false
}

func TestOnlyIfValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := OnlyIf(alwaysTrue, Min(1))
		assert.NotNil(t, v)
		// Validator implementation should be "inherited"
		// thanks to composition.
		assert.Equal(t, "min", v.Name())
		assert.False(t, v.IsType())
		assert.True(t, v.IsTypeDependent())
		assert.Equal(t, []string{":min", "1"}, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		validator Validator
		condition func(*Context) bool
		desc      string
		want      bool
	}{
		{desc: "condition_true", validator: Max(3), condition: alwaysTrue, value: "string", want: false},
		{desc: "condition_false", validator: Max(3), condition: alwaysFalse, value: "string", want: true}, // true because Max shouldn't be executed.
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%s", c.desc), func(t *testing.T) {
			v := OnlyIf(c.condition, c.validator)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}
