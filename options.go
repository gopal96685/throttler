package throttler

import (
	"fmt"
	"time"

	"github.com/gopal96685/throttler/internal/circuit"
	"github.com/gopal96685/throttler/internal/metrics"
)

// Option defines a functional configuration option for WorkerPool.
type Option[T, R any] func(*WorkerPool[T, R]) error

// WithWorkers configures the number of workers in the pool.
func WithWorkers[T, R any](n int) Option[T, R] {
	return func(wp *WorkerPool[T, R]) error {
		if n > 0 {
			wp.numWorkers = n
			return nil
		}
		return fmt.Errorf("invalid number of workers")
	}
}

// WithRateLimit configures the rate limit in requests per second for the pool.
func WithRateLimit[T, R any](rps int) Option[T, R] {
	return func(wp *WorkerPool[T, R]) error {
		if rps > 0 {
			wp.rateLimit = newRateLimiter(time.Second / time.Duration(rps))
			return nil
		}
		return fmt.Errorf("invalid value for rate limit (RPS)")
	}
}

// WithCustomMetrics injects a custom metrics implementation into the WorkerPool.
func WithCustomMetrics[T, R any](metrics *metrics.DefaultCollector) Option[T, R] {
	return func(wp *WorkerPool[T, R]) error {
		if metrics != nil {
			wp.metrics = metrics
			return nil
		}
		return fmt.Errorf("metrics struct cannot be nil")
	}
}

// WithCircuitBreaker enables a circuit breaker for the WorkerPool.
func WithCircuitBreaker[T, R any](threshold int, resetTimeout time.Duration) Option[T, R] {
	return func(wp *WorkerPool[T, R]) error {
		if threshold <= 0 || resetTimeout <= 0 {
			return fmt.Errorf("invalid circuit breaker configuration")
		}
		wp.circuitBreaker = circuit.NewCircuitBreaker(threshold, resetTimeout)
		return nil
	}
}
