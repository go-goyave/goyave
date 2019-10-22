package response

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
	"github.com/System-Glitch/goyave/helpers"
)

// Status write the given status code
func Status(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}

// JSON write json data as a response.
// Also sets the "Content-Type" header automatically
func JSON(w http.ResponseWriter, responseCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	json.NewEncoder(w).Encode(data)
}

// String write a string as a response
func String(w http.ResponseWriter, responseCode int, message string) {
	w.WriteHeader(responseCode)
	w.Write([]byte(message))
}

func writeFile(w http.ResponseWriter, file string, disposition string) {
	mime, size, err := helpers.GetMimeType(file)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Disposition", disposition)
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))

	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	io.Copy(w, f)
}

// File write a file as an inline element.
// Automatically detects the file MIME type and sets the "Content-Type" header accordingly.
// If you want the file to be sent as a download ("Content-Disposition: attachment"), use the "Download" function instead
func File(w http.ResponseWriter, file string) {
	writeFile(w, file, "inline")
}

// Download write a file as an attachment element.
// Automatically detects the file MIME type and sets the "Content-Type" header accordingly.
// If you want the file to be sent as an inline element ("Content-Disposition: inline"), use the "File" function instead
//
// The "fileName" parameter defines the name the client will see. In other words, it sets the header "Content-Disposition" to
// "attachment; filename="${fileName}""
func Download(w http.ResponseWriter, file string, fileName string) {
	writeFile(w, file, fmt.Sprintf("attachment; filename=\"%s\"", fileName))
}

// Error log the error and return it as error code 500
// If debugging is enabled in the config, the error is also written in the response.
func Error(w http.ResponseWriter, err interface{}) {
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
		JSON(w, http.StatusInternalServerError, map[string]interface{}{"error": message})
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
