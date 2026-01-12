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
	pollInterval   time.Duration
	reportInterval time.Duration

	gauges   atomic.Value
	counters atomic.Value
}

func New(client *sender.Client, poll, report time.Duration) *App {
	return &App{
		client:         client,
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
			if g, ok := a.gauges.Load().(model.RuntimeMetrics); ok {
				for k, v := range g {
					if err := a.client.Gauge(k, v); err != nil {
						log.Printf("failed to send gauge %s: %v", k, err)
					}
				}
			}
			if c, ok := a.counters.Load().(model.CustomMetrics); ok {
				for k, v := range c {
					if err := a.client.Counter(k, v); err != nil {
						log.Printf("failed to send counter %s: %v", k, err)
					}
				}
			}
		}
	}
}
