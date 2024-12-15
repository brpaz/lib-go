package handler

import (
	"encoding/json"
	"net/http"

	"github.com/brpaz/lib-go/log"
)

// SetLogLevelRequest is the structure that will be used to parse the incoming JSON request.
type SetLogLevelRequest struct {
	Level string `json:"level"`
}

// LevelHandler is an HTTP handler that dynamically changes the log level of the logger.
func LevelHandler(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Ensure the content type is application/json
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		// Parse the request body to get the log level
		var req SetLogLevelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Log the request to change log level
		logger.Info(ctx, "Received log level change request", log.String("level", req.Level))

		// Validate the log level string and convert it to a Level type
		level, err := log.LevelFromString(req.Level)
		if err != nil {
			http.Error(w, "Invalid log level: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Attempt to set the new log level
		if err := logger.SetLevel(level); err != nil {
			http.Error(w, "Failed to set log level: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Log the success
		logger.Info(ctx, "Log level changed successfully", log.String("level", req.Level))

		// Respond with no content, since there's no response body
		w.WriteHeader(http.StatusNoContent)
	}
}
