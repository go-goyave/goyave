package goyave

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/System-Glitch/goyave/config"
	"github.com/System-Glitch/goyave/helpers/filesystem"
)

// Response represents a controller response.
type Response struct {
	// Used to check if controller didn't write anything so
	// core can write default 204 No Content.
	// See RFC 7231, 6.3.5
	empty bool

	httpRequest *http.Request
	writer      http.ResponseWriter
}

// Header returns the header map that will be sent.
func (r *Response) Header() http.Header {
	return r.writer.Header()
}

// Status write the given status code
func (r *Response) Status(status int) {
	r.empty = false
	r.writer.WriteHeader(status)
}

// JSON write json data as a response.
// Also sets the "Content-Type" header automatically
func (r *Response) JSON(responseCode int, data interface{}) {
	r.empty = false
	r.writer.Header().Set("Content-Type", "application/json")
	r.writer.WriteHeader(responseCode)
	json.NewEncoder(r.writer).Encode(data)
}

// String write a string as a response
func (r *Response) String(responseCode int, message string) {
	r.empty = false
	r.writer.WriteHeader(responseCode)
	r.writer.Write([]byte(message))
}

func (r *Response) writeFile(file string, disposition string) {
	r.empty = false
	mime, size := filesystem.GetMimeType(file)
	r.writer.Header().Set("Content-Disposition", disposition)
	r.writer.Header().Set("Content-Type", mime)
	r.writer.Header().Set("Content-Length", strconv.FormatInt(size, 10))

	f, _ := os.Open(file)
	defer f.Close()
	io.Copy(r.writer, f)
}

// File write a file as an inline element.
// Automatically detects the file MIME type and sets the "Content-Type" header accordingly.
// It is advised to call "filesystem.FileExists()" before sending a file to avoid a panic and return a 404 error
// if the file doesn't exist.
// The given path can be relative or absolute.
//
// If you want the file to be sent as a download ("Content-Disposition: attachment"), use the "Download" function instead.
func (r *Response) File(file string) {
	r.writeFile(file, "inline")
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
func (r *Response) Download(file string, fileName string) {
	r.writeFile(file, fmt.Sprintf("attachment; filename=\"%s\"", fileName))
}

// Error print the error in the console and return it with an error code 500.
// If debugging is enabled in the config, the error is also written in the response
// and the stacktrace is printed in the console.
func (r *Response) Error(err interface{}) {
	r.empty = false
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
		r.JSON(http.StatusInternalServerError, map[string]interface{}{"error": message})
	} else {
		r.writer.WriteHeader(http.StatusInternalServerError)
	}
}

// Cookie add a Set-Cookie header to the response.
// The provided cookie must have a valid Name. Invalid cookies may be
// silently dropped.
func (r *Response) Cookie(cookie *http.Cookie) {
	http.SetCookie(r.writer, cookie)
}

// Redirect send a permanent redirect response
func (r *Response) Redirect(url string) {
	r.empty = false
	http.Redirect(r.writer, r.httpRequest, url, http.StatusPermanentRedirect)
}

// TemporaryRedirect send a temporary redirect response
func (r *Response) TemporaryRedirect(url string) {
	r.empty = false
	http.Redirect(r.writer, r.httpRequest, url, http.StatusTemporaryRedirect)
}

// CreateTestResponse create an empty response with the given response writer.
// This function is aimed at making it easier to unit test Responses.
//
//  writer := httptest.NewRecorder()
//  response := goyave.CreateTestResponse(writer)
//  response.Status(http.StatusNoContent)
//  result := writer.Result()
//  fmt.Println(result.StatusCode) // 204
func CreateTestResponse(recorder http.ResponseWriter) *Response {
	return &Response{
		writer: recorder,
		empty:  true,
	}
}
