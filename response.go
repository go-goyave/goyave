package goyave

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"io"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"text/template"

	"gorm.io/gorm"
	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/util/fsutil"
)

var (
	// ErrNotHijackable returned by response.Hijack() if the underlying
	// http.ResponseWriter doesn't implement http.Hijacker. This can
	// happen with HTTP/2 connections.
	ErrNotHijackable = errors.New("Underlying http.ResponseWriter doesn't implement http.Hijacker")
)

// PreWriter is a writter that needs to alter the response headers or status
// before they are written.
// If implemented, PreWrite will be called right before the Write operation.
type PreWriter interface {
	PreWrite(b []byte)
}

// Response represents a controller response.
type Response struct {
	writer         io.Writer
	responseWriter http.ResponseWriter
	err            interface{}
	httpRequest    *http.Request
	stacktrace     string
	status         int

	// Used to check if controller didn't write anything so
	// core can write default 204 No Content.
	// See RFC 7231, 6.3.5
	empty       bool
	wroteHeader bool
	hijacked    bool
}

// newResponse create a new Response using the given http.ResponseWriter and raw request.
func newResponse(writer http.ResponseWriter, rawRequest *http.Request) *Response {
	return &Response{
		responseWriter: writer,
		writer:         writer,
		httpRequest:    rawRequest,
		empty:          true,
		status:         0,
		wroteHeader:    false,
		err:            nil,
	}
}

// --------------------------------------
// PreWriter implementation

// PreWrite writes the response header after calling PreWrite on the
// child writer if it implements PreWriter.
func (r *Response) PreWrite(b []byte) {
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
func (r *Response) Write(data []byte) (int, error) {
	r.PreWrite(data)
	return r.writer.Write(data)
}

// WriteHeader sends an HTTP response header with the provided
// status code.
// Prefer using "Status()" method instead.
// Calling this method a second time will have no effect.
func (r *Response) WriteHeader(status int) {
	if !r.wroteHeader {
		r.status = status
		r.wroteHeader = true
		r.responseWriter.WriteHeader(status)
	}
}

// Header returns the header map that will be sent.
func (r *Response) Header() http.Header {
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
func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
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
func (r *Response) Hijacked() bool {
	return r.hijacked
}

// --------------------------------------
// Chained writers

// Writer return the current writer used to write the response.
// Note that the returned writer is not necessarily a http.ResponseWriter, as
// it can be replaced using SetWriter.
func (r *Response) Writer() io.Writer {
	return r.writer
}

// SetWriter set the writer used to write the response.
// This can be used to chain writers, for example to enable
// gzip compression, or for logging.
//
// The original http.ResponseWriter is always kept.
func (r *Response) SetWriter(writer io.Writer) {
	r.writer = writer
}

func (r *Response) close() error {
	if wr, ok := r.writer.(io.Closer); ok {
		return wr.Close()
	}
	return nil
}

// --------------------------------------
// Accessors

// GetStatus return the response code for this request or 0 if not yet set.
func (r *Response) GetStatus() int {
	return r.status
}

// GetError return the value which caused a panic in the request's handling, or nil.
func (r *Response) GetError() interface{} {
	return r.err
}

// GetStacktrace return the stacktrace of when the error occurred, or an empty string.
// The stacktrace is captured by the recovery middleware.
func (r *Response) GetStacktrace() string {
	return r.stacktrace
}

// IsEmpty return true if nothing has been written to the response body yet.
func (r *Response) IsEmpty() bool {
	return r.empty
}

// IsHeaderWritten return true if the response header has been written.
// Once the response header is written, you cannot change the response status
// and headers anymore.
func (r *Response) IsHeaderWritten() bool {
	return r.wroteHeader
}

// --------------------------------------
// Write methods

// Status set the response status code.
// Calling this method a second time will have no effect.
func (r *Response) Status(status int) {
	if r.status == 0 {
		r.status = status
	}
}

// JSON write json data as a response.
// Also sets the "Content-Type" header automatically.
func (r *Response) JSON(responseCode int, data interface{}) error {
	r.responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	r.status = responseCode
	return json.NewEncoder(r).Encode(data)
}

// String write a string as a response
func (r *Response) String(responseCode int, message string) error {
	r.status = responseCode
	_, err := r.Write([]byte(message))
	return err
}

func (r *Response) writeFile(file string, disposition string) (int64, error) {
	if !fsutil.FileExists(file) {
		r.Status(http.StatusNotFound)
		return 0, &os.PathError{Op: "open", Path: file, Err: fmt.Errorf("no such file or directory")}
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
	defer f.Close()
	return io.Copy(r, f)
}

// File write a file as an inline element.
// Automatically detects the file MIME type and sets the "Content-Type" header accordingly.
// If the file doesn't exist, respond with status 404 Not Found.
// The given path can be relative or absolute.
//
// If you want the file to be sent as a download ("Content-Disposition: attachment"), use the "Download" function instead.
func (r *Response) File(file string) error {
	_, err := r.writeFile(file, "inline")
	return err
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
func (r *Response) Download(file string, fileName string) error {
	_, err := r.writeFile(file, fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	return err
}

// Error print the error in the console and return it with an error code 500 (or previously defined
// status code using `response.Status()`).
// If debugging is enabled in the config, the error is also written in the response
// and the stacktrace is printed in the console.
// If debugging is not enabled, only the status code is set, which means you can still
// write to the response, or use your error status handler.
func (r *Response) Error(err interface{}) error {
	ErrLogger.Println(err)
	return r.error(err)
}

func (r *Response) error(err interface{}) error {
	r.err = err
	if config.GetBool("app.debug") {
		stacktrace := r.stacktrace
		if stacktrace == "" {
			stacktrace = string(debug.Stack())
		}
		ErrLogger.Print(stacktrace)
		if !r.Hijacked() {
			var message interface{}
			if e, ok := err.(error); ok {
				message = e.Error()
			} else {
				message = err
			}
			status := http.StatusInternalServerError
			if r.status != 0 {
				status = r.status
			}
			return r.JSON(status, map[string]interface{}{"error": message})
		}
	}

	// Don't set r.empty to false to let error status handler process the error
	r.Status(http.StatusInternalServerError)
	return nil
}

// Cookie add a Set-Cookie header to the response.
// The provided cookie must have a valid Name. Invalid cookies may be
// silently dropped.
func (r *Response) Cookie(cookie *http.Cookie) {
	http.SetCookie(r.responseWriter, cookie)
}

// Redirect send a permanent redirect response
func (r *Response) Redirect(url string) {
	http.Redirect(r, r.httpRequest, url, http.StatusPermanentRedirect)
}

// TemporaryRedirect send a temporary redirect response
func (r *Response) TemporaryRedirect(url string) {
	http.Redirect(r, r.httpRequest, url, http.StatusTemporaryRedirect)
}

// Render a text template with the given data.
// The template path is relative to the "resources/template" directory.
func (r *Response) Render(responseCode int, templatePath string, data interface{}) error {
	tmplt, err := template.ParseFiles(r.getTemplateDirectory() + templatePath)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	if err := tmplt.Execute(&b, data); err != nil {
		return err
	}

	return r.String(responseCode, b.String())
}

// RenderHTML an HTML template with the given data.
// The template path is relative to the "resources/template" directory.
func (r *Response) RenderHTML(responseCode int, templatePath string, data interface{}) error {
	tmplt, err := htmltemplate.ParseFiles(r.getTemplateDirectory() + templatePath)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	if err := tmplt.Execute(&b, data); err != nil {
		return err
	}

	return r.String(responseCode, b.String())
}

func (r *Response) getTemplateDirectory() string {
	sep := string(os.PathSeparator)
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return workingDir + sep + "resources" + sep + "template" + sep
}

// HandleDatabaseError takes a database query result and checks if any error has occurred.
//
// Automatically writes HTTP status code 404 Not Found if the error is a "Not found" error.
// Calls "Response.Error()" if there is another type of error.
//
// Returns true if there is no error.
func (r *Response) HandleDatabaseError(db *gorm.DB) bool {
	if db.Error != nil {
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			r.Status(http.StatusNotFound)
		} else {
			r.Error(db.Error)
		}
		return false
	}
	return true
}
