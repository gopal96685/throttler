package throttler

import (
	"context"
	"errors"
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

// TaskQueue implements a priority queue as a min-heap.
type TaskQueue[T any, R any] struct {
	tasks []*TaskInfo[T, R]
}

// NewTaskQueue creates a new task queue.
func NewTaskQueue[T any, R any]() *TaskQueue[T, R] {
	return &TaskQueue[T, R]{
		tasks: make([]*TaskInfo[T, R], 0),
	}
}

// Len returns the number of tasks in the queue.
func (q *TaskQueue[T, R]) Len() int {
	return len(q.tasks)
}

// Insert adds a task into the queue and maintains heap properties.
func (q *TaskQueue[T, R]) Insert(task *TaskInfo[T, R]) {
	q.tasks = append(q.tasks, task)
	q.bubbleUp(len(q.tasks) - 1)
}

// Extract removes and returns the task with the highest priority.
func (q *TaskQueue[T, R]) Extract() (*TaskInfo[T, R], error) {
	if q.Len() == 0 {
		return nil, errors.New("task queue is empty")
	}

	// Swap root with last element.
	n := len(q.tasks) - 1
	q.swap(0, n)

	// Remove the last element (smallest task).
	task := q.tasks[n]
	q.tasks = q.tasks[:n]

	// Restore heap properties.
	if len(q.tasks) > 0 {
		q.bubbleDown(0)
	}

	return task, nil
}

// bubbleUp restores the heap property upwards from the given index.
func (q *TaskQueue[T, R]) bubbleUp(index int) {
	parent := (index - 1) / 2
	if parent >= 0 && q.less(index, parent) {
		q.swap(index, parent)
		q.bubbleUp(parent)
	}
}

// bubbleDown restores the heap property downwards from the given index.
func (q *TaskQueue[T, R]) bubbleDown(index int) {
	left := 2*index + 1
	right := 2*index + 2
	smallest := index

	if left < len(q.tasks) && q.less(left, smallest) {
		smallest = left
	}
	if right < len(q.tasks) && q.less(right, smallest) {
		smallest = right
	}

	if smallest != index {
		q.swap(index, smallest)
		q.bubbleDown(smallest)
	}
}

// less determines the priority between two tasks.
func (q *TaskQueue[T, R]) less(i, j int) bool {
	if q.tasks[i].Priority == q.tasks[j].Priority {
		return q.tasks[i].order < q.tasks[j].order
	}
	return q.tasks[i].Priority < q.tasks[j].Priority
}

// swap exchanges two tasks in the heap.
func (q *TaskQueue[T, R]) swap(i, j int) {
	q.tasks[i], q.tasks[j] = q.tasks[j], q.tasks[i]
}
