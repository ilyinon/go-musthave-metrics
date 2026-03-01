package agent

import (
	"log"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/agent/collector"
	"github.com/ilyinon/go-musthave-metrics/internal/agent/sender"
	"github.com/ilyinon/go-musthave-metrics/internal/model"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
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
			a.gauges.Store(collector.Runtime())
			a.counters.Store(collector.Custom())

			if vm, err := mem.VirtualMemory(); err == nil {
				g := a.gauges.Load().(model.RuntimeMetrics)
				g["TotalMemory"] = float64(vm.Total)
				g["FreeMemory"] = float64(vm.Free)
				a.gauges.Store(g)
			}

			if cpuPercents, err := cpu.Percent(0, true); err == nil {
				g := a.gauges.Load().(model.RuntimeMetrics)
				for i, v := range cpuPercents {
					g[formatCPU(i)] = v
				}
				a.gauges.Store(g)
			}

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

func formatCPU(i int) string {
	return "CPUutilization" + strconv.Itoa(i+1)
}
