package httputil

import (
	"encoding/json"
	"net/http"
)

// JSON writes a JSON response with the given status code and body.
func JSON(w http.ResponseWriter, statusCode int, body any) {
	w.Header().Set("Content-Type", "application/json")

	if body != nil {
		err := json.NewEncoder(w).Encode(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Set the status code
	w.WriteHeader(statusCode)
}

func HandleError(w http.ResponseWriter, err error) {
	// TODO: Implement
}
