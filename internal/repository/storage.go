package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
)

// ErrInvalidMetric is returned when a metric cannot be stored.
var ErrInvalidMetric = errors.New("invalid metric")

// Storage defines a unified interface for metric storage backends.
type Storage interface {
	Ping(ctx context.Context) error

	UpdateGauge(ctx context.Context, name string, value float64)
	UpdateCounter(ctx context.Context, name string, delta int64)
	UpdateBatch(ctx context.Context, metrics []model.Metrics) error

	GetGauge(ctx context.Context, name string) (float64, bool)
	GetCounter(ctx context.Context, name string) (int64, bool)

	GetAllGauges(ctx context.Context) map[string]float64
	GetAllCounters(ctx context.Context) map[string]int64

	ListGauges(ctx context.Context) map[string]float64
	ListCounters(ctx context.Context) map[string]int64
}

// ValidateMetrics verifies a batch before storage applies it.
func ValidateMetrics(metrics []model.Metrics) error {
	for _, metric := range metrics {
		switch metric.MType {
		case model.MetricGauge:
			if metric.Value == nil {
				return fmt.Errorf("%w: missing gauge value for %q", ErrInvalidMetric, metric.ID)
			}

		case model.MetricCounter:
			if metric.Delta == nil {
				return fmt.Errorf("%w: missing counter delta for %q", ErrInvalidMetric, metric.ID)
			}

		default:
			return fmt.Errorf("%w: unknown metric type %q", ErrInvalidMetric, metric.MType)
		}
	}

	return nil
}
