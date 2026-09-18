package config

import (
	"encoding/json/v2"
	"io"
	"io/fs"
	"os"
	"strings"

	"goyave.dev/goyave/v5/util/errwrap"
	"goyave.dev/goyave/v5/util/fsutil/osfs"
)

// UnmarshalFunc generic unmarshal function compatible with most file formats (json, yml, toml, ...)
type UnmarshalFunc func(in []byte, out any) error

// UnmarshalReadFunc generic unmarshal function taking an `io.Reader` as input instead of raw bytes.
// Useful for data streaming or large configuration files.
type UnmarshalReadFunc func(in io.Reader, out any) error

// UnsmarshalJSON adapter for std json/v2.
func UnsmarshalJSON(opts ...json.Options) UnmarshalFunc {
	return func(in []byte, out any) error {
		return json.Unmarshal(in, out, opts...)
	}
}

// UnsmarshalReadJSON adapter for std json/v2.
func UnsmarshalReadJSON(opts ...json.Options) UnmarshalReadFunc {
	return func(r io.Reader, out any) error {
		return json.UnmarshalRead(r, out, opts...)
	}
}

// UnsmarshalReadAll is an adapter for unmarshalers that don't support reader inputs.
// It drains the reader using `io.ReadAll` before calling the given UnmarshalFunc with the result.
//
//	config.FromReader(reader, config.UnsmarshalReadAll(yaml.Unmarshal))
func UnsmarshalReadAll(fn UnmarshalFunc) UnmarshalReadFunc {
	return func(in io.Reader, out any) error {
		data, err := io.ReadAll(in)
		if err != nil {
			return errwrap.New(err)
		}
		return fn(data, out)
	}
}

// Source for configuration loading.
// Implementations can have different kinds of inputs (raw bytes, reader, filename).
// They are responsible of gathering the data from the inputs and feeding it
// to an unmarshal function (`UnmarshalFunc` or `UnmarshalReadFunc`).
//
// This design provides a unified interface for sourcing and loading configuration
// that can come from anywhere in any format (json, yml, toml, ...).
type Source interface {
	Read() (any, error)
}

type bytesSource struct {
	fn UnmarshalFunc
	b  []byte
}

func (s bytesSource) Read() (any, error) {
	var raw any
	err := s.fn(s.b, &raw)
	return raw, errwrap.New(err)
}

type readerSource struct {
	fn UnmarshalReadFunc
	r  io.Reader
}

func (s readerSource) Read() (any, error) {
	var raw any
	err := s.fn(s.r, &raw)
	return raw, errwrap.New(err)
}

type fileSource struct {
	fn       UnmarshalReadFunc
	fs       fs.FS
	fileName string
}

func (s fileSource) Read() (any, error) {
	var raw any

	reader, err := s.fs.Open(s.fileName)
	if err != nil {
		// We don't want to silently skip the source if the file doesn't exist
		// because we don't want silent failures/fallbacks.
		// If this specific need arises, a new Source can be implemented.
		return nil, errwrap.New(err)
	}

	err = s.fn(reader, &raw)
	errClose := reader.Close()
	return raw, errwrap.New([]error{errwrap.New(err), errwrap.New(errClose)})
}

// FromBytes returns a configuration Source that unmarshals the given bytes directly.
func FromBytes(b []byte, fn UnmarshalFunc) Source {
	return &bytesSource{
		b:  b,
		fn: fn,
	}
}

// FromReader returns a configuration Source that supports data streaming.
// Usually, the unmarshaler will consume the entirety of the reader until io.EOF
// is encountered. This behavior depends on the given UnmarshalReadFunc.
func FromReader(r io.Reader, fn UnmarshalReadFunc) Source {
	return &readerSource{
		r:  r,
		fn: fn,
	}
}

// FromFile returns a configuration Source reading a file. The file can be from
// the OS filesystem (using `*osfs.FS` or `*os.Root.FS()`) or an embedded file system.
//
//	config.FromFile(&osfs.FS{}, "config.json", config.UnsmarshalReadJSON())
func FromFile(fs fs.FS, fileName string, fn UnmarshalReadFunc) Source {
	return &fileSource{
		fs:       fs,
		fileName: fileName,
		fn:       fn,
	}
}

// Default source loads default config and the config.json file in the current working directory.
// If the "ENV" env variable is set, the config file will be picked like so:
//   - "production": "config.production.json"
//   - "test": "config.test.json"
//   - By default: "config.json"
func Default() Source {
	return FromFile(&osfs.FS{}, getConfigFilePath(os.Getenv("ENV")), UnsmarshalReadJSON())
}

func getConfigFilePath(env string) string {
	env = strings.ToLower(env)
	if env == "local" || env == "localhost" || env == "" {
		return "config.json"
	}
	return "config." + env + ".json"
}
