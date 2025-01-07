package main

import (
	"context"
	"fmt"

	"github.com/gopal96685/throttler"
)

func main() {
	ctx := context.Background()
	wp, _ := throttler.NewWorkerPool[int, int](
		ctx,
		throttler.WithWorkers[int, int](3),
		throttler.WithRateLimit[int, int](10),
	)

	wp.Start()
	defer wp.Stop()

	// Queue multiple tasks
	results := make([]<-chan throttler.TaskResult[int], 0)
	for i := 1; i <= 50; i++ {
		taskChan, _ := wp.AddFunction(throttler.TaskOptions{ID: fmt.Sprintf("task-%d", i)}, i, func(ctx context.Context, input int) (int, error) {
			return input * input, nil // Square the number
		})
		results = append(results, taskChan)
	}

	// Collect results
	for i, taskChan := range results {
		result := <-taskChan
		fmt.Printf("Task %d Result: %d, Error: %v\n", i+1, result.Result, result.Error)
	}
}
