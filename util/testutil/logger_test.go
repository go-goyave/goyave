package testutil

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockT struct {
	buf *bytes.Buffer
}

func (t mockT) Log(args ...any) {
	t.buf.Write([]byte(fmt.Sprint(args...)))
}

func TestLogWriter(t *testing.T) {

	buf := &bytes.Buffer{}
	writerLogger := &LogWriter{
		t: mockT{buf: buf},
	}

	n, err := writerLogger.Write([]byte("logs"))
	require.NoError(t, err)
	assert.Equal(t, 4, n)

	assert.Equal(t, "logs", buf.String())
}
