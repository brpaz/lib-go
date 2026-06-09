package httpapi

// Location identifies where in an HTTP request a field error originated.
// Values mirror the OpenAPI "in" vocabulary.
type Location string

const (
	LocationBody   Location = "body"
	LocationQuery  Location = "query"
	LocationPath   Location = "path"
	LocationHeader Location = "header"
)

// ErrorResponse is the base error response body shape.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}

// FieldErrorResponse describes a single field-level validation error in an HTTP response.
type FieldErrorResponse struct {
	Location string         `json:"location,omitempty"`
	Field    string         `json:"field"`
	Code     string         `json:"code"`
	Message  string         `json:"message"`
	Params   map[string]any `json:"params,omitempty"`
}

// ValidationErrorResponse is used for 422 responses; Errors is always a non-nil array.
type ValidationErrorResponse struct {
	ErrorResponse
	Errors []FieldErrorResponse `json:"errors"`
}
