package logger

import (
	"context"
	"sync"
	"time"
)

// DefaultThrottler with default config
var DefaultThrottler = GetThrottler(context.Background(), 5*time.Second, true)

// TrottlePrintf printf with default throttler and logger
func TrottlePrintf(format string, args ...interface{}) {
	DefaultThrottler.Trigger(func(ctx context.Context) {
		logger.Printf(format, args...)
	})
}

// GetThrottler return Throttler and start throttle ticker
func GetThrottler(ctx context.Context, duration time.Duration, trail bool) *Throttler {
	throttler := &Throttler{
		ctx:      ctx,
		duration: duration,
		trail:    trail,
		next:     nil,
	}
	return throttler
}

// ThrottleTriggerFunc trigger function of Throttler
type ThrottleTriggerFunc func(ctx context.Context)

// Throttler is instance of throttle
type Throttler struct {
	ctx      context.Context
	duration time.Duration
	trail    bool

	mutex sync.Mutex
	next  ThrottleTriggerFunc
	timer *time.Timer
}

// Trigger throttle function
func (t *Throttler) Trigger(f ThrottleTriggerFunc) {
	doFunc := false
	t.mutex.Lock()
	if t.timer == nil {
		t.timer = time.NewTimer(t.duration)
		t.next = nil
		doFunc = true
		go func() {
			select {
			case <-t.ctx.Done():
				return
			case <-t.timer.C:
				t.doNextFunc()
				t.mutex.Lock()
				if t.next == nil {
					t.timer.Stop()
					t.timer = nil
					t.mutex.Unlock()
					return
				}
				t.timer.Reset(t.duration)
				t.mutex.Unlock()
			}
		}()
	} else {
		if t.trail {
			t.next = f
		}
	}
	t.mutex.Unlock()

	if doFunc {
		f(t.ctx)
	}
}

func (t *Throttler) SetContext(ctx context.Context) {
	t.ctx = ctx
}

func (t *Throttler) doNextFunc() {
	t.mutex.Lock()
	next := t.next
	t.next = nil
	t.mutex.Unlock()
	if next != nil {
		next(t.ctx)
	}
}
