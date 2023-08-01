package fsutil

import (
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"goyave.dev/goyave/v5/util/errors"
)

var contentTypeByExtension = map[string]string{
	".jsonld": "application/ld+json",
	".json":   "application/json",
	".js":     "text/javascript",
	".mjs":    "text/javascript",
	".css":    "text/css",
}

// GetFileExtension returns the last part of a file name.
// If the file doesn't have an extension, returns an empty string.
func GetFileExtension(file string) string {
	index := strings.LastIndex(file, ".")
	if index == -1 {
		return ""
	}
	return file[index+1:]
}

// GetMIMEType get the mime type and size of the given file.
// This function calls `http.DetectContentType`. If the detected content type
// could not be determined or if it's a text file, `GetMIMEType` will attempt to
// detect the MIME type based on the file extension. The following extensions are
// supported:
//   - `.jsonld`: "application/ld+json"
//   - `.json`: "application/json"
//   - `.js` / `.mjs`: "text/javascript"
//   - `.css`: "text/css"
//
// If a specific MIME type cannot be determined, returns "application/octet-stream" as a fallback.
func GetMIMEType(file string) (contentType string, size int64, err error) {
	var f *os.File
	f, err = os.Open(file)
	if err != nil {
		err = errors.NewSkip(err, 3)
		return
	}
	defer func() {
		errClose := f.Close()
		if err == nil && errClose != nil {
			err = errors.New(errClose)
		}
	}()

	var stat fs.FileInfo
	stat, err = f.Stat()
	if err != nil {
		err = errors.NewSkip(err, 3)
		return
	}

	size = stat.Size()

	buffer := make([]byte, 512)
	contentType = "application/octet-stream"

	if size != 0 {
		_, err = f.Read(buffer)
		if err != nil {
			err = errors.NewSkip(err, 3)
			return
		}

		contentType = http.DetectContentType(buffer)
	}

	if strings.HasPrefix(contentType, "application/octet-stream") || strings.HasPrefix(contentType, "text/plain") {
		for ext, t := range contentTypeByExtension {
			if strings.HasSuffix(file, ext) {
				tmp := t
				if i := strings.Index(contentType, ";"); i != -1 {
					tmp = t + contentType[i:]
				}
				contentType = tmp
				break
			}
		}
	}

	return
}

// FileExists returns true if the file at the given path exists and is readable.
// Returns false if the given file is a directory.
func FileExists(file string) bool {
	if stats, err := os.Stat(file); err == nil {
		return !stats.IsDir()
	}
	return false
}

// IsDirectory returns true if the file at the given path exists, is a directory and is readable.
func IsDirectory(path string) bool {
	if stats, err := os.Stat(path); err == nil {
		return stats.IsDir()
	}
	return false
}

// Delete the file at the given path.
//
// To avoid panics, you should check if the file exists.
func Delete(path string) {
	err := os.Remove(path)
	if err != nil {
		panic(errors.NewSkip(err, 3))
	}
}

func timestampFileName(name string) string {
	var prefix string
	var extension string
	index := strings.LastIndex(name, ".")
	if index == -1 {
		prefix = name
		extension = ""
	} else {
		prefix = name[:index]
		extension = name[index:]
	}
	return prefix + "-" + strconv.FormatInt(time.Now().UnixNano()/int64(time.Microsecond), 10) + extension
}
