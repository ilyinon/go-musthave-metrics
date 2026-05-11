package filestorage

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"
)

// Storage wraps an in-memory storage and adds file persistence.
type Storage struct {
	mem  repository.Storage
	path string
	mu   sync.Mutex
}

// New creates a new Storage with file persistence.
func New(mem repository.Storage, path string) *Storage {
	return &Storage{
		mem:  mem,
		path: path,
	}
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.mem.Ping(ctx)
}

// UpdateGauge updates gauge metric value in storage.
func (s *Storage) UpdateGauge(ctx context.Context, name string, value float64) {
	s.mem.UpdateGauge(ctx, name, value)
}

func (s *Storage) UpdateCounter(ctx context.Context, name string, value int64) {
	s.mem.UpdateCounter(ctx, name, value)
}

func (s *Storage) GetGauge(ctx context.Context, name string) (float64, bool) {
	return s.mem.GetGauge(ctx, name)
}

func (s *Storage) GetCounter(ctx context.Context, name string) (int64, bool) {
	return s.mem.GetCounter(ctx, name)
}

func (s *Storage) ListGauges(ctx context.Context) map[string]float64 {
	return s.mem.ListGauges(ctx)
}

func (s *Storage) ListCounters(ctx context.Context) map[string]int64 {
	return s.mem.ListCounters(ctx)
}

func (s *Storage) GetAllGauges(ctx context.Context) map[string]float64 {
	return s.mem.GetAllGauges(ctx)
}

func (s *Storage) GetAllCounters(ctx context.Context) map[string]int64 {
	return s.mem.GetAllCounters(ctx)
}

/*
	File persistence
*/

// Save writes all metrics to the configured file.
func (s *Storage) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()

	var data []model.Metrics

	for k, v := range s.mem.GetAllGauges(ctx) {
		val := v
		data = append(data, model.Metrics{
			ID:    k,
			MType: model.MetricGauge,
			Value: &val,
		})
	}

	for k, v := range s.mem.GetAllCounters(ctx) {
		d := v
		data = append(data, model.Metrics{
			ID:    k,
			MType: model.MetricCounter,
			Delta: &d,
		})
	}

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, b, 0644)
}

// Restore loads metrics from the configured file into memory.
func (s *Storage) Restore() error {
	b, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var data []model.Metrics
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	ctx := context.Background()

	for _, m := range data {
		switch m.MType {
		case model.MetricGauge:
			if m.Value != nil {
				s.mem.UpdateGauge(ctx, m.ID, *m.Value)
			}
		case model.MetricCounter:
			if m.Delta != nil {
				s.mem.UpdateCounter(ctx, m.ID, *m.Delta)
			}
		}
	}

	return nil
}
