package validation

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestURLValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := URL()
		assert.NotNil(t, v)
		assert.Equal(t, "url", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue *url.URL
		want      bool
	}{
		{value: "http://www.google.com", want: true, wantValue: lo.Must(url.ParseRequestURI("http://www.google.com"))},
		{value: lo.Must(url.ParseRequestURI("http://www.google.com")), want: true, wantValue: lo.Must(url.ParseRequestURI("http://www.google.com"))},
		{value: "https://www.google.com", want: true, wantValue: lo.Must(url.ParseRequestURI("https://www.google.com"))},
		{value: "https://www.google.com?q=a%20surprise%20to%20be%20sure", want: true, wantValue: lo.Must(url.ParseRequestURI("https://www.google.com?q=a%20surprise%20to%20be%20sure"))},
		{value: "https://www.google.com/#anchor", want: true, wantValue: lo.Must(url.ParseRequestURI("https://www.google.com/#anchor"))},
		{value: "https://www.google.com?q=hmm#anchor", want: true, wantValue: lo.Must(url.ParseRequestURI("https://www.google.com?q=hmm#anchor"))},
		{value: "https://www.google.com#anchor", want: false},
		{value: "www.google.com", want: false},
		{value: "w-w.google.com", want: false},
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
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := URL()
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
