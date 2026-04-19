package repository

import "context"

// Storage defines a unified interface for metric storage backends.
type Storage interface {
	Ping(ctx context.Context) error

	UpdateGauge(ctx context.Context, name string, value float64)
	UpdateCounter(ctx context.Context, name string, delta int64)

	GetGauge(ctx context.Context, name string) (float64, bool)
	GetCounter(ctx context.Context, name string) (int64, bool)

	GetAllGauges(ctx context.Context) map[string]float64
	GetAllCounters(ctx context.Context) map[string]int64

	ListGauges(ctx context.Context) map[string]float64
	ListCounters(ctx context.Context) map[string]int64
}
