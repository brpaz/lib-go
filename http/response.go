package http

import (
	"encoding/json"
	"net/http"
)

// Ok writes a 200 OK response with the given body.
func Ok(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")

	// Write the response body
	err := json.NewEncoder(w).Encode(body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// NoContent writes a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
