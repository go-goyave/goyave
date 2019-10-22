package response

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/System-Glitch/goyave/config"
)

// JSON write json data as a response.
// Also sets the "Content-Type" header automatically
func JSON(w http.ResponseWriter, responseCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	json.NewEncoder(w).Encode(data)
}

// Error log the error and return it as error code 500
// If debugging is enabled in the config, the error is also written in the response.
func Error(w http.ResponseWriter, err interface{}) {
	dbg := config.Get("debug").(bool)
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
