package timeutil_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/lib-go/timeutil"
)

func TestRealSleeper_Sleep(t *testing.T) {
	t.Parallel()

	before := time.Now()
	timeutil.RealSleeper{}.Sleep(10 * time.Millisecond)
	elapsed := time.Since(before)

	assert.GreaterOrEqual(t, elapsed, 10*time.Millisecond)
}

func TestFakeSleeper_Sleep(t *testing.T) {
	t.Parallel()

	s := timeutil.NewFakeSleeper()

	s.Sleep(100 * time.Millisecond)
	s.Sleep(200 * time.Millisecond)

	assert.Equal(t, []time.Duration{100 * time.Millisecond, 200 * time.Millisecond}, s.Calls())
	assert.Equal(t, 300*time.Millisecond, s.Total())
}

func TestFakeSleeper_DoesNotActuallySleep(t *testing.T) {
	t.Parallel()

	s := timeutil.NewFakeSleeper()

	before := time.Now()
	s.Sleep(time.Hour)
	elapsed := time.Since(before)

	assert.Less(t, elapsed, time.Second, "FakeSleeper must not block")
}

func TestFakeSleeper_ConcurrentSafe(t *testing.T) {
	t.Parallel()

	s := timeutil.NewFakeSleeper()

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Sleep(time.Millisecond)
		}()
	}
	wg.Wait()

	assert.Len(t, s.Calls(), 50)
	assert.Equal(t, 50*time.Millisecond, s.Total())
}

func TestSleeper_Interface(t *testing.T) {
	t.Parallel()

	var _ timeutil.Sleeper = timeutil.RealSleeper{}
	var _ timeutil.Sleeper = timeutil.NewFakeSleeper()
}
