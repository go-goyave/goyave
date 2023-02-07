package fsutil

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// File represents a file received from client.
type File struct {
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
func (file *File) Save(path string, name string) (filename string, err error) { // TODO handle io.FS
	filename = timestampFileName(name)

	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return
	}

	var f multipart.File
	f, err = file.Header.Open()
	if err != nil {
		return
	}
	defer func() {
		closeError := f.Close()
		if err == nil {
			err = closeError
		}
	}()

	var writer *os.File
	writer, err = os.OpenFile(path+string(os.PathSeparator)+filename, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		return
	}
	defer func() {
		closeError := writer.Close()
		if err == nil {
			err = closeError
		}
	}()
	_, err = io.Copy(writer, f)
	return
}

// ParseMultipartFiles parse a single file field in a request.
func ParseMultipartFiles(headers []*multipart.FileHeader) ([]File, error) {
	files := []File{}
	for _, fh := range headers {
		f, err := fh.Open()
		if err != nil {
			return nil, err
		}

		fileHeader := make([]byte, 512)

		if fh.Size != 0 {
			if _, err := f.Read(fileHeader); err != nil {
				return nil, err
			}

			if _, err := f.Seek(0, 0); err != nil {
				return nil, err
			}
		}

		file := File{
			Header:   fh,
			MIMEType: http.DetectContentType(fileHeader),
		}
		files = append(files, file)
	}
	return files, nil
}
