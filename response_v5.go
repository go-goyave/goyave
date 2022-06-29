package goyave

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"

	"goyave.dev/goyave/v4/util/fsutil"
)

type ResponseV5 struct {
	writer         io.Writer
	responseWriter http.ResponseWriter
	httpRequest    *http.Request
	status         int

	// Used to check if controller didn't write anything so
	// core can write default 204 No Content.
	// See RFC 7231, 6.3.5
	empty       bool
	wroteHeader bool
	hijacked    bool
}

// newResponse create a new Response using the given http.ResponseWriter and raw request.
func newResponseV5(writer http.ResponseWriter, rawRequest *http.Request) *ResponseV5 {
	return &ResponseV5{
		responseWriter: writer,
		writer:         writer,
		httpRequest:    rawRequest,
		empty:          true,
		status:         0,
		wroteHeader:    false,
	}
}

// --------------------------------------
// PreWriter implementation

// PreWrite writes the response header after calling PreWrite on the
// child writer if it implements PreWriter.
func (r *ResponseV5) PreWrite(b []byte) {
	r.empty = false
	if pr, ok := r.writer.(PreWriter); ok {
		pr.PreWrite(b)
	}
	if !r.wroteHeader {
		if r.status == 0 {
			r.status = http.StatusOK
		}
		r.WriteHeader(r.status)
	}
}

// --------------------------------------
// http.ResponseWriter implementation

// Write writes the data as a response.
// See http.ResponseWriter.Write
func (r *ResponseV5) Write(data []byte) (int, error) {
	r.PreWrite(data)
	return r.writer.Write(data)
}

// WriteHeader sends an HTTP response header with the provided
// status code.
// Prefer using "Status()" method instead.
// Calling this method a second time will have no effect.
func (r *ResponseV5) WriteHeader(status int) {
	if !r.wroteHeader {
		r.status = status
		r.wroteHeader = true
		r.responseWriter.WriteHeader(status)
	}
}

// Header returns the header map that will be sent.
func (r *ResponseV5) Header() http.Header {
	return r.responseWriter.Header()
}

// --------------------------------------
// http.Hijacker implementation

// Hijack implements the Hijacker.Hijack method.
// For more details, check http.Hijacker.
//
// Returns ErrNotHijackable if the underlying http.ResponseWriter doesn't
// implement http.Hijacker. This can happen with HTTP/2 connections.
//
// Middleware executed after controller handlers, as well as status handlers,
// keep working as usual after a connection has been hijacked.
// Callers should properly set the response status to ensure middleware and
// status handler execute correctly. Usually, callers of the Hijack method
// set the HTTP status to http.StatusSwitchingProtocols.
// If no status is set, the regular behavior will be kept and `204 No Content`
// will be set as the response status.
func (r *ResponseV5) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.responseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, ErrNotHijackable
	}
	c, b, e := hijacker.Hijack()
	if e == nil {
		r.hijacked = true
	}
	return c, b, e
}

// Hijacked returns true if the underlying connection has been successfully hijacked
// via the Hijack method.
func (r *ResponseV5) Hijacked() bool {
	return r.hijacked
}

// --------------------------------------
// Chained writers

// Writer return the current writer used to write the response.
// Note that the returned writer is not necessarily a http.ResponseWriter, as
// it can be replaced using SetWriter.
func (r *ResponseV5) Writer() io.Writer {
	return r.writer
}

// SetWriter set the writer used to write the response.
// This can be used to chain writers, for example to enable
// gzip compression, or for logging.
//
// The original http.ResponseWriter is always kept.
func (r *ResponseV5) SetWriter(writer io.Writer) {
	r.writer = writer
}

func (r *ResponseV5) close() error {
	if wr, ok := r.writer.(io.Closer); ok {
		return wr.Close()
	}
	return nil
}

// --------------------------------------
// Accessors

// GetStatus return the response code for this request or 0 if not yet set.
func (r *ResponseV5) GetStatus() int {
	return r.status
}

// IsEmpty return true if nothing has been written to the response body yet.
func (r *ResponseV5) IsEmpty() bool {
	return r.empty
}

// IsHeaderWritten return true if the response header has been written.
// Once the response header is written, you cannot change the response status
// and headers anymore.
func (r *ResponseV5) IsHeaderWritten() bool {
	return r.wroteHeader
}

// --------------------------------------
// Write methods

// Status set the response status code.
// Calling this method a second time will have no effect.
func (r *ResponseV5) Status(status int) {
	if r.status == 0 {
		r.status = status
	}
}

// JSON write json data as a response.
// Also sets the "Content-Type" header automatically.
func (r *ResponseV5) JSON(responseCode int, data interface{}) {
	r.responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	r.status = responseCode
	if err := json.NewEncoder(r).Encode(data); err != nil {
		panic(err)
	}
}

// String write a string as a response
func (r *ResponseV5) String(responseCode int, message string) {
	r.status = responseCode
	if _, err := r.Write([]byte(message)); err != nil {
		panic(err)
	}
}

func (r *ResponseV5) writeFile(file string, disposition string) { // TODO handle io.FS
	if !fsutil.FileExists(file) {
		r.Status(http.StatusNotFound)
		return
	}
	r.empty = false
	r.status = http.StatusOK
	mime, size := fsutil.GetMIMEType(file)
	header := r.responseWriter.Header()
	header.Set("Content-Disposition", disposition)

	if header.Get("Content-Type") == "" {
		header.Set("Content-Type", mime)
	}

	header.Set("Content-Length", strconv.FormatInt(size, 10))

	f, _ := os.Open(file)
	// No need to check for errors, fsutil.FileExists(file) and
	// fsutil.GetMIMEType(file) already handled that.
	defer func() {
		_ = f.Close()
	}()
	if _, err := io.Copy(r, f); err != nil {
		panic(err)
	}
}

// File write a file as an inline element.
// Automatically detects the file MIME type and sets the "Content-Type" header accordingly.
// If the file doesn't exist, respond with status 404 Not Found.
// The given path can be relative or absolute.
//
// If you want the file to be sent as a download ("Content-Disposition: attachment"), use the "Download" function instead.
func (r *ResponseV5) File(file string) {
	r.writeFile(file, "inline")
}

// Download write a file as an attachment element.
// Automatically detects the file MIME type and sets the "Content-Type" header accordingly.
// If the file doesn't exist, respond with status 404 Not Found.
// The given path can be relative or absolute.
//
// The "fileName" parameter defines the name the client will see. In other words, it sets the header "Content-Disposition" to
// "attachment; filename="${fileName}""
//
// If you want the file to be sent as an inline element ("Content-Disposition: inline"), use the "File" function instead.
func (r *ResponseV5) Download(file string, fileName string) {
	r.writeFile(file, fmt.Sprintf("attachment; filename=\"%s\"", fileName))
}
