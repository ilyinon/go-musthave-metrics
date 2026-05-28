package agent

import (
	"testing"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
)

func TestMetricsBatchSendsCounterDelta(t *testing.T) {
	app := New(nil, "", time.Second, time.Second, 1)

	app.counters = model.CustomMetrics{"PollCount": 3}
	first := app.metricsBatch()
	assertCounterDelta(t, first, 3)

	app.counters = model.CustomMetrics{"PollCount": 5}
	second := app.metricsBatch()
	assertCounterDelta(t, second, 2)
}

func assertCounterDelta(t *testing.T, metrics []model.Metrics, want int64) {
	t.Helper()

	for _, metric := range metrics {
		if metric.ID == "PollCount" && metric.MType == model.MetricCounter {
			if metric.Delta == nil {
				t.Fatal("PollCount delta is nil")
			}
			if *metric.Delta != want {
				t.Fatalf("PollCount delta = %d, want %d", *metric.Delta, want)
			}
			return
		}
	}

	t.Fatal("PollCount metric not found")
}
