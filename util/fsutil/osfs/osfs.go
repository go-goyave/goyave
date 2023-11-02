package osfs

import (
	"io/fs"
	"os"
)

// TODO document and test OSFS

type FS struct{}

func (FS) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func (FS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

func (FS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (FS) Getwd() (string, error) {
	return os.Getwd()
}

// FileExists returns true if the file at the given path exists and is readable.
// Returns false if the given file is a directory.
func (fs FS) FileExists(file string) bool {
	if stats, err := os.Stat(file); err == nil {
		return !stats.IsDir()
	}
	return false
}

// IsDirectory returns true if the file at the given path exists, is a directory and is readable.
func (fs FS) IsDirectory(path string) bool {
	if stats, err := os.Stat(path); err == nil {
		return stats.IsDir()
	}
	return false
}
