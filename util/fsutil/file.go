package fsutil

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"goyave.dev/goyave/v5/util/errors"
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
		err = errors.NewSkip(err, 3)
		return
	}

	var f multipart.File
	f, err = file.Header.Open()
	if err != nil {
		err = errors.NewSkip(err, 3)
		return
	}
	defer func() {
		closeError := f.Close()
		if err == nil && closeError != nil {
			err = errors.New(closeError)
		}
	}()

	var writer *os.File
	writer, err = os.OpenFile(path+string(os.PathSeparator)+filename, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		err = errors.NewSkip(err, 3)
		return
	}
	defer func() {
		closeError := writer.Close()
		if err == nil && closeError != nil {
			err = errors.New(closeError)
		}
	}()
	_, err = io.Copy(writer, f)
	if err != nil {
		err = errors.New(err)
	}
	return
}

// ParseMultipartFiles parse a single file field in a request.
func ParseMultipartFiles(headers []*multipart.FileHeader) ([]File, error) {
	files := []File{}
	for _, fh := range headers {

		fileHeader := make([]byte, 512)

		if fh.Size != 0 {
			f, err := fh.Open()
			if err != nil {
				return nil, errors.New(err)
			}
			if _, err := f.Read(fileHeader); err != nil {
				_ = f.Close()
				return nil, errors.New(err)
			}

			if _, err := f.Seek(0, 0); err != nil {
				_ = f.Close()
				return nil, errors.New(err)
			}
			if err := f.Close(); err != nil {
				return nil, errors.New(err)
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
