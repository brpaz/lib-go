package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/log"
	"github.com/brpaz/lib-go/log/middleware"
)

func TestLoggableResponseWriter_WriteHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	lrw := &middleware.LoggableResponseWriter{ResponseWriter: recorder}

	lrw.WriteHeader(http.StatusCreated)
	assert.Equal(t, http.StatusCreated, lrw.StatusCode)
}

func TestLoggableResponseWriter_Write(t *testing.T) {
	recorder := httptest.NewRecorder()
	lrw := &middleware.LoggableResponseWriter{ResponseWriter: recorder}
	data := []byte("Hello, World!")

	n, err := lrw.Write(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, len(data), lrw.ResponseSize)
	assert.True(t, lrw.Written)
	assert.Equal(t, "Hello, World!", recorder.Body.String())
}

func TestRequestLogger_UsesDefaultConfig_OnNilConfig(t *testing.T) {
	logger := log.NewInMemory(log.LevelInfo)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	requestLogger := middleware.RequestLogger(logger, nil)
	server := httptest.NewServer(requestLogger(handler))
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)

	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	logs := logger.Entries()
	assert.Len(t, logs, 1)

	logEntry := logs[0]
	assert.NotEmpty(t, logEntry.Timestamp)
	assert.Equal(t, logEntry.Message, "incoming request")
	assert.Len(t, logEntry.Fields, 8)

	expectedFields := []string{
		"method",
		"path",
		"status",
		"duration",
		"request_size",
		"response_size",
		"remote_addr",
		"user_agent",
	}

	for _, fieldName := range expectedFields {
		_, ok := logEntry.GetField(fieldName)
		assert.True(t, ok, "Field '%s' not found in log entry", fieldName)
	}
}

func TestRequestLogger_WithCustomConfig(t *testing.T) {
	logger := log.NewInMemory(log.LevelInfo)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	config := &middleware.RequestLoggerConfig{
		LogMethod:   true,
		LogPath:     true,
		LogStatus:   true,
		LogDuration: false,
	}

	requestLogger := middleware.RequestLogger(logger, config)
	server := httptest.NewServer(requestLogger(handler))
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)

	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	logs := logger.Entries()
	assert.Len(t, logs, 1)

	logEntry := logs[0]

	assert.Len(t, logEntry.Fields, 3)

	expectedFields := []string{
		"method",
		"path",
		"status",
	}

	for _, fieldName := range expectedFields {
		_, ok := logEntry.GetField(fieldName)
		assert.True(t, ok, "Field '%s' not found in log entry", fieldName)
	}
}

func TestRequestLogger_DefaultRequestLoggerConfig(t *testing.T) {
	config := middleware.DefaultRequestLoggerConfig()
	assert.NotNil(t, config)
}

func TestRequestLogger_WithFullRequestFields(t *testing.T) {
	logger := log.NewInMemory(log.LevelInfo)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	loggerConfig := &middleware.RequestLoggerConfig{
		LogRequestHeaders:     true,
		LogRequestQueryParams: true,
		LogRequestBody:        true,
	}

	requestLogger := middleware.RequestLogger(logger, loggerConfig)
	server := httptest.NewServer(requestLogger(handler))
	defer server.Close()

	reqBody := []byte(`{"test": "test"}`)

	req, _ := http.NewRequest(http.MethodGet, server.URL, bytes.NewReader(reqBody))

	req.URL.RawQuery = "test=test"
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Test-Header", "test-header")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	logs := logger.Entries()
	assert.Len(t, logs, 1)

	logEntry := logs[0]

	expectedHeaders := map[string]string{
		"header_user-agent":    "test-agent",
		"header_x-test-header": "test-header",
	}

	for key, value := range expectedHeaders {
		field, ok := logEntry.GetField(key)
		assert.True(t, ok, "Field '%s' not found in log entry", key)
		assert.Equal(t, value, field.String)
	}

	queryParamField, ok := logEntry.GetField("query")
	assert.True(t, ok, "Field 'query' not found in log entry")
	assert.Equal(t, "test=test", queryParamField.String)

	requestBodyField, ok := logEntry.GetField("request_body")

	assert.True(t, ok, "Field 'request_body' not found in log entry")
	assert.Equal(t, `{"test": "test"}`, requestBodyField.String)
}
