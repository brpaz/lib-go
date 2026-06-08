package timeutil

import (
	"sync"
	"time"
)

// Sleeper pauses execution for a duration. Inject it into structs that sleep
// (retry/backoff, polling, throttling) so tests can run instantly instead of
// waiting on real time.
type Sleeper interface {
	Sleep(d time.Duration)
}

// RealSleeper sleeps for real, delegating to [time.Sleep].
type RealSleeper struct{}

func (RealSleeper) Sleep(d time.Duration) {
	time.Sleep(d)
}

// FakeSleeper records every requested duration instead of sleeping. Use it in
// tests to assert on retry/backoff schedules without waiting for real time to
// pass. Safe for concurrent use.
type FakeSleeper struct {
	mu    sync.Mutex
	Slept []time.Duration
}

// NewFakeSleeper creates an empty FakeSleeper.
func NewFakeSleeper() *FakeSleeper {
	return &FakeSleeper{}
}

// Sleep records d instead of sleeping.
func (s *FakeSleeper) Sleep(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Slept = append(s.Slept, d)
}

// Calls returns a copy of every duration recorded so far, in call order.
func (s *FakeSleeper) Calls() []time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	calls := make([]time.Duration, len(s.Slept))
	copy(calls, s.Slept)

	return calls
}

// Total returns the sum of every duration recorded so far.
func (s *FakeSleeper) Total() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	var total time.Duration
	for _, d := range s.Slept {
		total += d
	}

	return total
}
