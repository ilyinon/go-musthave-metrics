package grpcserver

import (
	"context"

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
	for _, metric := range req.GetMetrics() {
		switch metric.GetType() {
		case metricspb.Metric_GAUGE:
			s.storage.UpdateGauge(ctx, metric.GetId(), metric.GetValue())

		case metricspb.Metric_COUNTER:
			s.storage.UpdateCounter(ctx, metric.GetId(), metric.GetDelta())

		default:
			return nil, status.Errorf(codes.InvalidArgument, "unknown metric type %q", metric.GetType().String())
		}
	}

	return &metricspb.UpdateMetricsResponse{}, nil
}
