package fsutil

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// File represents a file received from client.
type File struct {
	Data     multipart.File
	Header   *multipart.FileHeader
	MIMEType string
}

// Save writes the given file on the disk.
// Appends a timestamp to the given file name to avoid duplicate file names.
// The file is not readable anymore once saved as its FileReader has already been
// closed.
//
// Creates directories if needed.
//
// Returns the actual file name.
func (file *File) Save(path string, name string) string {
	name = timestampFileName(name)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		panic(err)
	}
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

		if fh.Size != 0 {
			if _, err := f.Read(fileHeader); err != nil {
				panic(err)
			}

			if _, err := f.Seek(0, 0); err != nil {
				panic(err)
			}
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
