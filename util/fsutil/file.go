package fsutil

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"sync"

	pathutil "path"

	"github.com/google/uuid"
	"goyave.dev/goyave/v5/util/errors"
)

// marshalCache temporarily stores files' `*multipart.FileHeader`. This type
// cannot be marshaled, making the use of `fsutil.file` inconvenient with DTO conversion.
// The key should a be unique ID. The key is removed from the map.
// To avoid infinite growth of this cache, leading to potential memory problems, this map
// is reset every time its length goes back to 0.
var marshalCache = map[string]*multipart.FileHeader{}
var cacheMu sync.RWMutex

// File represents a file received from client.
//
// File implements `json.Marshaler` and `json.Unmarshaler` to be able
// to be used in DTO conversion (`typeutil.Convert()`). This works with a global
// concurrency-safe map that acts as a cache for the `*multipart.FileHeader`.
// When marshaling, a UUID v1 is generated and used as a key. This UUID is the actual value
// used when marhsaling the `Header` field. When unmarshaling, the `*multipart.FileHeader` is
// retrieved then deleted from the cache. To avoid orphans clogging up the cache, you should
// never JSON marshal this type outside of `typeutil.Convert()`: if a marshaled File never gets
// unmarshaled, its UUID would remain in the cache forever.
type File struct {
	Header   *multipart.FileHeader
	MIMEType string
}

type marshaledFile struct {
	MIMEType string
	Header   string
}

// MarshalJSON implementation of `json.Marhsaler`.
func (file File) MarshalJSON() ([]byte, error) {
	headerUID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	uidStr := headerUID.String()
	cacheMu.Lock()
	marshalCache[uidStr] = file.Header
	cacheMu.Unlock()

	return json.Marshal(marshaledFile{
		Header:   uidStr,
		MIMEType: file.MIMEType,
	})
}

// UnmarshalJSON implementation of `json.Unmarhsaler`.
func (file *File) UnmarshalJSON(data []byte) error {
	var v marshaledFile
	if err := json.Unmarshal(data, &v); err != nil {
		return errors.New(err)
	}

	file.MIMEType = v.MIMEType

	cacheMu.RLock()
	header, ok := marshalCache[v.Header]
	cacheMu.RUnlock()
	if !ok {
		return errors.New("cannot unmarshal fsutil.File: multipart header not found in cache")
	}

	cacheMu.Lock()
	delete(marshalCache, v.Header)
	if len(marshalCache) == 0 {
		// Maps never shrink, let's allocate a new empty map to reset the cache capacity
		// and allow garbage collecting.
		marshalCache = map[string]*multipart.FileHeader{}
	}
	cacheMu.Unlock()

	file.Header = header
	return nil
}

// Save writes the file's content to a new file in the given file system.
// Appends a timestamp to the given file name to avoid duplicate file names.
// The file is not readable anymore once saved as its FileReader has already been
// closed.
//
// Creates directories if needed.
//
// Returns the actual file name.
func (file *File) Save(fs WritableFS, path string, name string) (filename string, err error) {
	filename = timestampFileName(name)

	if mkdirFS, ok := fs.(MkdirFS); ok {
		if err = mkdirFS.MkdirAll(path, os.ModePerm); err != nil {
			err = errors.New(err)
			return
		}
	}

	var f multipart.File
	f, err = file.Header.Open()
	if err != nil {
		err = errors.New(err)
		return
	}
	defer func() {
		closeError := f.Close()
		if err == nil && closeError != nil {
			err = errors.New(closeError)
		}
	}()

	var writer io.ReadWriteCloser
	writer, err = fs.OpenFile(pathutil.Join(path, filename), os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		err = errors.New(err)
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
