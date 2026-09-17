package config

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5/util/fsutil/osfs"

	_ "embed"
)

//go:embed custom.json
var embedCustomJSON []byte

func TestSource(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		src := Default()
		fileSrc, ok := src.(*fileSource)
		require.True(t, ok)

		want := &fileSource{
			fs:       &osfs.FS{},
			fileName: "config.json",
		}
		assert.Equal(t, want.fs, fileSrc.fs)
		assert.Equal(t, want.fileName, fileSrc.fileName)
		assert.NotNil(t, fileSrc.fn)
		// Read test done in FromFile
	})

	t.Run("getConfigFilePath", func(t *testing.T) {
		cases := []struct {
			env  string
			want string
		}{
			{env: "", want: "config.json"},
			{env: "local", want: "config.json"},
			{env: "LOCAL", want: "config.json"},
			{env: "localhost", want: "config.json"},
			{env: "LocalHost", want: "config.json"},
			{env: "testing", want: "config.testing.json"},
			{env: "prod", want: "config.prod.json"},
			{env: "Prod", want: "config.prod.json"},
		}

		for _, c := range cases {
			t.Run(c.env, func(t *testing.T) {
				assert.Equal(t, c.want, getConfigFilePath(c.env))
			})
		}
	})

	t.Run("FromFile", func(t *testing.T) {
		fs := &osfs.FS{}
		src := FromFile(fs, "custom.json", UnsmarshalReadJSON())
		fileSrc, ok := src.(*fileSource)
		require.True(t, ok)

		want := &fileSource{
			fs:       fs,
			fileName: "custom.json",
		}
		assert.Equal(t, want.fs, fileSrc.fs)
		assert.Equal(t, want.fileName, fileSrc.fileName)
		assert.NotNil(t, fileSrc.fn)

		result, err := src.Read()
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"App": map[string]any{"Name": "test"}}, result)

		t.Run("ENOENT", func(t *testing.T) {
			fs := &osfs.FS{}
			src := FromFile(fs, "notafile.json", UnsmarshalReadJSON())
			result, err := src.Read()
			assert.Nil(t, result)
			assert.ErrorContains(t, err, "no such file")
		})
	})

	t.Run("FromReader", func(t *testing.T) {
		buf := bytes.NewBuffer(embedCustomJSON)

		src := FromReader(buf, UnsmarshalReadJSON())

		readerSrc, ok := src.(*readerSource)
		require.True(t, ok)

		assert.Equal(t, buf, readerSrc.r)
		assert.NotNil(t, readerSrc.fn)

		result, err := src.Read()
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"App": map[string]any{"Name": "test"}}, result)

		t.Run("read_file", func(t *testing.T) {
			f, err := os.Open("custom.json")
			require.NoError(t, err)
			t.Cleanup(func() { _ = f.Close() })
			src := FromReader(f, UnsmarshalReadJSON())

			result, err := src.Read()
			require.NoError(t, err)
			assert.Equal(t, map[string]any{"App": map[string]any{"Name": "test"}}, result)
		})
	})

	t.Run("FromBytes", func(t *testing.T) {
		src := FromBytes(embedCustomJSON, UnsmarshalJSON())

		bytesSrc, ok := src.(*bytesSource)
		require.True(t, ok)

		assert.Equal(t, embedCustomJSON, bytesSrc.b)
		assert.NotNil(t, bytesSrc.fn)

		result, err := src.Read()
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"App": map[string]any{"Name": "test"}}, result)
	})
}

func TestUnmarshalReadAll(t *testing.T) {
	fn := UnsmarshalReadAll(UnsmarshalJSON())
	buf := bytes.NewBufferString(`{"str": "value"}`)

	res := map[string]any{}
	require.NoError(t, fn(buf, &res))
	assert.Equal(t, map[string]any{"str": "value"}, res)
}
