package repository

import "context"

type Storage interface {
	Ping(ctx context.Context) error

	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)

	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)

	ListGauges() map[string]float64
	ListCounters() map[string]int64

	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}
