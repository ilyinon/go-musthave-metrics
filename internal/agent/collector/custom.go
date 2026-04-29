package collector

import (
	"sync/atomic"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
)

var pollCount int64

// Custom returns custom metrics collected by the agent.
// PollCount is incremented atomically on each call.
func Custom() model.CustomMetrics {
	return model.CustomMetrics{
		"PollCount": atomic.AddInt64(&pollCount, 1),
	}
}
