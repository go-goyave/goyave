package osfs

import (
	"io"
	"io/fs"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5/util/fsutil"
)

func setRootWorkingDirectory() {
	_, filename, _, _ := runtime.Caller(1)
	directory := path.Dir(filename)
	for !fsutil.FileExists(&FS{}, path.Join(directory, "go.mod")) {
		directory = path.Join(directory, "..")
		if !fsutil.IsDirectory(&FS{}, directory) {
			panic("Couldn't find project's root directory.")
		}
	}
	if err := os.Chdir(directory); err != nil {
		panic(err)
	}
}

func TestOSFS(t *testing.T) {

	setRootWorkingDirectory()

	t.Run("Open", func(t *testing.T) {
		fs := &FS{}
		file, err := fs.Open("resources/test_file.txt")
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, file.Close())
		}()
		contents, err := io.ReadAll(file)
		assert.NoError(t, err)
		assert.Equal(t, append([]byte{0xef, 0xbb, 0xbf}, []byte("utf-8 with BOM content")...), contents)
	})

	t.Run("OpenFile", func(t *testing.T) {
		fs := &FS{}
		file, err := fs.OpenFile("resources/test_file.txt", os.O_RDONLY, 0660)
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, file.Close())
		}()
		contents, err := io.ReadAll(file)
		assert.NoError(t, err)
		assert.Equal(t, append([]byte{0xef, 0xbb, 0xbf}, []byte("utf-8 with BOM content")...), contents)
	})

	t.Run("ReadDir", func(t *testing.T) {
		osfs := &FS{}
		entries, err := osfs.ReadDir("resources/lang")
		require.NoError(t, err)

		type result struct {
			name  string
			isdir bool
		}

		expected := []result{
			{name: "en-UK", isdir: true},
			{name: "en-US", isdir: true},
			{name: "invalid.json", isdir: false},
		}

		assert.Equal(t, expected, lo.Map(entries, func(e fs.DirEntry, _ int) result {
			return result{name: e.Name(), isdir: e.IsDir()}
		}))
	})

	t.Run("Stat", func(t *testing.T) {
		fs := &FS{}
		info, err := fs.Stat("resources/test_file.txt")
		require.NoError(t, err)

		assert.False(t, info.IsDir())
		assert.Equal(t, "test_file.txt", info.Name())
	})

	t.Run("Getwd", func(t *testing.T) {
		fs := &FS{}
		wd, err := fs.Getwd()
		require.NoError(t, err)
		assert.NotEmpty(t, wd)
	})

	t.Run("FileExists", func(t *testing.T) {
		fs := &FS{}
		assert.True(t, fs.FileExists("resources/test_file.txt"))
		assert.False(t, fs.FileExists("resources"))
		assert.False(t, fs.FileExists("resources/notafile.txt"))
	})

	t.Run("IsDirectory", func(t *testing.T) {
		fs := &FS{}
		assert.False(t, fs.IsDirectory("resources/test_file.txt"))
		assert.True(t, fs.IsDirectory("resources"))
		assert.False(t, fs.IsDirectory("resources/notadir"))
	})

	t.Run("Mkdir", func(t *testing.T) {
		fs := &FS{}
		path := "resources/testdir"

		t.Cleanup(func() {
			if err := os.RemoveAll(path); err != nil {
				panic(err)
			}
		})

		require.NoError(t, fs.Mkdir(path, 0770))
		assert.True(t, fs.IsDirectory(path))
	})

	t.Run("MkdirAll", func(t *testing.T) {
		fs := &FS{}
		path := "resources/testdirall/subdir"

		t.Cleanup(func() {
			if err := os.RemoveAll("resources/testdirall"); err != nil {
				panic(err)
			}
		})

		require.NoError(t, fs.MkdirAll(path, 0770))
		assert.True(t, fs.IsDirectory(path))
	})

	t.Run("Remove", func(t *testing.T) {
		fs := &FS{}
		path := "resources/testdirremove"
		assert.NoError(t, fs.MkdirAll(path, 0770))
		assert.True(t, fs.IsDirectory(path))
		assert.NoError(t, fs.Remove(path))
		assert.False(t, fs.IsDirectory(path))
	})

	t.Run("RemoveAll", func(t *testing.T) {
		fs := &FS{}
		path := "resources/testdirremoveall/subdir"
		assert.NoError(t, fs.MkdirAll(path, 0770))
		assert.True(t, fs.IsDirectory(path))
		assert.NoError(t, fs.RemoveAll("resources/testdirremoveall"))
		assert.False(t, fs.IsDirectory(path))
		assert.False(t, fs.IsDirectory("resources/testdirremoveall"))
	})

	t.Run("Sub", func(t *testing.T) {
		f, err := (&FS{}).Sub("resources")
		require.NoError(t, err)
		require.Equal(t, "resources", f.dir)

		file, err := f.Open("test_file.txt")
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, file.Close())
		}()
		contents, err := io.ReadAll(file)
		assert.NoError(t, err)
		assert.Equal(t, append([]byte{0xef, 0xbb, 0xbf}, []byte("utf-8 with BOM content")...), contents)

		subfs, err := f.Sub("lang")
		require.NoError(t, err)
		require.Equal(t, "resources/lang", subfs.dir)

		file2, err := subfs.Open("invalid.json")
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, file2.Close())
		}()
		contents, err = io.ReadAll(file2)
		assert.NoError(t, err)
		assert.Equal(t, []byte("[]"), contents)

		dotFS, err := subfs.Sub(".")
		assert.NoError(t, err)
		assert.Same(t, subfs, dotFS)

		_, err = subfs.Sub("\xC0")
		require.Error(t, err)
		if pathErr, ok := err.(*fs.PathError); assert.True(t, ok) {
			assert.Equal(t, "sub", pathErr.Op)
			assert.Equal(t, "resources/lang/\xC0", pathErr.Path)
			assert.Equal(t, "invalid name", pathErr.Err.Error())
		}
	})

	t.Run("New", func(t *testing.T) {
		fs := New("/home")
		assert.Equal(t, "/home", fs.dir)
	})
}
