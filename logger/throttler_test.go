package logger

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThrottlerWithoutTrail(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	var count int64
	triggerFunc := func(ctx context.Context) {
		atomic.AddInt64(&count, 1)
	}

	start := time.Now()
	throttler := GetThrottler(context.Background(), 50*time.Millisecond, false)
	throttler.Trigger(triggerFunc)
	require.EqualValues(1, count)
	require.WithinDuration(start, time.Now(), 10*time.Millisecond)

	throttler.Trigger(triggerFunc)
	require.EqualValues(1, count)
	require.WithinDuration(start, time.Now(), 10*time.Millisecond)

	time.Sleep(100 * time.Millisecond)
	require.EqualValues(1, count)
	require.WithinDuration(start, time.Now(), 110*time.Millisecond)

	throttler.Trigger(triggerFunc)
	throttler.Trigger(triggerFunc)
	require.EqualValues(2, count)
	require.WithinDuration(start, time.Now(), 110*time.Millisecond)

	time.Sleep(100 * time.Millisecond)
	require.EqualValues(2, count)
	require.WithinDuration(start, time.Now(), 210*time.Millisecond)
}

func TestThrottlerWithTrail(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	var count int64
	triggerFunc := func(ctx context.Context) {
		atomic.AddInt64(&count, 1)
	}

	start := time.Now()
	throttler := GetThrottler(context.Background(), 50*time.Millisecond, true)
	throttler.Trigger(triggerFunc)
	require.EqualValues(1, count)
	require.WithinDuration(start, time.Now(), 10*time.Millisecond)

	throttler.Trigger(triggerFunc)
	require.EqualValues(1, count)
	require.WithinDuration(start, time.Now(), 10*time.Millisecond)

	time.Sleep(100 * time.Millisecond)
	require.EqualValues(2, count)
	require.WithinDuration(start, time.Now(), 110*time.Millisecond)

	throttler.Trigger(triggerFunc)
	throttler.Trigger(triggerFunc)
	require.EqualValues(3, count)
	require.WithinDuration(start, time.Now(), 110*time.Millisecond)

	time.Sleep(100 * time.Millisecond)
	require.EqualValues(4, count)
	require.WithinDuration(start, time.Now(), 210*time.Millisecond)
}
