package validation

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUUIDValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := UUID()
		assert.NotNil(t, v)
		assert.Equal(t, "uuid", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.AcceptedVersions)

		v = UUID(4, 5)
		assert.NotNil(t, v)
		assert.Equal(t, "uuid", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []uuid.Version{4, 5}, v.AcceptedVersions)
	})

	cases := []struct {
		value            any
		acceptedVersions []uuid.Version
		wantValue        uuid.UUID
		want             bool
	}{
		{value: "123e4567-e89b-12d3-a456-426655440000", want: true, wantValue: uuid.MustParse("123e4567-e89b-12d3-a456-426655440000")},                                      // v1
		{value: "9125a8dc-52ee-365b-a5aa-81b0b3681cf6", want: true, wantValue: uuid.MustParse("9125a8dc-52ee-365b-a5aa-81b0b3681cf6")},                                      // v3
		{value: "9125a8dc52ee365ba5aa81b0b3681cf6", want: true, wantValue: uuid.MustParse("9125a8dc52ee365ba5aa81b0b3681cf6")},                                              // v3 no hyphen
		{value: "11bf5b37-e0b8-42e0-8dcf-dc8c4aefc000", want: true, wantValue: uuid.MustParse("11bf5b37-e0b8-42e0-8dcf-dc8c4aefc000")},                                      // v4
		{value: "11bf5b37e0b842e08dcfdc8c4aefc000", want: true, wantValue: uuid.MustParse("11bf5b37e0b842e08dcfdc8c4aefc000")},                                              // v4 no hyphen
		{value: "fdda765f-fc57-5604-a269-52a7df8164ec", want: true, wantValue: uuid.MustParse("fdda765f-fc57-5604-a269-52a7df8164ec")},                                      // v5
		{value: "3bbcee75cecc5b568031b6641c1ed1f1", want: true, wantValue: uuid.MustParse("3bbcee75cecc5b568031b6641c1ed1f1")},                                              // v5 no hyphen
		{value: "fdda765f-fc57-5604-a269-52a7df8164ec", want: true, wantValue: uuid.MustParse("fdda765f-fc57-5604-a269-52a7df8164ec"), acceptedVersions: []uuid.Version{5}}, // v5 only
		{value: "9125a8dc-52ee-365b-a5aa-81b0b3681cf6", want: false, acceptedVersions: []uuid.Version{5}},                                                                   // v5 only
		{value: uuid.MustParse("fdda765f-fc57-5604-a269-52a7df8164ec"), want: true, wantValue: uuid.MustParse("fdda765f-fc57-5604-a269-52a7df8164ec"), acceptedVersions: []uuid.Version{5}},
		{value: uuid.MustParse("9125a8dc-52ee-365b-a5aa-81b0b3681cf6"), want: false, wantValue: uuid.MustParse("9125a8dc-52ee-365b-a5aa-81b0b3681cf6"), acceptedVersions: []uuid.Version{5}}, // Already UUID but not v5
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
			v := UUID(c.acceptedVersions...)
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
