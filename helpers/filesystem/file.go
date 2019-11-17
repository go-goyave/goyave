package filesystem

import (
	"io"
	"mime/multipart"
	"os"
)

// File represents a file received from client.
type File struct {
	Header   *multipart.FileHeader
	MIMEType string
	Data     multipart.File
}

// Save writes the given file on the disk.
// Appends a timestamp to the given file name to avoid duplicate file names.
// The file is not readable anymore once saved as its FileReader has already been
// closed.
//
// Returns the actual path to the saved file.
func (file *File) Save(path string, name string) string {
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
