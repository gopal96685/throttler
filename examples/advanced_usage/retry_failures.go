package main

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gopal96685/throttler"
)

func main() {
	ctx := context.Background()

	wp, _ := throttler.NewWorkerPool[string, string](
		ctx,
		throttler.WithWorkers[string, string](2),
		throttler.WithRateLimit[string, string](10),
	)

	wp.Start()
	defer wp.Stop()

	retryCounter := atomic.Int32{}
	maxRetries := 3

	// Add a task with retry logic
	resultChan, _ := wp.AddFunction(throttler.TaskOptions{
		ID:         "retry-task",
		MaxRetries: maxRetries,
		Backoff:    time.Millisecond * 100,
	}, "SampleInput", func(ctx context.Context, input string) (string, error) {
		count := retryCounter.Add(1)
		if count <= int32(maxRetries) {
			return "", errors.New("simulated failure")
		}
		return fmt.Sprintf("Succeeded after %d retries", count), nil
	})

	// Process the result
	result := <-resultChan
	if result.Error == nil {
		fmt.Println(result.Result)
	} else {
		fmt.Println("Failed with error:", result.Error)
	}
}
