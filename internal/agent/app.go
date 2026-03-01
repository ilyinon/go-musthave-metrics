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
	rateLimit      int

	gauges   atomic.Value
	counters atomic.Value
}

func New(
	client *sender.Client,
	serverURL string,
	poll, report time.Duration,
	rateLimit int,
) *App {
	if rateLimit <= 0 {
		rateLimit = 1
	}

	return &App{
		client:         client,
		serverURL:      serverURL,
		pollInterval:   poll,
		reportInterval: report,
		rateLimit:      rateLimit,
	}
}

func (a *App) Run() {
	pollTicker := time.NewTicker(a.pollInterval)
	reportTicker := time.NewTicker(a.reportInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	sendCh := make(chan []model.Metrics, a.rateLimit)

	for i := 0; i < a.rateLimit; i++ {
		go func() {
			for batch := range sendCh {
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
		}()
	}

	for {
		select {
		case <-pollTicker.C:
			g := collector.Runtime()

			for k, v := range collector.Gopsutil() {
				g[k] = v
			}

			a.gauges.Store(g)

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

			sendCh <- batch
		}
	}
}
