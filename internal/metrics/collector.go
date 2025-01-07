// internal/metrics/collector.go
package metrics

import (
	"sync/atomic"
)

type DefaultCollector struct {
	TasksCompleted   atomic.Int64
	TasksFailed      atomic.Int64
	TasksRetried     atomic.Int64
	AverageLatency   atomic.Int64
	CurrentQueueSize atomic.Int64
}

func (dc *DefaultCollector) Load() map[string]int64 {
	return map[string]int64{
		"tasks_completed":    dc.TasksCompleted.Load(),
		"tasks_failed":       dc.TasksFailed.Load(),
		"tasks_retried":      dc.TasksRetried.Load(),
		"average_latency_ms": dc.AverageLatency.Load(),
		"queue_size":         dc.CurrentQueueSize.Load(),
	}
}
