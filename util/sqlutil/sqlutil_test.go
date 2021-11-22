package sqlutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeLike(t *testing.T) {
	assert.Equal(t, "se\\%r\\_h", EscapeLike("se%r_h"))
	assert.Equal(t, "se\\%r\\%\\_h\\_", EscapeLike("se%r%_h_"))
}
