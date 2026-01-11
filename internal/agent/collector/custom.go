package collector

import (
	"sync/atomic"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
)

var pollCount int64

func Custom() model.CustomMetrics {
	return model.CustomMetrics{
		"PollCount": atomic.AddInt64(&pollCount, 1),
	}
}
