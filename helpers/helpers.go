package helpers

import (
	"net/http"
	"os"
)

// FileExists returns true if the file at the given path exists and is readable.
// Returns false if the given file is a directory
func FileExists(file string) bool {
	if stats, err := os.Stat(file); err == nil {
		return !stats.IsDir()
	}
	return false
}

// IsDirectory returns true if the file at the given paht exists, is a directory and is readable.
func IsDirectory(path string) bool {
	if stats, err := os.Stat(path); err == nil {
		return stats.IsDir()
	}
	return false
}

// GetMimeType get the mime type and size of the given file.
func GetMimeType(file string) (string, int64, error) {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buffer := make([]byte, 512)

	_, errRead := f.Read(buffer)
	if errRead != nil {
		panic(errRead)
	}

	stat, errStat := f.Stat()
	if errStat != nil {
		panic(errStat)
	}

	return http.DetectContentType(buffer), stat.Size(), nil
}
