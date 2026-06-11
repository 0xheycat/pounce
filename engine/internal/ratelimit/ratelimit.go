// Package ratelimit provides a simple token-bucket limiter measured in
// bytes/second, used to throttle download throughput.
package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter is a token-bucket rate limiter. A rate of 0 means unlimited.
type Limiter struct {
	mu     sync.Mutex
	rate   float64 // bytes per second
	tokens float64
	max    float64
	last   time.Time
}

// New returns a Limiter for the given rate in bytes/second (0 = unlimited).
func New(rate int64) *Limiter {
	l := &Limiter{last: time.Now()}
	l.SetRate(rate)
	return l
}

// minBurst is the smallest bucket capacity. It must be at least as large as the
// biggest single WaitN request (the download read-chunk size, 32 KiB) so that
// rates below the chunk size still throttle instead of letting every oversized
// chunk pass straight through.
const minBurst = 64 * 1024

// SetRate updates the limit live. 0 disables throttling.
func (l *Limiter) SetRate(rate int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rate = float64(rate)
	if rate <= 0 {
		l.max = 0
		l.tokens = 0
		l.last = time.Now()
		return
	}
	// Allow up to a second of burst, but never less than minBurst so a single
	// read chunk can always accumulate enough tokens to be admitted (and thus
	// throttled) instead of bypassing the limiter at low rates.
	l.max = float64(rate)
	if l.max < minBurst {
		l.max = minBurst
	}
	l.tokens = l.max
	l.last = time.Now()
}

// WaitN blocks until n bytes may be consumed or ctx is done. It returns
// immediately when the limiter is unlimited.
func (l *Limiter) WaitN(ctx context.Context, n int) error {
	l.mu.Lock()
	if l.rate <= 0 {
		l.mu.Unlock()
		return nil
	}
	l.mu.Unlock()

	for {
		l.mu.Lock()
		now := time.Now()
		l.tokens += now.Sub(l.last).Seconds() * l.rate
		l.last = now
		if l.tokens > l.max {
			l.tokens = l.max
		}
		// If a single chunk is larger than the bucket, allow it through to
		// avoid a permanent stall.
		if float64(n) >= l.max || l.tokens >= float64(n) {
			l.tokens -= float64(n)
			l.mu.Unlock()
			return nil
		}
		wait := time.Duration((float64(n) - l.tokens) / l.rate * float64(time.Second))
		l.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
}
