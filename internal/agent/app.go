package agent

import (
	"log"
	"sync"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/agent/collector"
	"github.com/ilyinon/go-musthave-metrics/internal/agent/sender"
	"github.com/ilyinon/go-musthave-metrics/internal/model"
)

const defaultRateLimit = 1

type App struct {
	client         *sender.Client
	serverURL      string
	pollInterval   time.Duration
	reportInterval time.Duration
	rateLimit      int

	mu       sync.RWMutex
	gauges   model.RuntimeMetrics
	counters model.CustomMetrics
}

func New(
	client *sender.Client,
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
					log.Printf("ERROR: batch send failed, fallback to single: %v", err)

					for _, m := range batch {
						switch m.MType {
						case model.MetricGauge:
							a.client.Gauge(m.ID, *m.Value)
						case model.MetricCounter:
							a.client.Counter(m.ID, *m.Delta)
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

			c := collector.Custom()

			a.mu.Lock()
			a.gauges = g
			a.counters = c
			a.mu.Unlock()

		case <-reportTicker.C:
			var batch []model.Metrics

			a.mu.RLock()
			g := a.gauges
			c := a.counters
			a.mu.RUnlock()

			for k, v := range g {
				val := v
				batch = append(batch, model.Metrics{
					ID:    k,
					MType: model.MetricGauge,
					Value: &val,
				})
			}

			for k, v := range c {
				delta := v
				batch = append(batch, model.Metrics{
					ID:    k,
					MType: model.MetricCounter,
					Delta: &delta,
				})
			}

			if len(batch) == 0 {
				continue
			}

			sendCh <- batch
		}
	}
}
