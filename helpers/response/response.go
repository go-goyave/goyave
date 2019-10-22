package response

import (
	"encoding/json"
	"net/http"
)

// JSON write json data as a response.
// Also sets the "Content-Type" header automatically
func JSON(w http.ResponseWriter, responseCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	json.NewEncoder(w).Encode(data)
}
