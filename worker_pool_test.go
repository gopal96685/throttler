package throttler_test

import (
	"context"
	"testing"
	"time"

	"github.com/gopal96685/throttler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockExecutable struct {
	mock.Mock
}

func (m *mockExecutable) Execute(ctx context.Context, input string) (string, error) {
	args := m.Called(ctx, input)
	return args.String(0), args.Error(1)
}

func TestNewWorkerPool(t *testing.T) {

	t.Run("Valid Worker Pool", func(t *testing.T) {
		wp, err := throttler.NewWorkerPool[string, string](context.Background(),
			throttler.WithWorkers[string, string](2), throttler.WithRateLimit[string, string](5))
		assert.NoError(t, err)
		assert.NotNil(t, wp)
	})

	t.Run("Invalid Worker Pool", func(t *testing.T) {
		_, err := throttler.NewWorkerPool[string, string](context.Background(),
			throttler.WithWorkers[string, string](0), throttler.WithRateLimit[string, string](5))
		assert.Error(t, err, throttler.ErrInvalidWorkers.Error())
	})
}

func TestWorkerPool_AddTask(t *testing.T) {
	t.Run("Add Task to Worker Pool", func(t *testing.T) {

		wp, _ := throttler.NewWorkerPool[string, string](context.Background(),
			throttler.WithWorkers[string, string](2), throttler.WithRateLimit[string, string](5))

		exec := &mockExecutable{}
		exec.On("Execute", mock.Anything, "input").Return("output", nil)

		opts := throttler.TaskOptions{ID: "task1"}
		resultChan, err := wp.AddTask(opts, "input", exec)
		assert.NoError(t, err)
		assert.NotNil(t, resultChan)

	})

	t.Run("Pool Closed", func(t *testing.T) {
		wp, _ := throttler.NewWorkerPool[string, string](context.Background(),
			throttler.WithWorkers[string, string](2), throttler.WithRateLimit[string, string](5))
		wp.Stop()
		exec := &mockExecutable{}
		exec.On("Execute", mock.Anything, "input").Return("output", nil)

		opts := throttler.TaskOptions{ID: "task1"}
		_, err := wp.AddTask(opts, "input", exec)
		assert.Error(t, err, throttler.ErrPoolClosed.Error())
	})
}

func TestWorkerPool_AddFunction(t *testing.T) {
	t.Run("Add Function to Worker Pool", func(t *testing.T) {

		wp, _ := throttler.NewWorkerPool[string, string](context.Background(),
			throttler.WithWorkers[string, string](1), throttler.WithRateLimit[string, string](5))
		wp.Start()

		opts := throttler.TaskOptions{ID: "task1"}
		resultChan, err := wp.AddFunction(opts, "input", func(ctx context.Context, input string) (string, error) {
			return "executed", nil
		})

		result := <-resultChan
		assert.Equal(t, result.Result, "executed")
		assert.NoError(t, err)
		assert.NotNil(t, resultChan)

	})
}

func TestWorkerPool_Start(t *testing.T) {
	t.Run("Start Worker Pool", func(t *testing.T) {
		wp, _ := throttler.NewWorkerPool[string, string](context.Background(),
			throttler.WithWorkers[string, string](2), throttler.WithRateLimit[string, string](5))
		wp.Start()
		assert.NotNil(t, wp)
	})
}

func TestWorkerPool_Stop(t *testing.T) {
	t.Run("Stop Worker Pool", func(t *testing.T) {
		wp, _ := throttler.NewWorkerPool[string, string](context.Background(),
			throttler.WithWorkers[string, string](2), throttler.WithRateLimit[string, string](5))
		wp.Start()
		wp.Stop()

		// Attempting to add tasks after stopping should return error
		exec := &mockExecutable{}
		exec.On("Execute", mock.Anything, "input").Return("output", nil)

		opts := throttler.TaskOptions{ID: "task1"}
		_, err := wp.AddTask(opts, "input", exec)
		assert.Error(t, err, throttler.ErrPoolClosed.Error())
	})
}

func TestWorkerPool_SetRateLimit(t *testing.T) {
	t.Run("Valid Rate Limit", func(t *testing.T) {
		wp, _ := throttler.NewWorkerPool[string, string](context.Background(),
			throttler.WithWorkers[string, string](2), throttler.WithRateLimit[string, string](5))
		err := wp.SetRateLimit(10)
		assert.NoError(t, err)
	})

	t.Run("Invalid Rate Limit", func(t *testing.T) {
		wp, _ := throttler.NewWorkerPool[string, string](context.Background(),
			throttler.WithWorkers[string, string](2), throttler.WithRateLimit[string, string](5))
		err := wp.SetRateLimit(0)
		assert.Error(t, err, throttler.ErrInvalidRateLimit.Error())
	})
}

func TestWorkerPool_GetMetrics(t *testing.T) {
	t.Run("Get Metrics", func(t *testing.T) {
		wp, _ := throttler.NewWorkerPool[string, string](context.Background(),
			throttler.WithWorkers[string, string](2), throttler.WithRateLimit[string, string](5))
		wp.Start()

		// Add some tasks
		exec := &mockExecutable{}
		exec.On("Execute", mock.Anything, "input").Return("output", nil)

		opts := throttler.TaskOptions{ID: "task1"}
		resultChan, err := wp.AddTask(opts, "input", exec)
		assert.NoError(t, err)

		// Wait for task result and check the metrics
		go func() {
			result := <-resultChan
			assert.NoError(t, result.Error)
		}()

		time.Sleep(100 * time.Millisecond)

		metrics := wp.GetMetrics()
		assert.NotNil(t, metrics)
		assert.GreaterOrEqual(t, metrics["tasks_completed"], int64(0))
		assert.GreaterOrEqual(t, metrics["queue_size"], int64(0))
	})
}
