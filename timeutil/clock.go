package timeutil

import "time"

// Clock interface for time operations.
type Clock interface {
	Now() time.Time
}

// NewRealClock creates a new instance of RealClock.
// RealClock uses time.Now() to get the current time.
func NewRealClock() RealClock {
	return RealClock{}
}

// NewMockClock creates a new instance of MockClock.
// MockClock uses a fixed time to get the current time.
func NewMockClock(now time.Time) MockClock {
	return MockClock{FixedTime: now}
}

// RealClock is the real implementation using time.Now().
type RealClock struct{}

func (r RealClock) Now() time.Time {
	return time.Now()
}

// MockClock is a mock implementation of Clock.
type MockClock struct {
	FixedTime time.Time
}

func (m MockClock) Now() time.Time {
	return m.FixedTime
}
