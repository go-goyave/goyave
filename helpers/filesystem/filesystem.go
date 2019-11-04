package filesystem

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// File represents a file received from client.
type File struct {
	Header   *multipart.FileHeader
	MIMEType string
	Data     multipart.File
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

// GetMimeType get the mime type and size of the given file.
func GetMimeType(file string) (string, int64) {
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

	return http.DetectContentType(buffer), stat.Size()
}

// FileExists returns true if the file at the given path exists and is readable.
// Returns false if the given file is a directory
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

// Save writes the given file on the disk.
// Appends a timestamp to the given file name to avoid duplicate file names.
// The file is not readable anymore once saved as its FileReader has already been
// closed.
//
// Returns the actual path to the saved file.
func Save(file File, path string, name string) string {
	name = timestampFileName(name)
	writer, err := os.OpenFile(path+string(os.PathSeparator)+name, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		panic(err)
	}
	defer writer.Close()
	_, errCopy := io.Copy(writer, file.Data)
	if errCopy != nil {
		panic(errCopy)
	}
	file.Data.Close()
	return name
}

// Delete the file at the given path.
//
// To avoid panics, you should check if the file exists.
func Delete(path string) {
	err := os.Remove(path)
	if err != nil {
		panic(err)
	}
}

// ParseMultipartFiles parse a single file field in a request.
func ParseMultipartFiles(request *http.Request, field string) []File {
	files := []File{}
	for _, fh := range request.MultipartForm.File[field] {
		f, err := fh.Open()
		if err != nil {
			panic(err)
		}
		defer f.Close()

		fileHeader := make([]byte, 512)

		if _, err := f.Read(fileHeader); err != nil {
			panic(err)
		}

		if _, err := f.Seek(0, 0); err != nil {
			panic(err)
		}

		file := File{
			Header:   fh,
			MIMEType: http.DetectContentType(fileHeader),
			Data:     f,
		}
		files = append(files, file)
	}
	return files
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
