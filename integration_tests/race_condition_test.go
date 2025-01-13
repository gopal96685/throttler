package test

import (
	"context"
	"sync"
	"testing"

	"github.com/gopal96685/throttler"
)

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
