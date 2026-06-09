package httpapi

import (
	"encoding/json"
	"net/http"
)

// Response writes a JSON response with the given status code.
func Response(w http.ResponseWriter, statusCode int, data any) {
	writeJSON(w, statusCode, data)
}

// Ok writes a 200 JSON response.
func Ok(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, data)
}

// Created writes a 201 JSON response.
func Created(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusCreated, data)
}

// Accepted writes a 202 JSON response.
func Accepted(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusAccepted, data)
}

// Error writes a JSON error response with the given status code, error code, and message.
func Error(w http.ResponseWriter, statusCode int, code, message string) {
	writeJSON(w, statusCode, ErrorResponse{Code: code, Message: message})
}

// BadRequest writes a 400 JSON response.
func BadRequest(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusBadRequest, data)
}

// Unauthorized writes a 401 JSON response.
func Unauthorized(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusUnauthorized, data)
}

// Forbidden writes a 403 JSON response.
func Forbidden(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusForbidden, data)
}

// NotFound writes a 404 JSON response.
func NotFound(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusNotFound, data)
}

// TooManyRequests writes a 429 JSON response.
func TooManyRequests(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusTooManyRequests, data)
}

// InternalServerError writes a 500 JSON response.
func InternalServerError(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusInternalServerError, data)
}

// Conflict writes a 409 JSON response.
func Conflict(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusConflict, data)
}

// UnprocessableEntity writes a 422 JSON response.
func UnprocessableEntity(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusUnprocessableEntity, data)
}

// NoContent writes a 204 response with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
