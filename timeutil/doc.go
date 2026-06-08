// Package timeutil provides utilities for working with time.
//
// # Clock
//
// Inject [Clock] into any struct that needs the current time:
//
//	type Cache struct {
//	    clock timeutil.Clock
//	}
//
//	func NewCache(clock timeutil.Clock) *Cache {
//	    return &Cache{clock: clock}
//	}
//
//	func (c *Cache) IsExpired(expiresAt time.Time) bool {
//	    return c.clock.Now().After(expiresAt)
//	}
//
// In production, pass [RealClock]:
//
//	cache := NewCache(timeutil.RealClock{})
//
// In tests, pass [FixedClock] to control time precisely:
//
//	fixed := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
//	cache := NewCache(timeutil.NewFixedClock(fixed))
//
//	assert.True(t, cache.IsExpired(fixed.Add(-time.Second)))
//	assert.False(t, cache.IsExpired(fixed.Add(time.Second)))
//
// # Sleeper
//
// Inject [Sleeper] into any struct that pauses execution — retry/backoff,
// polling, throttling — so tests run instantly instead of waiting on real time:
//
//	type Retrier struct {
//	    sleeper timeutil.Sleeper
//	}
//
//	func (r *Retrier) Do(fn func() error) error {
//	    backoff := 100 * time.Millisecond
//	    for i := 0; i < 5; i++ {
//	        if err := fn(); err == nil {
//	            return nil
//	        }
//	        r.sleeper.Sleep(backoff)
//	        backoff *= 2
//	    }
//	    return errors.New("max retries exceeded")
//	}
//
// In production, pass [RealSleeper]:
//
//	r := &Retrier{sleeper: timeutil.RealSleeper{}}
//
// In tests, pass [FakeSleeper] to assert on the schedule without waiting:
//
//	fake := timeutil.NewFakeSleeper()
//	r    := &Retrier{sleeper: fake}
//
//	_ = r.Do(failingFn)
//	assert.Equal(t, []time.Duration{100 * time.Millisecond, 200 * time.Millisecond}, fake.Calls())
//
// # Ticker
//
// Inject [Ticker] into any struct that runs periodic work — pollers,
// schedulers, heartbeats — so tests can drive ticks manually instead of
// waiting on real intervals:
//
//	type Poller struct {
//	    ticker timeutil.Ticker
//	}
//
//	func (p *Poller) Run(ctx context.Context, fn func()) {
//	    defer p.ticker.Stop()
//	    for {
//	        select {
//	        case <-ctx.Done():
//	            return
//	        case <-p.ticker.C():
//	            fn()
//	        }
//	    }
//	}
//
// In production, pass [RealTicker]:
//
//	p := &Poller{ticker: timeutil.NewRealTicker(time.Minute)}
//
// In tests, pass [FakeTicker] to trigger ticks on demand:
//
//	fake := timeutil.NewFakeTicker()
//	p    := &Poller{ticker: fake}
//
//	go p.Run(ctx, fn)
//	fake.Tick(time.Now())
//	fake.Tick(time.Now())
package timeutil
