package fsutil

import (
	"bytes"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"goyave.dev/goyave/v5/util/errwrap"
)

var contentTypeByExtension = map[string]string{
	".ai":     "application/postscript",
	".apk":    "application/vnd.android.package-archive",
	".apng":   "image/apng",
	".avif":   "image/avif",
	".bin":    "application/octet-stream",
	".bmp":    "image/bmp",
	".com":    "application/octet-stream",
	".doc":    "application/msword",
	".css":    "text/css",
	".csv":    "text/csv",
	".docx":   "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".ehtml":  "text/html",
	".eml":    "message/rfc822",
	".eps":    "application/postscript",
	".exe":    "application/octet-stream",
	".flac":   "audio/flac",
	".gif":    "image/gif",
	".gz":     "application/gzip",
	".htm":    "text/html",
	".html":   "text/html",
	".ico":    "image/vnd.microsoft.icon",
	".ics":    "text/calendar",
	".jfif":   "image/jpeg",
	".jpg":    "image/jpeg",
	".jpeg":   "image/jpeg",
	".js":     "text/javascript",
	".jsonld": "application/ld+json",
	".json":   "application/json",
	".m4a":    "audio/mp4",
	".mjs":    "text/javascript",
	".mp3":    "audio/mpeg",
	".mp4":    "video/mp4",
	".mpeg":   "audio/mpeg",
	".ods":    "application/vnd.oasis.opendocument.spreadsheet",
	".odt":    "application/vnd.oasis.opendocument.text",
	".oga":    "audio/ogg",
	".ogv":    "video/ogg",
	".ogx":    "application/ogg",
	".opus":   "audio/ogg",
	".otf":    "font/otf",
	".png":    "image/png",
	".pdf":    "application/pdf",
	".pjp":    "image/jpeg",
	".pjpeg":  "image/jpeg",
	".ppt":    "application/vnd.ms-powerpoint",
	".pptx":   "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".ps":     "application/postscript",
	".rdf":    "application/rdf+xml",
	".rtf":    "application/rtf",
	".sh":     "application/x-sh",
	".shtml":  "text/html",
	".svg":    "image/svg+xml",
	".tar":    "application/x-tar",
	".text":   "text/plain",
	".tif":    "image/tiff",
	".tiff":   "image/tiff",
	".ts":     "video/mp2t",
	".ttf":    "font/ttf",
	".txt":    "text/plain",
	".vtt":    "text/vtt",
	".wasm":   "application/wasm",
	".wav":    "audio/wav",
	".weba":   "audio/webm",
	".webm":   "audio/webm",
	".webp":   "image/webp",
	".woff":   "font/woff",
	".woff2":  "font/woff2",
	".xbl":    "text/xml",
	".xbm":    "image/x-xbitmap",
	".xht":    "application/xhtml+xml",
	".xhtml":  "application/xhtml+xml",
	".xls":    "application/vnd.ms-excel",
	".xlsx":   "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".xml":    "application/xml",
	".xsl":    "text/xml",
	".zip":    "application/zip",
	".7z":     "application/x-7z-compressed",
}

// AddExtensionType set the MIME type associated with the given extension.
// The extension should begin with a dot (e.g.: ".html").
// The mimeType should not include the charset parameter nd be written in lowercase.
//
// Passing an extension that is already registered overrides the previous value.
//
// This function is not safe for concurrent use.
func AddExtensionType(ext, mimeType string) error {
	if !strings.HasPrefix(ext, ".") {
		return errwrap.Errorf("fsutil: extension %q missing leading dot", ext)
	}
	_, params, err := mime.ParseMediaType(mimeType)
	if err != nil {
		return errwrap.New(err)
	}
	if len(params) > 0 {
		return errwrap.Errorf("fsutil: MIME type %q contains a parameter", mimeType)
	}
	contentTypeByExtension[ext] = mimeType
	return nil
}

// GetFileExtension returns the last part of a file name, without the leading dot.
// If the file doesn't have an extension, returns an empty string.
// For files with multiple extensions like `.tar.gz`, only `gz` is returned.
func GetFileExtension(filename string) string {
	index := strings.LastIndex(filename, ".")
	if index == -1 {
		return ""
	}
	return filename[index+1:]
}

// GetMIMEType get the mime type and size of the given file.
// This function opens the file, stats it and calls `fsutil.DetectContentType`.
// If the file is empty (size of 0), the content-type will be detected using `fsutil.DetectContentTypeByExtension`.
// If a specific MIME type cannot be determined, returns "application/octet-stream" as a fallback.
func GetMIMEType(filesystem fs.FS, file string) (contentType string, size int64, err error) {
	var f fs.File
	f, err = filesystem.Open(file)
	if err != nil {
		err = errwrap.New(err)
		return
	}
	defer func() {
		errClose := f.Close()
		if err == nil && errClose != nil {
			err = errwrap.New(errClose)
		}
	}()

	var stat fs.FileInfo
	stat, err = f.Stat()
	if err != nil {
		err = errwrap.New(err)
		return
	}

	size = stat.Size()

	if size == 0 {
		contentType = DetectContentTypeByExtension(file)
		return
	}

	contentType, err = DetectContentType(f, file)
	if err != nil {
		err = errwrap.New(err)
	}

	return
}

// DetectContentType by sniffing the first 512 bytes of the given reader using `http.DetectContentType`.
//
// If the detected content type is `"application/octet-stream"` or `"text/plain"`, this function will attempt to
// find a more precise one using `fsutil.DetectContentTypeByExtension`, unless `fileName` is empty.
// If the detected content type is `"text/xml"` or `"application/xml"`, this function promotes it to
// `"image/svg+xml"` only if the content signature indicates SVG.
// The header parameter is retained (e.g: `charset=utf-8`).
//
// If there is no error, this function always returns a valid MIME type. If it cannot determine a more specific one,
// it returns `"application/octet-stream"`.
//
// If the given reader implements `io.Seeker`, the reader's offset is reset to the start.
func DetectContentType(r io.Reader, fileName string) (string, error) {
	buffer := make([]byte, 512)
	_, err := r.Read(buffer)
	if err != nil {
		return "", errwrap.New(err)
	}
	if seeker, ok := r.(io.Seeker); ok {
		_, err = seeker.Seek(0, io.SeekStart)
		if err != nil {
			return "", errwrap.New(err)
		}
	}

	contentType := http.DetectContentType(buffer)
	if strings.HasPrefix(contentType, "application/octet-stream") || strings.HasPrefix(contentType, "text/plain") {
		contentType = detectContentTypeByExtension(fileName, contentType)
	} else if (strings.HasPrefix(contentType, "text/xml") || strings.HasPrefix(contentType, "application/xml")) && hasSVGSignature(buffer) {
		contentType = "image/svg+xml"
	}
	return contentType, nil
}

func hasSVGSignature(buffer []byte) bool {
	start := bytes.IndexRune(buffer, '<')
	if start == -1 {
		return false
	}

	content := strings.TrimSpace(strings.ToLower(string(buffer[start:])))
	// Skip optional XML tag and possible comments
	for {
		if strings.HasPrefix(content, "<?xml") {
			end := strings.Index(content, "?>")
			if end == -1 {
				break
			}
			content = strings.TrimSpace(content[end+2:])
			continue
		}
		if strings.HasPrefix(content, "<!--") {
			end := strings.Index(content, "-->")
			if end == -1 {
				break
			}
			content = strings.TrimSpace(content[end+3:])
			continue
		}
		break
	}

	return strings.HasPrefix(content, "<svg")
}

func detectContentTypeByExtension(fileName, contentType string) string {
	if fileName == "" {
		return contentType
	}
	for ext, t := range contentTypeByExtension {
		if strings.HasSuffix(fileName, ext) {
			tmp := t
			if i := strings.Index(contentType, ";"); i != -1 {
				tmp = t + contentType[i:] // Keep the "charset" arguments
			}
			contentType = tmp
			break
		}
	}
	return contentType
}

// DetectContentTypeByExtension returns a MIME type associated with the extension (suffix) of the given file name.
// Note that this function should be used as a fallback if sniffing using `fsutil.DetectContentType`
// isn't possible due to:
//   - the file being empty (size of 0)
//   - the file requiring to be read after the sniffing but its reader doesn't implement `io.Seeker` for rewinding.
//
// The local database covers the most common MIME types.
// If the extension is not known to the `fsutil` package, returns `"application/octet-stream"`.
func DetectContentTypeByExtension(fileName string) string {
	return detectContentTypeByExtension(fileName, "application/octet-stream")
}

// FileExists returns true if the file at the given path exists and is readable.
// Returns false if the given file is a directory.
func FileExists(fs fs.StatFS, file string) bool {
	if stats, err := fs.Stat(file); err == nil {
		return !stats.IsDir()
	}
	return false
}

// IsDirectory returns true if the file at the given path exists, is a directory and is readable.
func IsDirectory(fs fs.StatFS, path string) bool {
	if stats, err := fs.Stat(path); err == nil {
		return stats.IsDir()
	}
	return false
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

// An FS provides access to a hierarchical file system
// and implements `io/fs`'s `FS`, `ReadDirFS` and `StatFS` interfaces.
type FS interface {
	fs.ReadDirFS
	fs.StatFS
}

// A WorkingDirFS is a file system with a `Getwd()` method.
type WorkingDirFS interface {
	// Getwd returns a rooted path name corresponding to the
	// current directory. If the current directory can be
	// reached via multiple paths (due to symbolic links),
	// Getwd may return any one of them.
	Getwd() (dir string, err error)
}

// A MkdirFS is a file system with a `Mkdir()` and a `MkdirAll()` methods.
type MkdirFS interface {
	// MkdirAll creates a directory named path,
	// along with any necessary parents, and returns `nil`,
	// or else returns an error.
	// The permission bits perm (before umask) are used for all
	// directories that `MkdirAll` creates.
	// If path is already a directory, `MkdirAll` does nothing
	// and returns `nil`.
	MkdirAll(path string, perm fs.FileMode) error

	// Mkdir creates a new directory with the specified name and permission
	// bits (before umask).
	// If there is an error, it will be of type `*PathError`.
	Mkdir(path string, perm fs.FileMode) error
}

// A WritableFS is a file system with a `OpenFile()` method.
type WritableFS interface {
	// OpenFile is the generalized open call. It opens the named file with specified flag
	// (`O_RDONLY` etc.). If the file does not exist, and the `O_CREATE` flag
	// is passed, it is created with mode perm (before umask). If successful,
	// methods on the returned file can be used for I/O.
	// If there is an error, it will be of type `*PathError`.
	OpenFile(path string, flag int, perm fs.FileMode) (io.ReadWriteCloser, error)
}

// A RemoveFS is a file system with a `Remove()` and a `RemoveAll()` methods.
type RemoveFS interface {
	// Remove removes the named file or (empty) directory.
	// If there is an error, it will be of type `*PathError`.
	Remove(path string) error

	// RemoveAll removes path and any children it contains.
	// It removes everything it can but returns the first error
	// it encounters. If the path does not exist, `RemoveAll`
	// returns `nil` (no error).
	// If there is an error, it will be of type `*PathError`.
	RemoveAll(path string) error
}

// Embed is an extension of aimed at improving `embed.FS` by
// implementing `fs.StatFS` and a `Sub()` function.
type Embed struct {
	FS fs.ReadDirFS
}

// NewEmbed returns a new Embed with the given FS.
func NewEmbed(fs fs.ReadDirFS) Embed {
	return Embed{
		FS: fs,
	}
}

// Open opens the named file.
//
// When Open returns an error, it should be of type *PathError
// with the Op field set to "open", the Path field set to name,
// and the Err field describing the problem.
//
// Open should reject attempts to open names that do not satisfy
// ValidPath(name), returning a *PathError with Err set to
// ErrInvalid or ErrNotExist.
func (e Embed) Open(name string) (fs.File, error) {
	f, err := e.FS.Open(name)
	return f, errwrap.NewSkip(err, 3)
}

// ReadDir reads the named directory
// and returns a list of directory entries sorted by filename.
func (e Embed) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := e.FS.ReadDir(name)
	return entries, errwrap.NewSkip(err, 3)
}

// Stat returns a FileInfo describing the file.
func (e Embed) Stat(name string) (fileinfo fs.FileInfo, err error) {
	if statsFS, ok := e.FS.(fs.StatFS); ok {
		return statsFS.Stat(name)
	}
	f, err := e.FS.Open(name)
	if err != nil {
		return nil, errwrap.New(err)
	}
	defer func() {
		e := f.Close()
		if err == nil && e != nil {
			err = errwrap.New(&fs.PathError{Op: "close", Path: name, Err: e})
		}
	}()

	fileinfo, err = f.Stat()
	if err != nil {
		err = errwrap.New(err)
	}
	return
}

// Sub returns an Embed FS corresponding to the subtree rooted at dir.
// Returns and error if the underlying sub FS doesn't implement `fs.ReadDirFS`.
func (e Embed) Sub(dir string) (Embed, error) {
	sub, err := fs.Sub(e.FS, dir)
	if err != nil {
		return Embed{}, errwrap.NewSkip(err, 3)
	}
	subFS, ok := sub.(fs.ReadDirFS)
	if !ok {
		return Embed{}, errwrap.NewSkip("fsutil.Embed: cannot Sub, underlying sub FS doesn't implement fsutil.FS", 3)
	}
	return Embed{FS: subFS}, nil
}
