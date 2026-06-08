package timeutil_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/lib-go/timeutil"
)

func TestRealClock_Now(t *testing.T) {
	t.Parallel()

	before := time.Now().UTC()
	got := timeutil.RealClock{}.Now()
	after := time.Now().UTC()

	assert.False(t, got.Before(before), "Now() should not be before call time")
	assert.False(t, got.After(after), "Now() should not be after call time")
	assert.Equal(t, time.UTC, got.Location(), "Now() should return UTC")
}

func TestFixedClock_Now(t *testing.T) {
	t.Parallel()

	fixed := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	c := timeutil.NewFixedClock(fixed)

	assert.Equal(t, fixed, c.Now())
	assert.Equal(t, fixed, c.Now(), "repeated calls return same time")
}

func TestClock_Interface(t *testing.T) {
	t.Parallel()

	// Both types satisfy the Clock interface.
	var _ timeutil.Clock = timeutil.RealClock{}
	var _ timeutil.Clock = timeutil.FixedClock{}
}
