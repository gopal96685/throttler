package throttler

import (
	"sync"
	"sync/atomic"
	"time"
)

// RateLimiter handles rate limiting with atomic operations for thread safety
type RateLimiter struct {
	ticker   *time.Ticker
	interval atomic.Int64 // Stores interval in nanoseconds
	mu       sync.Mutex   // Protects ticker updates
}

// newRateLimiter creates a new rate limiter with the specified interval
func newRateLimiter(interval time.Duration) *RateLimiter {
	r := &RateLimiter{}
	r.interval.Store(interval.Nanoseconds())
	r.ticker = time.NewTicker(interval)
	return r
}

// setInterval updates the rate limiter's interval
// interval is the new duration between ticks
func (r *RateLimiter) setInterval(interval time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store new interval
	r.interval.Store(interval.Nanoseconds())

	// Safely replace the ticker
	if r.ticker != nil {
		r.ticker.Stop()
	}
	r.ticker = time.NewTicker(interval)
}

// GetInterval returns the current interval in time.Duration
func (r *RateLimiter) GetInterval() time.Duration {
	return time.Duration(r.interval.Load())
}

// Stop stops the ticker
func (r *RateLimiter) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.ticker != nil {
		r.ticker.Stop()
		r.ticker = nil
	}
}
