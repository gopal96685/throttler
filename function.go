package throttler

import "context"

// A wrapper type to turn closures into `Executable`
type ExecutableFunc[T, R any] func(ctx context.Context, input T) (R, error)

func (f ExecutableFunc[T, R]) Execute(ctx context.Context, input T) (R, error) {
	return f(ctx, input)
}
