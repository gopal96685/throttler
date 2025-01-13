package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gopal96685/throttler"
)

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
