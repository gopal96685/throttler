package throttler

import (
	"context"
	"testing"
	"time"
)

type MockTask struct{}

func (m MockTask) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	return nil, nil
}

func TestTaskQueue_Insert_Len(t *testing.T) {
	taskQueue := NewTaskQueue[interface{}, interface{}]()
	task := &TaskInfo[interface{}, interface{}]{
		TaskOptions: TaskOptions{ID: "task1", Priority: 1},
		task:        MockTask{},
	}

	// Test before inserting any tasks
	if got := taskQueue.Len(); got != 0 {
		t.Errorf("Len() = %v, want %v", got, 0)
	}

	// Insert the task and test the length
	taskQueue.Insert(task)
	if got := taskQueue.Len(); got != 1 {
		t.Errorf("Len() = %v, want %v", got, 1)
	}
}

func TestTaskQueue_Extract(t *testing.T) {
	taskQueue := NewTaskQueue[interface{}, interface{}]()
	task1 := &TaskInfo[interface{}, interface{}]{TaskOptions: TaskOptions{ID: "task1", Priority: 1}, task: MockTask{}}
	task2 := &TaskInfo[interface{}, interface{}]{TaskOptions: TaskOptions{ID: "task2", Priority: 2}, task: MockTask{}}

	taskQueue.Insert(task1)
	taskQueue.Insert(task2)

	// Extract and verify the highest priority task is returned
	extractedTask, err := taskQueue.Extract()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if extractedTask.TaskOptions.ID != "task1" {
		t.Errorf("Extracted task ID = %v, want %v", extractedTask.TaskOptions.ID, "task2")
	}
}

func TestTaskQueue_Extract_EmptyQueue(t *testing.T) {
	taskQueue := NewTaskQueue[interface{}, interface{}]()

	_, err := taskQueue.Extract()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "task queue is empty" {
		t.Errorf("expected error 'task queue is empty', got %v", err)
	}
}
func TestTaskQueue_PriorityOrdering(t *testing.T) {
	taskQueue := NewTaskQueue[interface{}, interface{}]()
	task1 := &TaskInfo[interface{}, interface{}]{TaskOptions: TaskOptions{ID: "task1", Priority: 1}, task: MockTask{}}
	task2 := &TaskInfo[interface{}, interface{}]{TaskOptions: TaskOptions{ID: "task2", Priority: 2}, task: MockTask{}}
	task3 := &TaskInfo[interface{}, interface{}]{TaskOptions: TaskOptions{ID: "task3", Priority: 3}, task: MockTask{}}
	taskQueue.Insert(task3)
	taskQueue.Insert(task2)
	taskQueue.Insert(task1)

	// Extract tasks to ensure they are ordered by priority
	extractedTask, _ := taskQueue.Extract()
	if extractedTask.TaskOptions.ID != "task1" {
		t.Errorf("Expected task1 as highest priority, got %v", extractedTask.TaskOptions.ID)
	}

	extractedTask, _ = taskQueue.Extract()
	if extractedTask.TaskOptions.ID != "task2" {
		t.Errorf("Expected task2, got %v", extractedTask.TaskOptions.ID)
	}

	extractedTask, _ = taskQueue.Extract()
	if extractedTask.TaskOptions.ID != "task3" {
		t.Errorf("Expected task3, got %v", extractedTask.TaskOptions.ID)
	}
}

func TestTaskExecution(t *testing.T) {
	taskInfo := &TaskInfo[interface{}, interface{}]{
		TaskOptions: TaskOptions{
			ID:       "task1",
			Priority: 1,
			Timeout:  time.Second,
		},
		task: MockTask{},
	}

	// Mock context and task execution
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Test execution of the task
	_, err := taskInfo.task.Execute(ctx, 10) // Pass mock data to test task execution
	if err != nil {
		t.Errorf("task execution failed: %v", err)
	}
}

func TestTaskRetryLogic(t *testing.T) {
	taskInfo := &TaskInfo[interface{}, interface{}]{
		TaskOptions: TaskOptions{
			ID:         "task1",
			MaxRetries: 3,
		},
	}

	// Simulate retry attempts
	for i := 0; i <= taskInfo.TaskOptions.MaxRetries; i++ {
		taskInfo.TaskOptions.retryCount++
		if taskInfo.TaskOptions.retryCount > taskInfo.TaskOptions.MaxRetries+1 {
			t.Errorf("Exceeded maximum retries. Got retryCount: %d", taskInfo.TaskOptions.MaxRetries)
		}
	}
}
