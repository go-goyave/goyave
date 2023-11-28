package validation

import (
	"fmt"
	"net/mail"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Email()
		assert.NotNil(t, v)
		assert.Equal(t, "email", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
	})

	cases := []struct {
		value     any
		wantValue string
		want      bool
	}{
		{value: "johndoe@example.org", want: true, wantValue: "johndoe@example.org"},
		{value: &mail.Address{Address: "johndoe@example.org"}, want: true, wantValue: "johndoe@example.org"},
		{value: "johndoe+ext@example.org", want: true, wantValue: "johndoe+ext@example.org"},
		{value: "Barry Gibbs <bg@example.com>", want: true, wantValue: "bg@example.com"},
		{value: "+@b.io", want: true, wantValue: "+@b.io"},
		{value: "a@b.io", want: true, wantValue: "a@b.io"},
		{value: "a@b", want: true, wantValue: "a@b"},
		{value: "a@b.c", want: true, wantValue: "a@b.c"},
		{value: "string", want: false},
		{value: "", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Email()
			ctx := &Context{
				Value: c.value,
			}
			ok := v.Validate(ctx)
			if assert.Equal(t, c.want, ok) && ok {
				assert.Equal(t, c.wantValue, ctx.Value)
			}
		})
	}
}
