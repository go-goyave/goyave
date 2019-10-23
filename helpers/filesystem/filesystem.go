package filesystem

import (
	"io"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"time"
)

// File represents a file received from client.
type File struct {
	Header *multipart.FileHeader
	Data   multipart.File
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

// Save writes the given file on the disk.
// Appends a timestamp to the given file name to avoid duplicate file names.
//
// Returns the actual path to the saved file.
func Save(file File, path string, name string) string {
	name = timestampFileName(name)
	writer, err := os.OpenFile(path+string(os.PathSeparator)+name, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		panic(err)
	}
	defer writer.Close()
	io.Copy(writer, file.Data)
	return name
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
	return prefix + "-" + strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10) + extension
}
