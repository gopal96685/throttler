package main

import (
	"context"
	"fmt"

	"github.com/gopal96685/throttler"
)

// ExampleWorker implements Executable interface.
type ExampleWorker struct{}

func (w *ExampleWorker) Execute(ctx context.Context, input int) (string, error) {
	return fmt.Sprintf("Processed number: %d", input), nil
}

func main() {
	ctx := context.Background()
	wp, _ := throttler.NewWorkerPool[int, string](
		ctx,
		throttler.WithWorkers[int, string](3),
		throttler.WithRateLimit[int, string](10),
	)

	wp.Start()
	defer wp.Stop()

	// Add a task to process the number 42
	resultChan, _ := wp.AddTask(throttler.TaskOptions{ID: "task1"}, 42, &ExampleWorker{})

	// Get the result
	result := <-resultChan
	if result.Error == nil {
		fmt.Println("Success:", result.Result)
	} else {
		fmt.Println("Error:", result.Error)
	}
}
