package test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gopal96685/throttler"
)

func TestWorkerPoolGracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wp, err := throttler.NewWorkerPool(ctx, throttler.WithWorkers[any, any](1), throttler.WithRateLimit[any, any](2))
	if err != nil {
		t.Fatalf("Failed to create WorkerPool: %v", err)
	}

	numTasks := 5
	tasksCompleted := int32(0)

	wp.Start()

	for i := 0; i < numTasks; i++ {
		_, err := wp.AddFunction(
			throttler.TaskOptions{},
			"input",
			func(ctx context.Context, inputs interface{}) (interface{}, error) {
				time.Sleep(100 * time.Millisecond) // Simulate work
				atomic.AddInt32(&tasksCompleted, 1)
				return nil, nil
			},
		)
		if err != nil {
			t.Errorf("Failed to add task: %v", err)
		}
	}
	time.Sleep(time.Millisecond * 1050)

	// Shutdown the wp and validate tasks completed
	wp.Stop()

	if tasksCompleted != 2 {
		t.Errorf("expected 2 tasks to get completed but got : %d", tasksCompleted)
	}
	if tasksCompleted == int32(numTasks) {
		t.Errorf("Unexpected: All tasks completed even though shutdown occurred")
	}
}
