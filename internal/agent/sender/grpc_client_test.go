package sender

import (
	"context"
	"net"
	"testing"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/proto/metricspb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type recordingMetricsServer struct {
	metricspb.UnimplementedMetricsServer

	metrics []*metricspb.Metric
	realIP  string
}

func (s *recordingMetricsServer) UpdateMetrics(ctx context.Context, req *metricspb.UpdateMetricsRequest) (*metricspb.UpdateMetricsResponse, error) {
	s.metrics = req.GetMetrics()

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get(realIPMetadataKey)
		if len(values) > 0 {
			s.realIP = values[0]
		}
	}

	return &metricspb.UpdateMetricsResponse{}, nil
}

func TestGRPCClientSendsBatchAndRealIPMetadata(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	recordingServer := &recordingMetricsServer{}
	metricspb.RegisterMetricsServer(grpcServer, recordingServer)

	go func() {
		_ = grpcServer.Serve(listener)
	}()
	defer grpcServer.Stop()

	client, err := NewGRPC(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	client.SetRealIP("192.168.1.10")

	gaugeValue := 12.5
	counterValue := int64(3)
	err = client.Batch([]model.Metrics{
		{ID: "Load", MType: model.MetricGauge, Value: &gaugeValue},
		{ID: "PollCount", MType: model.MetricCounter, Delta: &counterValue},
	})
	if err != nil {
		t.Fatal(err)
	}

	if recordingServer.realIP != "192.168.1.10" {
		t.Fatalf("realIP = %q, want 192.168.1.10", recordingServer.realIP)
	}
	if len(recordingServer.metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(recordingServer.metrics))
	}
	if recordingServer.metrics[0].GetId() != "Load" || recordingServer.metrics[0].GetValue() != 12.5 {
		t.Fatalf("unexpected gauge metric: %+v", recordingServer.metrics[0])
	}
	if recordingServer.metrics[1].GetId() != "PollCount" || recordingServer.metrics[1].GetDelta() != 3 {
		t.Fatalf("unexpected counter metric: %+v", recordingServer.metrics[1])
	}
}
