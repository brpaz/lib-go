package health

import (
	"context"
	"sync"

	"github.com/brpaz/lib-go/timeutil"
)

// Service is the main health check service.
// It implements the HealthProcessor interface and is responsible for executing all health checks registered in the service.
// Example Usage:
//
//	hs := health.New(
//	    health.WithName("my-service"),
//	    health.WithVersion("1.0.0"),
//	    health.WithRevision("abc123"),
//	    health.WithChecks(
//	        &checks.StubCheck{Result: true},
//	        &checks.StubCheck{Result: false},
//	    ),
//	)
//
//	healthResult := hs.Execute(context.Background())
type Service struct {
	Name        string
	Description string
	Version     string
	Revision    string
	Checks      []Checker
	Clock       timeutil.Clock
}

// Option is a function that configures the HealthService.
type Option func(*Service)

// checkRun is a struct that represents the result of a health check, associated with its name.
type checkRun struct {
	Name   string
	Result CheckResult
}

// New creates a new HealthService instance with the provided options.
func New(options ...Option) *Service {
	hs := &Service{
		Clock:  timeutil.NewRealClock(),
		Checks: make([]Checker, 0),
	}

	for _, opt := range options {
		opt(hs)
	}

	return hs
}

// WithName sets the service name in the HealthService.
func WithName(name string) Option {
	return func(hs *Service) {
		hs.Name = name
	}
}

func WithDescription(description string) Option {
	return func(hs *Service) {
		hs.Description = description
	}
}

// WithVersion sets the version in the HealthService.
func WithVersion(version string) Option {
	return func(hs *Service) {
		hs.Version = version
	}
}

// WithRevision sets the revision in the HealthService.
func WithRevision(revision string) Option {
	return func(hs *Service) {
		hs.Revision = revision
	}
}

// WithChecks adds health checks to the HealthService.
func WithChecks(checks ...Checker) Option {
	return func(hs *Service) {
		hs.Checks = append(hs.Checks, checks...)
	}
}

// WithClock sets the clock to be used by the HealthService.
func WithClock(clock timeutil.Clock) Option {
	return func(hs *Service) {
		hs.Clock = clock
	}
}

// AddCheck adds a new health check to the HealthService.
func (hs *Service) AddCheck(check Checker) {
	hs.Checks = append(hs.Checks, check)
}

// Execute runs all registered health checks and returns the aggregated result.
func (hs *Service) Execute(ctx context.Context) HealthResult {
	result := HealthResult{
		Service:     hs.Name,
		Description: hs.Description,
		Version:     hs.Version,
		Commit:      hs.Revision,
		Status:      StatusPass,
		Timestamp:   hs.Clock.Now().Unix(),
		Checks:      make(map[string]CheckResult, len(hs.Checks)),
	}

	var wg sync.WaitGroup
	checkRuns := make(chan checkRun, len(hs.Checks))

	// Run all checks concurrently and send the results to the checkRuns channel
	for _, check := range hs.Checks {
		wg.Add(1)
		go func(c Checker) {
			defer wg.Done()
			result := c.Check(ctx)
			checkRuns <- checkRun{
				Name:   c.GetName(),
				Result: result,
			}
		}(check)
	}

	// Wait for all goroutines to finish and close the channel
	wg.Wait()
	close(checkRuns)

	// Aggregate results
	for checkRun := range checkRuns {
		result.Checks[checkRun.Name] = checkRun.Result

		if checkRun.Result.Status == StatusFail {
			result.Status = StatusFail
		}
	}

	return result
}
