package goyave

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorUnwrap(t *testing.T) {

	e := Error{
		err: fmt.Errorf("reason"),
	}

	assert.Equal(t, e.err, e.Unwrap())
}
