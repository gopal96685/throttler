package test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gopal96685/throttler"
	"github.com/stretchr/testify/assert"
)

// RetryTask implements the Executable interface and simulates retry behavior.
type RetryTask struct {
	RetryCounter *atomic.Int32
	MaxRetries   int
}

// Execute increments the retry counter and fails until retries are exhausted.
func (rt *RetryTask) Execute(ctx context.Context, input any) (any, error) {
	count := rt.RetryCounter.Add(1)
	if count <= int32(rt.MaxRetries) {
		return "", errors.New("simulated failure")
	}
	return "success", nil
}

func TestWorkerPoolWithRetryTask(t *testing.T) {
	// Create a worker pool with 1 worker and rate limit of 5 tasks per second.
	ctx := context.Background()
	wp, _ := throttler.NewWorkerPool(ctx, throttler.WithWorkers[any, any](1),
		throttler.WithRateLimit[any, any](2))

	wp.Start()
	defer wp.Stop()

	// Define task properties.
	maxRetries := 3
	retryCounter := atomic.Int32{}

	// Create task options.
	taskOptions := throttler.TaskOptions{
		ID:         "retry-task",
		MaxRetries: maxRetries,
		Backoff:    10 * time.Millisecond,
	}

	// Add the task to the pool.
	resultChan, err := wp.AddTask(taskOptions, "input-data", &RetryTask{
		RetryCounter: &retryCounter,
		MaxRetries:   maxRetries,
	})
	assert.NoError(t, err, "Failed to add task to worker pool")

	// Wait for task result.
	select {
	case res := <-resultChan:
		assert.NoError(t, res.Error, "Task execution failed unexpectedly")
		assert.Equal(t, "success", res.Result, "Unexpected task result")
		assert.Equal(t, int32(maxRetries+1), retryCounter.Load(), "Unexpected number of retries")
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for task result")
	}
}
