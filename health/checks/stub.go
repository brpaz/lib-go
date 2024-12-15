package checks

import (
	"context"
	"errors"

	"github.com/brpaz/lib-go/health"
)

// StubCheck is a stub health check that returns a fixed result
type StubCheck struct {
	name      string
	retResult bool
}

// NewStubCheck creates a new StubCheck instance with the provided result
func NewStubCheck(name string, result bool) *StubCheck {
	return &StubCheck{
		name:      name,
		retResult: result,
	}
}

// GetName returns the name of the check
func (s *StubCheck) GetName() string {
	return s.name
}

// Check returns an error if the Result field is false
func (s *StubCheck) Check(ctx context.Context) health.CheckResult {
	if !s.retResult {
		return health.CheckResult{
			Status:  health.StatusFail,
			Error:   errors.New("stub check failed"),
			Message: "Stub check failed",
		}
	}

	return health.CheckResult{
		Status: health.StatusPass,
	}
}
