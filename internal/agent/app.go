package agent

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/agent/collector"
	"github.com/ilyinon/go-musthave-metrics/internal/agent/sender"
	"github.com/ilyinon/go-musthave-metrics/internal/model"
)

type App struct {
	client         *sender.Client
	serverURL      string
	pollInterval   time.Duration
	reportInterval time.Duration

	gauges   atomic.Value
	counters atomic.Value
}

func New(
	client *sender.Client,
	serverURL string,
	poll, report time.Duration,
) *App {
	return &App{
		client:         client,
		serverURL:      serverURL,
		pollInterval:   poll,
		reportInterval: report,
	}
}

func (a *App) Run() {
	pollTicker := time.NewTicker(a.pollInterval)
	reportTicker := time.NewTicker(a.reportInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			a.gauges.Store(collector.Runtime())
			a.counters.Store(collector.Custom())

		case <-reportTicker.C:
			var batch []model.Metrics

			if g, ok := a.gauges.Load().(model.RuntimeMetrics); ok {
				for k, v := range g {
					val := v
					batch = append(batch, model.Metrics{
						ID:    k,
						MType: model.MetricGauge,
						Value: &val,
					})
				}
			}

			if c, ok := a.counters.Load().(model.CustomMetrics); ok {
				for k, v := range c {
					delta := v
					batch = append(batch, model.Metrics{
						ID:    k,
						MType: model.MetricCounter,
						Delta: &delta,
					})
				}
			}

			if len(batch) == 0 {
				continue
			}

			if err := a.client.Batch(batch); err != nil {
				log.Printf("batch send failed, fallback to single: %v", err)

				for _, m := range batch {
					switch m.MType {
					case model.MetricGauge:
						_ = a.client.Gauge(m.ID, *m.Value)
					case model.MetricCounter:
						_ = a.client.Counter(m.ID, *m.Delta)
					}
				}
			}
		}
	}
}
