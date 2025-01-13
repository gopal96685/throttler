package test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gopal96685/throttler"
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
