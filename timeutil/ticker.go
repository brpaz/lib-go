package timeutil

import (
	"sync"
	"time"
)

// Ticker delivers the current time on a channel at regular intervals. Inject
// it into structs that run periodic work — pollers, schedulers, heartbeats —
// so tests can drive ticks manually instead of waiting on real intervals.
type Ticker interface {
	// C returns the channel on which ticks are delivered.
	C() <-chan time.Time

	// Stop turns off the ticker. No more ticks are sent after Stop returns.
	Stop()
}

// RealTicker ticks for real, delegating to [time.Ticker].
type RealTicker struct {
	ticker *time.Ticker
}

// NewRealTicker creates a RealTicker that ticks every d.
func NewRealTicker(d time.Duration) *RealTicker {
	return &RealTicker{ticker: time.NewTicker(d)}
}

func (t *RealTicker) C() <-chan time.Time {
	return t.ticker.C
}

func (t *RealTicker) Stop() {
	t.ticker.Stop()
}

// FakeTicker delivers ticks only when told to via Tick. Use it in tests to
// drive periodic code deterministically without waiting on real time. Safe
// for concurrent use.
type FakeTicker struct {
	ch chan time.Time

	mu      sync.Mutex
	stopped bool
}

// NewFakeTicker creates a FakeTicker with an unbuffered channel.
func NewFakeTicker() *FakeTicker {
	return &FakeTicker{ch: make(chan time.Time)}
}

func (t *FakeTicker) C() <-chan time.Time {
	return t.ch
}

// Tick sends tm on the ticker's channel, simulating an elapsed interval. It
// blocks until a receiver reads it.
func (t *FakeTicker) Tick(tm time.Time) {
	t.ch <- tm
}

// Stop marks the ticker as stopped. Safe to call multiple times.
func (t *FakeTicker) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.stopped = true
}

// Stopped reports whether Stop has been called.
func (t *FakeTicker) Stopped() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.stopped
}
