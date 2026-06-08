package timeutil

import "time"

// Clock provides the current time. Inject it into structs that need time so tests can control it.
type Clock interface {
	Now() time.Time
}

// RealClock returns the actual current time.
type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now().UTC()
}

// FixedClock always returns the same time. Use in tests.
type FixedClock struct {
	T time.Time
}

func (c FixedClock) Now() time.Time {
	return c.T
}

// NewFixedClock creates a FixedClock at the given time.
func NewFixedClock(t time.Time) FixedClock {
	return FixedClock{T: t}
}
