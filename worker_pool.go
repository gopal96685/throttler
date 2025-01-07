package throttler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gopal96685/throttler/internal/circuit"
	"github.com/gopal96685/throttler/internal/metrics"
)

var (
	ErrPoolClosed       = errors.New("worker pool is closed")
	ErrInvalidWorkers   = errors.New("number of workers must be positive")
	ErrInvalidRateLimit = errors.New("rate limit must be positive")
	ErrTaskCanceled     = errors.New("task was canceled")
)

type WorkerPool[T any, R any] struct {
	numWorkers int
	taskQueue  *TaskQueue[T, R]
	rateLimit  *RateLimiter

	metrics        *metrics.DefaultCollector
	circuitBreaker *circuit.CircuitBreaker

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	status int32 // Using atomic operations
	mu     sync.RWMutex
}

func NewWorkerPool[T any, R any](ctx context.Context, opts ...Option[T, R]) (*WorkerPool[T, R], error) {
	wp := &WorkerPool[T, R]{
		numWorkers: 1,
		taskQueue:  NewTaskQueue[T, R](),
		metrics:    &metrics.DefaultCollector{},
	}

	for _, opt := range opts {
		if err := opt(wp); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Validate that critical configurations are set
	if wp.numWorkers <= 0 {
		return nil, ErrInvalidWorkers
	}

	if wp.rateLimit == nil {
		return nil, errors.New("rate limiter is required but not provided")
	}

	wp.ctx, wp.cancel = context.WithCancel(ctx)
	return wp, nil
}

func (wp *WorkerPool[T, R]) AddTask(opts TaskOptions, input T, task Executable[T, R],
) (<-chan TaskResult[R], error) {

	if atomic.LoadInt32(&wp.status) != 0 {
		return nil, ErrPoolClosed
	}

	if opts.Context == nil {
		opts.Context = context.Background()
	}

	resultChan := make(chan TaskResult[R], 1)
	taskInfo := &TaskInfo[T, R]{
		TaskOptions: opts,
		input:       input,
		task:        task,
		order:       time.Now().UnixNano(),
		result:      resultChan,
	}

	wp.mu.Lock()
	wp.taskQueue.Push(taskInfo)
	wp.metrics.CurrentQueueSize.Add(1)
	wp.mu.Unlock()

	return resultChan, nil
}

func (wp *WorkerPool[T, R]) AddFunction(
	opts TaskOptions,
	input T,
	executeFunc func(ctx context.Context, input T) (R, error),
) (<-chan TaskResult[R], error) {

	if atomic.LoadInt32(&wp.status) != 0 {
		return nil, ErrPoolClosed
	}

	if opts.Context == nil {
		opts.Context = context.Background()
	}

	resultChan := make(chan TaskResult[R], 1)
	taskInfo := &TaskInfo[T, R]{
		TaskOptions: opts,
		input:       input,
		task:        ExecutableFunc[T, R](executeFunc),
		order:       time.Now().UnixNano(),
		result:      resultChan,
	}

	wp.mu.Lock()
	wp.taskQueue.Push(taskInfo)
	wp.metrics.CurrentQueueSize.Add(1)
	wp.mu.Unlock()

	return resultChan, nil
}

func (wp *WorkerPool[T, R]) Start() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if atomic.LoadInt32(&wp.status) != 0 {
		fmt.Println("Worker pool is already running")
		return
	}

	wp.wg.Add(wp.numWorkers)
	for i := 0; i < wp.numWorkers; i++ {
		go wp.worker()
	}
}

func (wp *WorkerPool[T, R]) worker() {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			return
		default:
			<-wp.rateLimit.ticker.C
			wp.mu.Lock()
			if wp.taskQueue.Len() == 0 {
				wp.mu.Unlock()
				time.Sleep(10 * time.Millisecond)
				continue
			}

			taskInfo := wp.taskQueue.Pop().(*TaskInfo[T, R])
			wp.metrics.CurrentQueueSize.Add(-1)
			wp.mu.Unlock()

			startTime := time.Now()

			taskCtx := taskInfo.Context
			if taskInfo.Timeout > 0 {
				var cancel context.CancelFunc
				taskCtx, cancel = context.WithTimeout(taskInfo.Context, taskInfo.Timeout)
				defer cancel()
			}

			result, err := taskInfo.task.Execute(taskCtx, taskInfo.input)

			taskResult := TaskResult[R]{
				TaskID:      taskInfo.ID,
				RetryCount:  taskInfo.retryCount,
				CompletedAt: time.Now(),
				Error:       err,
				Result:      result,
			}

			if err != nil && !errors.Is(err, context.Canceled) {
				wp.metrics.TasksFailed.Add(1)
				if taskInfo.retryCount < taskInfo.MaxRetries {
					taskInfo.retryCount++
					wp.metrics.TasksRetried.Add(1)
					time.Sleep(taskInfo.Backoff)

					wp.mu.Lock()
					wp.taskQueue.Push(taskInfo)
					wp.metrics.CurrentQueueSize.Add(1)
					wp.mu.Unlock()
					continue
				}
			} else {
				wp.metrics.TasksCompleted.Add(1)
			}

			latency := time.Since(startTime).Milliseconds()
			wp.metrics.AverageLatency.Store(latency)

			select {
			case taskInfo.result <- taskResult:
			default:
			}
		}
	}
}

func (wp *WorkerPool[T, R]) Stop() {
	if !atomic.CompareAndSwapInt32(&wp.status, 0, 1) {
		return
	}

	wp.cancel()
	wp.wg.Wait()
}

func (wp *WorkerPool[T, R]) SetRateLimit(requestsPerSecond int) error {
	if requestsPerSecond <= 0 {
		return ErrInvalidRateLimit
	}

	interval := time.Second / time.Duration(requestsPerSecond)
	wp.rateLimit.setInterval(interval)
	return nil
}

func (wp *WorkerPool[T, R]) GetMetrics() map[string]int64 {
	return map[string]int64{
		"tasks_completed":    wp.metrics.TasksCompleted.Load(),
		"tasks_failed":       wp.metrics.TasksFailed.Load(),
		"tasks_retried":      wp.metrics.TasksRetried.Load(),
		"average_latency_ms": wp.metrics.AverageLatency.Load(),
		"queue_size":         wp.metrics.CurrentQueueSize.Load(),
	}
}
