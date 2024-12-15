package timeutil_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/lib-go/timeutil"
)

func TestNewRealClock(t *testing.T) {
	t.Parallel()

	// Create a new RealClock instance.
	clock := timeutil.NewRealClock()

	now := clock.Now()
	assert.NotZero(t, now)
}

func TestNewMockClock(t *testing.T) {
	t.Parallel()

	mockTime := time.Date(2023, time.October, 10, 9, 0, 0, 0, time.UTC)

	clock := timeutil.NewMockClock(mockTime)

	now := clock.Now()
	assert.Equal(t, mockTime, now)
}
