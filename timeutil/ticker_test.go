package timeutil_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/lib-go/timeutil"
)

func TestRealTicker_C(t *testing.T) {
	t.Parallel()

	ticker := timeutil.NewRealTicker(10 * time.Millisecond)
	defer ticker.Stop()

	select {
	case got := <-ticker.C():
		assert.False(t, got.IsZero())
	case <-time.After(time.Second):
		t.Fatal("ticker did not tick within 1s")
	}
}

func TestFakeTicker_Tick(t *testing.T) {
	t.Parallel()

	ft := timeutil.NewFakeTicker()
	tm := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	go ft.Tick(tm)

	assert.Equal(t, tm, <-ft.C())
}

func TestFakeTicker_DoesNotTickOnItsOwn(t *testing.T) {
	t.Parallel()

	ft := timeutil.NewFakeTicker()

	select {
	case <-ft.C():
		t.Fatal("FakeTicker must not tick without an explicit Tick call")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestFakeTicker_Stop(t *testing.T) {
	t.Parallel()

	ft := timeutil.NewFakeTicker()

	assert.False(t, ft.Stopped())
	ft.Stop()
	assert.True(t, ft.Stopped())
	ft.Stop()
	assert.True(t, ft.Stopped(), "Stop should be idempotent")
}

func TestTicker_Interface(t *testing.T) {
	t.Parallel()

	var _ timeutil.Ticker = timeutil.NewRealTicker(time.Second)
	var _ timeutil.Ticker = timeutil.NewFakeTicker()
}
