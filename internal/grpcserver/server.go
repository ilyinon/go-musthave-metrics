package grpcserver

import (
	"context"
	"errors"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/proto/metricspb"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the Metrics gRPC service.
type Server struct {
	metricspb.UnimplementedMetricsServer

	storage repository.Storage
}

// New creates a Metrics gRPC service implementation.
func New(storage repository.Storage) *Server {
	return &Server{storage: storage}
}

// UpdateMetrics updates a batch of metrics.
func (s *Server) UpdateMetrics(ctx context.Context, req *metricspb.UpdateMetricsRequest) (*metricspb.UpdateMetricsResponse, error) {
	metrics := make([]model.Metrics, 0, len(req.GetMetrics()))

	for _, metric := range req.GetMetrics() {
		switch metric.GetType() {
		case metricspb.Metric_GAUGE:
			value := metric.GetValue()
			metrics = append(metrics, model.Metrics{
				ID:    metric.GetId(),
				MType: model.MetricGauge,
				Value: &value,
			})

		case metricspb.Metric_COUNTER:
			delta := metric.GetDelta()
			metrics = append(metrics, model.Metrics{
				ID:    metric.GetId(),
				MType: model.MetricCounter,
				Delta: &delta,
			})

		default:
			return nil, status.Errorf(codes.InvalidArgument, "unknown metric type %q", metric.GetType().String())
		}
	}

	if err := s.storage.UpdateBatch(ctx, metrics); err != nil {
		if errors.Is(err, repository.ErrInvalidMetric) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return metricspb.UpdateMetricsResponse_builder{}.Build(), nil
}
