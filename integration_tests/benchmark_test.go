package test

import (
	"context"
	"testing"

	"github.com/gopal96685/throttler"
)

func BenchmarkWorkerPoolExecution(b *testing.B) {
	ctx := context.Background()
	wp, _ := throttler.NewWorkerPool(ctx, throttler.WithWorkers[any, any](4), throttler.WithRateLimit[any, any](10000)) // High throughput
	wp.Start()
	defer wp.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resultChan, _ := wp.AddFunction(throttler.TaskOptions{
				ID: "benchmark-task",
			}, nil, func(ctx context.Context, inputs interface{}) (interface{}, error) {
				return nil, nil
			})
			<-resultChan
		}
	})
}
