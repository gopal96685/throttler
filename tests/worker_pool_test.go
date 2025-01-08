package test

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gopal96685/throttler"
	"github.com/stretchr/testify/assert"
)

type DummyTask struct {
	CompletedTasks *atomic.Int32
}

func (t *DummyTask) Execute(ctx context.Context, input any) (any, error) {

	input1 := input.(int)
	result := 1

	for i := 1; i <= input1; i++ {
		result *= i
	}

	t.CompletedTasks.Add(1)
	return result, nil
}

func TestWorkerPoolBehavior(t *testing.T) {
	tests := []struct {
		name       string
		numWorkers int
		rateLimit  int
		tasks      int
		wantErr    bool
	}{
		{"basic_task_execution", 2, 10, 5, false},
		{"excessive_tasks_low_workers", 1, 100, 50, false},
		{"invalid_worker_count", -1, 10, 1, true},
		{"zero_rate_limit", 2, 0, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			wp, err := throttler.NewWorkerPool(ctx,
				throttler.WithWorkers[any, any](tt.numWorkers),
				throttler.WithRateLimit[any, any](tt.rateLimit))

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			wp.Start()

			completedTasks := atomic.Int32{}
			for i := 0; i < tt.tasks; i++ {
				task := DummyTask{CompletedTasks: &completedTasks}
				_, err := wp.AddTask(throttler.TaskOptions{}, i, &task)

				if err != nil {
					t.Errorf("failed to add task %d: %v", i, err)
				}
			}

			time.Sleep(time.Duration(tt.tasks*1000/tt.rateLimit+100) * time.Millisecond)
			wp.Stop()

			if completedTasks.Load() != int32(tt.tasks) {
				t.Errorf("expected %d completed tasks, got %d", tt.tasks, completedTasks.Load())
			}
		})
	}
}

func TestRateLimitingAccuracy(t *testing.T) {
	ctx := context.Background()
	wp, _ := throttler.NewWorkerPool[interface{}, interface{}](ctx,
		throttler.WithWorkers[interface{}, interface{}](8),
		throttler.WithRateLimit[interface{}, interface{}](10)) // 10 tasks per second
	wp.Start()
	defer wp.Stop()

	taskCount := 50
	timestamps := make([]time.Time, taskCount)
	for i := 0; i < taskCount; i++ {
		resultChan, _ := wp.AddFunction(
			throttler.TaskOptions{ID: fmt.Sprintf("task-%d", i)},
			nil,
			func(ctx context.Context, inputs interface{}) (interface{}, error) {
				timestamps[i] = time.Now()
				return nil, nil
			})
		<-resultChan
	}

	for i := 1; i < taskCount; i++ {
		duration := timestamps[i].Sub(timestamps[i-1])
		expectedMin := time.Second / 10
		if duration < expectedMin*9/10 { // 10% tolerance
			t.Errorf("tasks ran faster than expected; interval between task %d and %d: %v", i-1, i, duration)
		}
	}
}

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

func TestHeavyLoadScenarios(t *testing.T) {
	ctx := context.Background()
	wp, _ := throttler.NewWorkerPool(ctx,
		throttler.WithWorkers[any, any](8),
		throttler.WithRateLimit[any, any](100)) //100 tasks/sec
	wp.Start()
	defer wp.Stop()

	numTasks := 500

	completedTasks := atomic.Int32{}
	for j := 0; j < numTasks; j++ {
		resultChan, _ := wp.AddFunction(throttler.TaskOptions{
			MaxRetries: 1,
			Backoff:    time.Millisecond * 10,
		}, nil, func(ctx context.Context, input interface{}) (interface{}, error) {
			rand.Seed(time.Now().UnixNano())
			randInt := rand.Intn(300)
			result := 1

			for i := 1; i <= randInt; i++ {
				result *= i
			}
			completedTasks.Add(1)
			return result, nil
		})

		go func() {
			res := <-resultChan
			if res.Error != nil {
				t.Errorf("task failed: %v", res.Error)
			}
		}()
	}

	time.Sleep(time.Second * 5)
	wp.Stop()

	if completedTasks.Load() != int32(numTasks) {
		t.Errorf("expected %d completed tasks, got %d", numTasks, completedTasks.Load())
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wp, _ := throttler.NewWorkerPool(context.Background(), throttler.WithWorkers[any, any](2), throttler.WithRateLimit[any, any](10)) // 10 tasks/sec
	wp.Start()
	defer wp.Stop()

	cancelled := make(chan struct{})
	resultChan, _ := wp.AddFunction(throttler.TaskOptions{
		ID:      "cancelled-task",
		Context: ctx,
	}, nil, func(ctx context.Context, inputs interface{}) (interface{}, error) {
		<-ctx.Done()
		close(cancelled)
		return nil, ctx.Err()
	})
	cancel()
	res := <-resultChan

	<-cancelled
	if res.Error == nil || !errors.Is(res.Error, context.Canceled) {
		t.Errorf("expected context cancellation error, got: %v", res.Error)
	}
}

func BenchmarkWorkerPoolExecution(b *testing.B) {
	ctx := context.Background()
	wp, _ := throttler.NewWorkerPool(ctx, throttler.WithWorkers[any, any](4), throttler.WithRateLimit[any, any](10000)) // High throughput
	wp.Start()
	defer wp.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resultChan, _ := wp.AddFunction(throttler.TaskOptions{
				ID: "benchmark-task",
			}, nil, func(ctx context.Context, inputs interface{}) (interface{}, error) {
				return nil, nil
			})
			<-resultChan
		}
	})
}

func TestWorkerPoolGracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wp, err := throttler.NewWorkerPool(ctx, throttler.WithWorkers[any, any](10), throttler.WithRateLimit[any, any](20))
	if err != nil {
		t.Fatalf("Failed to create WorkerPool: %v", err)
	}

	numTasks := 40
	tasksCompleted := int32(0)

	wp.Start()

	for i := 0; i < numTasks; i++ {
		_, err := wp.AddFunction(
			throttler.TaskOptions{},
			"input",
			func(ctx context.Context, inputs interface{}) (interface{}, error) {
				time.Sleep(10 * time.Millisecond) // Simulate work
				atomic.AddInt32(&tasksCompleted, 1)
				return nil, nil
			},
		)
		if err != nil {
			t.Errorf("Failed to add task: %v", err)
		}
	}

	// Shutdown the wp and validate tasks completed
	wp.Stop()

	if tasksCompleted == 0 {
		t.Errorf("No tasks completed before shutdown")
	}
	if tasksCompleted == int32(numTasks) {
		t.Errorf("Unexpected: All tasks completed even though shutdown occurred")
	}
}

func TestWorkerPoolRaceCondition(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := throttler.NewWorkerPool(ctx, throttler.WithWorkers[any, any](10), throttler.WithRateLimit[any, any](100))
	if err != nil {
		t.Fatalf("Failed to create WorkerPool: %v", err)
	}

	pool.Start()

	numTasks := 1000
	var wg sync.WaitGroup

	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = pool.AddFunction(
				throttler.TaskOptions{},
				"input",
				func(ctx context.Context, inputs interface{}) (interface{}, error) {
					return nil, nil
				},
			)
		}()
	}

	wg.Wait()
	pool.Stop()
}
