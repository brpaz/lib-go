package health

import (
	"context"
)

// Status constants represent the possible statuses of a health check.
const (
	StatusPass = "pass"
	StatusWarn = "warn"
	StatusFail = "fail"
)

// Checker is the interface that must be implemented by every health check.
type Checker interface {
	GetName() string
	Check(ctx context.Context) CheckResult
}

// Struct that represents the result of a health check.
// All checks must return an instance of this struct.
type CheckResult struct {
	Status  string         `json:"status"`
	Message string         `json:"message,omitempty"`
	Error   error          `json:"-"`
	Details map[string]any `json:"details,omitempty"`
}

// HealthProcessor is an interface that should be implemented by the main health check service.
type HealthProcessor interface {
	Execute(ctx context.Context) HealthResult
}

// HealthResult is the root response object for the health check endpoint.
// It aggregates the results of all available health checks and sets the overall status of the service.
type HealthResult struct {
	Service     string                 `json:"service"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Commit      string                 `json:"commit"`
	Status      string                 `json:"status"`
	Message     string                 `json:"message,omitempty"`
	Checks      map[string]CheckResult `json:"checks,omitempty"`
	Timestamp   int64                  `json:"timestamp"`
}
