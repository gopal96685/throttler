package test

import (
	"context"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gopal96685/throttler"
)

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
