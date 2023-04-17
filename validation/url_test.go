package validation

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v4/util/typeutil"
)

func TestURLValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := URL()
		assert.NotNil(t, v)
		assert.Equal(t, "url", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&ContextV5{}))
	})

	cases := []struct {
		value     any
		wantValue *url.URL
		want      bool
	}{
		{value: "http://www.google.com", want: true, wantValue: typeutil.Must(url.ParseRequestURI("http://www.google.com"))},
		{value: typeutil.Must(url.ParseRequestURI("http://www.google.com")), want: true, wantValue: typeutil.Must(url.ParseRequestURI("http://www.google.com"))},
		{value: "https://www.google.com", want: true, wantValue: typeutil.Must(url.ParseRequestURI("https://www.google.com"))},
		{value: "https://www.google.com?q=a%20surprise%20to%20be%20sure", want: true, wantValue: typeutil.Must(url.ParseRequestURI("https://www.google.com?q=a%20surprise%20to%20be%20sure"))},
		{value: "https://www.google.com/#anchor", want: true, wantValue: typeutil.Must(url.ParseRequestURI("https://www.google.com/#anchor"))},
		{value: "https://www.google.com?q=hmm#anchor", want: true, wantValue: typeutil.Must(url.ParseRequestURI("https://www.google.com?q=hmm#anchor"))},
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
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := URL()
			ctx := &ContextV5{
				Value: c.value,
			}
			ok := v.Validate(ctx)
			if assert.Equal(t, c.want, ok) && ok {
				assert.Equal(t, c.wantValue, ctx.Value)
			}
		})
	}
}
