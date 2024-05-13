package osfs

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"

	"goyave.dev/goyave/v5/util/errors"
)

// FS implementation of `fsutil.FS` for the local OS file system.
type FS struct {
	// dir the path prefix for Sub file systems.
	dir string
}

// New create a new OS file system for the tree of files rooted at the directory "baseDir".
//
// Giving an empty string will use the current working directory as a base.
// The path is cleaned using `path.Clean()`.
func New(baseDir string) *FS {
	return &FS{dir: path.Clean(baseDir)}
}

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode `O_RDONLY`.
// If there is an error, it will be of type `*PathError“.
func (f *FS) Open(name string) (fs.File, error) {
	file, err := os.Open(path.Join(f.dir, name))
	return file, errors.NewSkip(err, 3)
}

// OpenFile is the generalized open call. It opens the named file with specified flag
// (`O_RDONLY` etc.). If the file does not exist, and the `O_CREATE` flag
// is passed, it is created with mode perm (before umask). If successful,
// methods on the returned file can be used for I/O.
// If there is an error, it will be of type `*PathError`.
func (f *FS) OpenFile(name string, flag int, perm fs.FileMode) (io.ReadWriteCloser, error) {
	rwc, err := os.OpenFile(path.Join(f.dir, name), flag, perm)
	return rwc, errors.NewSkip(err, 3)
}

// ReadDir reads the named directory,
// returning all its directory entries sorted by filename.
// If an error occurs reading the directory,
// ReadDir returns the entries it was able to read before the error,
// along with the error.
func (f *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := os.ReadDir(path.Join(f.dir, name))
	return entries, errors.NewSkip(err, 3)
}

// Stat returns a FileInfo describing the named file.
// If there is an error, it will be of type `*PathError`.
func (f *FS) Stat(name string) (fs.FileInfo, error) {
	info, err := os.Stat(path.Join(f.dir, name))
	return info, errors.NewSkip(err, 3)
}

// Getwd returns a rooted path name corresponding to the
// current directory. If the current directory can be
// reached via multiple paths (due to symbolic links),
// Getwd may return any one of them.
func (*FS) Getwd() (string, error) {
	wd, err := os.Getwd()
	return wd, errors.NewSkip(err, 3)
}

// FileExists returns true if the file at the given path exists and is readable.
// Returns false if the given file is a directory.
func (f *FS) FileExists(name string) bool {
	if stats, err := f.Stat(path.Join(f.dir, name)); err == nil {
		return !stats.IsDir()
	}
	return false
}

// IsDirectory returns true if the file at the given path exists, is a directory and is readable.
func (f *FS) IsDirectory(name string) bool {
	if stats, err := f.Stat(path.Join(f.dir, name)); err == nil {
		return stats.IsDir()
	}
	return false
}

// MkdirAll creates a directory, along with any necessary parents,
// and returns `nil`, or else returns an error.
// The permission bits perm (before umask) are used for all
// directories that `MkdirAll` creates.
// If path is already a directory, `MkdirAll` does nothing
// and returns `nil`.
func (f *FS) MkdirAll(name string, perm fs.FileMode) error {
	return errors.NewSkip(os.MkdirAll(path.Join(f.dir, name), perm), 3)
}

// Mkdir creates a new directory with the specified name and permission
// bits (before umask).
// If there is an error, it will be of type `*PathError`.
func (f *FS) Mkdir(name string, perm fs.FileMode) error {
	return errors.NewSkip(os.Mkdir(path.Join(f.dir, name), perm), 3)
}

// Remove removes the named file or (empty) directory.
// If there is an error, it will be of type `*PathError`.
func (f *FS) Remove(name string) error {
	return errors.NewSkip(os.Remove(path.Join(f.dir, name)), 3)
}

// RemoveAll removes the element at the given path and any children it contains.
// It removes everything it can but returns the first error
// it encounters. If the path does not exist, `RemoveAll`
// returns `nil` (no error).
// If there is an error, it will be of type `*PathError`.
func (f *FS) RemoveAll(name string) error {
	return errors.NewSkip(os.RemoveAll(path.Join(f.dir, name)), 3)
}

// Sub returns an `*osfs.FS` corresponding to the subtree rooted at this fs's dir.
// If dir is ".", the same `&osfs.FS` is returned.
//
// Because `*osfs.FS` internally uses the functions from the `os` package, bear in mind
// that changing the working directory will affect all instances of `*osfs.FS`.
//
// You can't use `Sub` if the current `osfs.FS` has a rooted prefix.
func (f *FS) Sub(dir string) (*FS, error) {
	if dir == "." {
		return f, nil
	}
	newDir := path.Join(f.dir, dir)
	if !fs.ValidPath(newDir) {
		return nil, &fs.PathError{Op: "sub", Path: newDir, Err: fmt.Errorf("invalid name")}
	}
	return &FS{dir: newDir}, nil
}
