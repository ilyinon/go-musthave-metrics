package agent

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/agent/collector"
	"github.com/ilyinon/go-musthave-metrics/internal/model"
)

const defaultRateLimit = 1

// MetricsSender sends metric batches.
type MetricsSender interface {
	Batch(ctx context.Context, metrics []model.Metrics) error
}

type singleMetricSender interface {
	Gauge(ctx context.Context, name string, value float64) error
	Counter(ctx context.Context, name string, value int64) error
}

// App represents the metrics agent responsible for collecting
// and sending metrics to the server.
type App struct {
	client         MetricsSender
	serverURL      string
	pollInterval   time.Duration
	reportInterval time.Duration
	rateLimit      int

	mu       sync.RWMutex
	gauges   model.RuntimeMetrics
	counters model.CustomMetrics

	lastCounters model.CustomMetrics
}

// New creates a new App with the given configuration.
func New(
	client MetricsSender,
	serverURL string,
	poll, report time.Duration,
	rateLimit int,
) *App {
	if rateLimit <= 0 {
		rateLimit = defaultRateLimit
	}

	return &App{
		client:         client,
		serverURL:      serverURL,
		pollInterval:   poll,
		reportInterval: report,
		rateLimit:      rateLimit,
		lastCounters:   make(model.CustomMetrics),
	}
}

// Run starts metric collection and reporting loops.
// The caller must pass a non-nil context.
func (a *App) Run(ctx context.Context) {
	pollTicker := time.NewTicker(a.pollInterval)
	reportTicker := time.NewTicker(a.reportInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	// worker pool for sending metrics batches concurrently
	sendCh := make(chan []model.Metrics, a.rateLimit)

	var wg sync.WaitGroup
	for i := 0; i < a.rateLimit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for batch := range sendCh {
				a.sendBatch(ctx, batch)
			}
		}()
	}

	for {
		select {
		case <-ctx.Done():
			close(sendCh)
			wg.Wait()
			return

		case <-pollTicker.C:
			a.collect()

		case <-reportTicker.C:
			batch := a.metricsBatch()
			if len(batch) == 0 {
				continue
			}

			select {
			case sendCh <- batch:
			case <-ctx.Done():
				close(sendCh)
				wg.Wait()
				return
			}
		}
	}
}

func (a *App) collect() {
	g := collector.Runtime()

	for k, v := range collector.Gopsutil() {
		g[k] = v
	}

	c := collector.Custom()

	a.mu.Lock()
	a.gauges = g
	a.counters = c
	a.mu.Unlock()
}

func (a *App) metricsBatch() []model.Metrics {
	a.mu.Lock()
	defer a.mu.Unlock()

	batch := make([]model.Metrics, 0, len(a.gauges)+len(a.counters))

	for k, v := range a.gauges {
		val := v
		batch = append(batch, model.Metrics{
			ID:    k,
			MType: model.MetricGauge,
			Value: &val,
		})
	}

	for k, v := range a.counters {
		delta := v - a.lastCounters[k]
		if delta == 0 {
			continue
		}
		if delta < 0 {
			delta = v
		}

		a.lastCounters[k] = v
		batch = append(batch, model.Metrics{
			ID:    k,
			MType: model.MetricCounter,
			Delta: &delta,
		})
	}

	return batch
}

func (a *App) sendBatch(ctx context.Context, batch []model.Metrics) {
	err := a.client.Batch(ctx, batch)
	if err == nil {
		return
	}
	if ctx.Err() != nil {
		return
	}

	log.Printf("ERROR: batch send failed: %v", err)

	singleSender, ok := a.client.(singleMetricSender)
	if !ok {
		return
	}

	for _, m := range batch {
		var err error

		switch m.MType {
		case model.MetricGauge:
			err = singleSender.Gauge(ctx, m.ID, *m.Value)
		case model.MetricCounter:
			err = singleSender.Counter(ctx, m.ID, *m.Delta)
		}

		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Printf("ERROR: metric send failed: id=%s type=%s err=%v", m.ID, m.MType, err)
		}
	}
}
