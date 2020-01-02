package goyave

import (
	"encoding/json"
	"fmt"
	htmltemplate "html/template"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"text/template"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/helper/filesystem"
)

// Response represents a controller response.
type Response struct {
	// Used to check if controller didn't write anything so
	// core can write default 204 No Content.
	// See RFC 7231, 6.3.5
	empty       bool
	status      int
	wroteHeader bool
	err         interface{}

	httpRequest *http.Request
	http.ResponseWriter
}

// --------------------------------------
// http.ResponseWriter implementation

// Write writes the data as a response.
// See http.ResponseWriter.Write
func (r *Response) Write(data []byte) (int, error) {
	r.empty = false
	if !r.wroteHeader {
		if r.status == 0 {
			r.status = 200
		}
		r.WriteHeader(r.status)
	}
	return r.ResponseWriter.Write(data)
}

// WriteHeader sends an HTTP response header with the provided
// status code.
// Prefer using "Status()" method instead.
// Calling this method a second time will have no effect.
func (r *Response) WriteHeader(status int) {
	if !r.wroteHeader {
		r.status = status
		r.wroteHeader = true
		r.ResponseWriter.WriteHeader(status)
	}
}

// Header returns the header map that will be sent.
func (r *Response) Header() http.Header {
	return r.ResponseWriter.Header()
}

// --------------------------------------

// GetStatus return the response code for this request or 0 if not yet set.
func (r *Response) GetStatus() int {
	return r.status
}

// GetError return the value which caused a panic in the request's handling, or nil.
func (r *Response) GetError() interface{} {
	return r.err
}

// Status set the response status code.
// Calling this method a second time will have no effect.
func (r *Response) Status(status int) {
	if r.status == 0 {
		r.status = status
	}
}

// JSON write json data as a response.
// Also sets the "Content-Type" header automatically
func (r *Response) JSON(responseCode int, data interface{}) error {
	r.ResponseWriter.Header().Set("Content-Type", "application/json")
	r.WriteHeader(responseCode)
	return json.NewEncoder(r).Encode(data)
}

// String write a string as a response
func (r *Response) String(responseCode int, message string) error {
	r.WriteHeader(responseCode)
	_, err := r.Write([]byte(message))
	return err
}

func (r *Response) writeFile(file string, disposition string) (int64, error) {
	r.empty = false
	r.status = http.StatusOK
	mime, size := filesystem.GetMIMEType(file)
	r.ResponseWriter.Header().Set("Content-Disposition", disposition)
	r.ResponseWriter.Header().Set("Content-Type", mime)
	r.ResponseWriter.Header().Set("Content-Length", strconv.FormatInt(size, 10))

	f, _ := os.Open(file)
	defer f.Close()
	return io.Copy(r.ResponseWriter, f)
}

// File write a file as an inline element.
// Automatically detects the file MIME type and sets the "Content-Type" header accordingly.
// It is advised to call "filesystem.FileExists()" before sending a file to avoid a panic and return a 404 error
// if the file doesn't exist.
// The given path can be relative or absolute.
//
// If you want the file to be sent as a download ("Content-Disposition: attachment"), use the "Download" function instead.
func (r *Response) File(file string) error {
	_, err := r.writeFile(file, "inline")
	return err
}

// Download write a file as an attachment element.
// Automatically detects the file MIME type and sets the "Content-Type" header accordingly.
// It is advised to call "filesystem.FileExists()" before sending a file to avoid a panic and return a 404 error
// if the file doesn't exist.
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

// Error print the error in the console and return it with an error code 500.
// If debugging is enabled in the config, the error is also written in the response
// and the stacktrace is printed in the console.
func (r *Response) Error(err interface{}) error {
	dbg := config.GetBool("debug")
	log.Println(err)
	if dbg {
		debug.PrintStack()
		var message interface{}
		if e, ok := err.(error); ok {
			message = e.Error()
		} else {
			message = err
		}
		return r.JSON(http.StatusInternalServerError, map[string]interface{}{"error": message})
	}

	r.empty = false
	r.WriteHeader(http.StatusInternalServerError)
	return nil
}

// Cookie add a Set-Cookie header to the response.
// The provided cookie must have a valid Name. Invalid cookies may be
// silently dropped.
func (r *Response) Cookie(cookie *http.Cookie) {
	http.SetCookie(r.ResponseWriter, cookie)
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
	r.WriteHeader(responseCode)
	tmplt, err := template.ParseFiles(r.getTemplateDirectory() + templatePath)
	if err != nil {
		return err
	}
	return tmplt.Execute(r, data)
}

// RenderHTML an HTML template with the given data.
// The template path is relative to the "resources/template" directory.
func (r *Response) RenderHTML(responseCode int, templatePath string, data interface{}) error {
	r.WriteHeader(responseCode)
	tmplt, err := htmltemplate.ParseFiles(r.getTemplateDirectory() + templatePath)
	if err != nil {
		return err
	}
	return tmplt.Execute(r, data)
}

func (r *Response) getTemplateDirectory() string {
	sep := string(os.PathSeparator)
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return workingDir + sep + "resources" + sep + "template" + sep
}

// CreateTestResponse create an empty response with the given response writer.
// This function is aimed at making it easier to unit test Responses.
//
// Deprecated: Use goyave.TestSuite.CreateTestResponse instead.
//
//  writer := httptest.NewRecorder()
//  response := goyave.CreateTestResponse(writer)
//  response.Status(http.StatusNoContent)
//  result := writer.Result()
//  fmt.Println(result.StatusCode) // 204
func CreateTestResponse(recorder http.ResponseWriter) *Response {
	return &Response{
		ResponseWriter: recorder,
		empty:          true,
	}
}
