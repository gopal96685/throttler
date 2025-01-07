package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/gopal96685/throttler"
)

func main() {
	ctx := context.Background()
	wp, _ := throttler.NewWorkerPool[string, string](
		ctx,
		throttler.WithWorkers[string, string](2),
		throttler.WithRateLimit[string, string](5), // Max 5 tasks/sec
	)

	wp.Start()
	defer wp.Stop()

	// Add an ad-hoc task using a closure
	resultChan, _ := wp.AddFunction(
		throttler.TaskOptions{ID: "closure-task"},
		"Hello",
		func(ctx context.Context, input string) (string, error) {
			if input == "" {
				return "", errors.New("empty input")
			}
			return fmt.Sprintf("Processed: %s", input), nil
		})

	// Process the result
	result := <-resultChan
	if result.Error == nil {
		fmt.Println("Success:", result.Result)
	} else {
		fmt.Println("Error:", result.Error)
	}
}
