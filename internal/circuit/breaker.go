// internal/circuit/breaker.go
package circuit

import (
	"sync/atomic"
	"time"
)

type CircuitBreaker struct {
	failureThreshold int
	failureCount     atomic.Int32
	lastFailure      atomic.Int64
	state            atomic.Int32
	resetTimeout     time.Duration
}

func NewCircuitBreaker(threshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: threshold,
		resetTimeout:     resetTimeout,
	}
}

func (cb *CircuitBreaker) AllowRequest() bool {
	if cb == nil {
		return true
	}
	return cb.failureCount.Load() < int32(cb.failureThreshold)
}

func (cb *CircuitBreaker) RecordSuccess() {
	if cb == nil {
		return
	}
	cb.failureCount.Store(0)
}

func (cb *CircuitBreaker) RecordFailure() {
	if cb == nil {
		return
	}
	cb.failureCount.Add(1)
	cb.lastFailure.Store(time.Now().UnixNano())
}
