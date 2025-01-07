package throttler

import (
	"context"
	"time"
)

// Executable interface that all tasks must implement.
type Executable[T any, R any] interface {
	Execute(ctx context.Context, input T) (R, error)
}

// TaskOptions provides common configurations for tasks.
type TaskOptions struct {
	ID         string
	Priority   int
	retryCount int
	MaxRetries int
	Backoff    time.Duration
	Timeout    time.Duration
	Context    context.Context
}

// Task wraps inputs, outputs, and execution logic for a generic task.
type TaskInfo[T any, R any] struct {
	TaskOptions
	input  T
	task   Executable[T, R]
	order  int64
	result chan TaskResult[R]
}

// TaskResult encapsulates the output or error after task execution.
type TaskResult[R any] struct {
	TaskID      string
	Result      R
	Error       error
	RetryCount  int
	CompletedAt time.Time
}

// TaskQueue implements a priority queue using a min-heap.
type TaskQueue[T any, R any] []*TaskInfo[T, R]

// NewTaskQueue creates a new task queue.
func NewTaskQueue[T any, R any]() *TaskQueue[T, R] {
	return &TaskQueue[T, R]{}
}

// Len returns the number of tasks in the queue.
func (h TaskQueue[T, R]) Len() int {
	return len(h)
}

// Less compares task priority and order for heap structure.
func (h TaskQueue[T, R]) Less(i, j int) bool {
	if h[i].Priority == h[j].Priority {
		return h[i].order < h[j].order
	}
	return h[i].Priority < h[j].Priority
}

// Swap switches two tasks in the heap.
func (h TaskQueue[T, R]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Push adds a task to the heap.
func (h *TaskQueue[T, R]) Push(x interface{}) {
	*h = append(*h, x.(*TaskInfo[T, R]))
}

// Pop removes and returns the highest-priority task.
func (h *TaskQueue[T, R]) Pop() interface{} {
	old := *h
	n := len(old)
	if n == 0 {
		return nil
	}
	item := old[0]
	*h = old[1:n]
	return item
}
