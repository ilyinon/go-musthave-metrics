package mem

import (
	"context"
	"sync"

	"github.com/ilyinon/go-musthave-metrics/internal/repository"
)

type Storage struct {
	mu       sync.RWMutex
	gauges   map[string]float64
	counters map[string]int64
}

func New() repository.Storage {
	return &Storage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *Storage) Ping(ctx context.Context) error {
	return nil
}

func (s *Storage) UpdateGauge(name string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges[name] = value
}

func (s *Storage) UpdateCounter(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counters[name] += value
}

func (s *Storage) GetGauge(name string) (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.gauges[name]
	return v, ok
}

func (s *Storage) GetCounter(name string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.counters[name]
	return v, ok
}

func (s *Storage) ListGauges() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c := make(map[string]float64, len(s.gauges))
	for k, v := range s.gauges {
		c[k] = v
	}
	return c
}

func (s *Storage) ListCounters() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c := make(map[string]int64, len(s.counters))
	for k, v := range s.counters {
		c[k] = v
	}
	return c
}

func (s *Storage) GetAllGauges() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make(map[string]float64, len(s.gauges))
	for k, v := range s.gauges {
		res[k] = v
	}
	return res
}

func (s *Storage) GetAllCounters() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make(map[string]int64, len(s.counters))
	for k, v := range s.counters {
		res[k] = v
	}
	return res
}
