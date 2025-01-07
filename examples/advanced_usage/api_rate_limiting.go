package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gopal96685/throttler"
)

func main() {
	ctx := context.Background()

	// Configure pool to mimic 10 requests/sec API rate limit
	wp, _ := throttler.NewWorkerPool[string, string](
		ctx,
		throttler.WithWorkers[string, string](5),
		throttler.WithRateLimit[string, string](10),
	)

	wp.Start()
	defer wp.Stop()

	// Queue API request simulation
	resultChan, _ := wp.AddFunction(throttler.TaskOptions{ID: "api-task"}, "https://api.example.com/data", func(ctx context.Context, url string) (string, error) {
		// Simulate API call
		time.Sleep(100 * time.Millisecond)
		return fmt.Sprintf("Fetched data from %s", url), nil
	})

	// Process the result
	result := <-resultChan
	fmt.Println("API Task:", result.Result)
}
