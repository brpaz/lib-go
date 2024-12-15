package checks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"time"

	"github.com/brpaz/lib-go/health"
)

// URLCheck is a health check that verifies the availability of a URL
type URLCheck struct {
	name             string
	targetURL        url.URL
	timeout          time.Duration
	httpClient       *http.Client
	validStatusCodes []int
}

// Option is a function that configures the HealthService.
type UrlCheckOption func(*URLCheck)

// WithURLCheckTimeout sets the timeout for the URLCheck
func WithURLCheckTimeout(timeout time.Duration) UrlCheckOption {
	return func(c *URLCheck) {
		c.timeout = timeout
	}
}

// WithURLCheckURL sets the target URL for the URLCheck
func WithURLCheckURL(url url.URL) UrlCheckOption {
	return func(c *URLCheck) {
		c.targetURL = url
	}
}

// WithURLCheckHTTPClient sets the HTTP client to be used by the URLCheck
func WithURLCheckHTTPClient(client *http.Client) UrlCheckOption {
	return func(c *URLCheck) {
		c.httpClient = client
	}
}

func WithURLCheckValidStatusCodes(codes []int) UrlCheckOption {
	return func(c *URLCheck) {
		c.validStatusCodes = codes
	}
}

// NewURLCheck creates a new URLCheck instance with the provided parameters
func NewURLCheck(name string, opts ...UrlCheckOption) (*URLCheck, error) {
	check := &URLCheck{
		name:             name,
		timeout:          5 * time.Second,
		httpClient:       http.DefaultClient,
		validStatusCodes: []int{http.StatusOK, http.StatusNoContent, http.StatusAccepted, http.StatusCreated},
	}

	for _, opt := range opts {
		opt(check)
	}

	if err := check.Validate(); err != nil {
		return nil, err
	}

	return check, nil
}

func (c *URLCheck) Validate() error {
	if c.targetURL.String() == "" {
		return errors.New("target URL is required")
	}

	return nil
}

func (c *URLCheck) isValidStatusCode(code int) bool {
	return slices.Contains(c.validStatusCodes, code)
}

// GetName returns the name of the check
func (c *URLCheck) GetName() string {
	return c.name
}

// Check returns an error if the Result field is false
func (c *URLCheck) Check(ctx context.Context) health.CheckResult {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.targetURL.String(), nil)
	if err != nil {
		errW := fmt.Errorf("Failed to create request: %v", err)
		return health.CheckResult{
			Status:  health.StatusFail,
			Error:   errW,
			Message: errW.Error(),
		}
	}

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		errW := fmt.Errorf("Failed to create request: %v", err)
		return health.CheckResult{
			Status:  health.StatusFail,
			Error:   errW,
			Message: errW.Error(),
		}
	}

	reqDuration := time.Since(start)

	defer resp.Body.Close()

	if !c.isValidStatusCode(resp.StatusCode) {

		b, _ := io.ReadAll(resp.Body)

		errW := fmt.Errorf("unexpected response. status: %d body: %s", resp.StatusCode, string(b))

		return health.CheckResult{
			Status:  health.StatusFail,
			Error:   errW,
			Message: errW.Error(),
		}
	}

	return health.CheckResult{
		Status: health.StatusPass,
		Details: map[string]any{
			"url":        c.targetURL.String(),
			"statusCode": resp.StatusCode,
			"duration":   reqDuration.String(),
		},
	}
}
