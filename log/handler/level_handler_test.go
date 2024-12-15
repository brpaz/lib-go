package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/lib-go/log"
	h "github.com/brpaz/lib-go/log/handler"
)

// Helper function to create a new HTTP request for the log level handler
func createLogLevelRequest(level string) (*http.Request, *httptest.ResponseRecorder, *log.InMemoryLogger) {
	logger := log.NewInMemory(log.LevelInfo)

	reqBody := h.SetLogLevelRequest{Level: level}
	body, err := json.Marshal(reqBody)
	if err != nil {
		panic("Failed to marshal request body") // Fail fast if marshalling fails
	}

	req, err := http.NewRequest(http.MethodPost, "/set-log-level", bytes.NewReader(body))
	if err != nil {
		panic("Failed to create request") // Fail fast if creating the request fails
	}

	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	return req, rr, logger
}

func TestLogLevelHandler(t *testing.T) {
	t.Run("With Valid Log Level", func(t *testing.T) {
		req, rr, logger := createLogLevelRequest("error")

		handler := h.LevelHandler(logger)
		handler.ServeHTTP(rr, req)

		// Assert the expected status code and log level
		assert.Equal(t, http.StatusNoContent, rr.Code)
		assert.Equal(t, log.LevelError, logger.GetLevel())
	})

	t.Run("With Invalid Log Level", func(t *testing.T) {
		req, rr, logger := createLogLevelRequest("invalid")

		handler := h.LevelHandler(logger)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, log.LevelInfo, logger.GetLevel())
	})

	t.Run("With Unsupported Media Type", func(t *testing.T) {
		// Simulate a request with an unsupported media type
		logger := log.NewInMemory(log.LevelInfo)
		req, _ := http.NewRequest(http.MethodPost, "/set-log-level", nil)
		rr := httptest.NewRecorder()

		handler := h.LevelHandler(logger)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnsupportedMediaType, rr.Code)
	})
	t.Run("Malformed JSON", func(t *testing.T) {
		// Simulate a malformed JSON request
		logger := log.NewInMemory(log.LevelInfo)
		req, _ := http.NewRequest(http.MethodPost, "/set-log-level", bytes.NewReader([]byte("{bad json}")))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := h.LevelHandler(logger)
		handler.ServeHTTP(rr, req)

		// Assert the expected error response and log level
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, log.LevelInfo, logger.GetLevel())
	})
}
