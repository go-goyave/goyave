package osfs

import (
	"io/fs"
	"os"
)

// TODO test OSFS

// FS implementation of `fsutil.FS` for the local OS file system.
type FS struct{}

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode `O_RDONLY`.
// If there is an error, it will be of type `*PathError“.
func (FS) Open(name string) (fs.File, error) {
	return os.Open(name)
}

// ReadDir reads the named directory,
// returning all its directory entries sorted by filename.
// If an error occurs reading the directory,
// ReadDir returns the entries it was able to read before the error,
// along with the error.
func (FS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

// Stat returns a FileInfo describing the named file.
// If there is an error, it will be of type `*PathError`.
func (FS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

// Getwd returns a rooted path name corresponding to the
// current directory. If the current directory can be
// reached via multiple paths (due to symbolic links),
// Getwd may return any one of them.
func (FS) Getwd() (string, error) {
	return os.Getwd()
}

// FileExists returns true if the file at the given path exists and is readable.
// Returns false if the given file is a directory.
func (fs FS) FileExists(file string) bool {
	if stats, err := fs.Stat(file); err == nil {
		return !stats.IsDir()
	}
	return false
}

// IsDirectory returns true if the file at the given path exists, is a directory and is readable.
func (fs FS) IsDirectory(path string) bool {
	if stats, err := fs.Stat(path); err == nil {
		return stats.IsDir()
	}
	return false
}
