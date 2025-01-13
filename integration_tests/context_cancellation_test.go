package test

import (
	"context"
	"errors"
	"testing"

	"github.com/gopal96685/throttler"
)

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wp, _ := throttler.NewWorkerPool(context.Background(), throttler.WithWorkers[any, any](2), throttler.WithRateLimit[any, any](10)) // 10 tasks/sec
	wp.Start()
	defer wp.Stop()

	cancelled := make(chan struct{})
	resultChan, _ := wp.AddFunction(throttler.TaskOptions{
		ID:      "cancelled-task",
		Context: ctx,
	}, nil, func(ctx context.Context, inputs interface{}) (interface{}, error) {
		<-ctx.Done()
		close(cancelled)
		return nil, ctx.Err()
	})
	cancel()
	res := <-resultChan

	<-cancelled
	if res.Error == nil || !errors.Is(res.Error, context.Canceled) {
		t.Errorf("expected context cancellation error, got: %v", res.Error)
	}
}
