package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gopal96685/throttler"
)

// Simulated database query
func dbQuery(ctx context.Context, input string) (string, error) {
	time.Sleep(50 * time.Millisecond) // Simulate query delay
	return fmt.Sprintf("Fetched record for key: %s", input), nil
}

func main() {
	ctx := context.Background()

	wp, _ := throttler.NewWorkerPool[string, string](
		ctx,
		throttler.WithWorkers[string, string](10),
		throttler.WithRateLimit[string, string](100), // Max 100 DB queries/sec
	)

	wp.Start()
	defer wp.Stop()

	keys := []string{"user-1", "user-2", "user-3"}
	for _, key := range keys {
		resultChan, _ := wp.AddFunction(throttler.TaskOptions{ID: key}, key, dbQuery)

		// Get the result
		result := <-resultChan
		fmt.Printf("Result for %s: %s (Error: %v)\n", key, result.Result, result.Error)
	}
}
