package grpcserver

import (
	"context"
	"net"
	"testing"

	"github.com/ilyinon/go-musthave-metrics/internal/proto/metricspb"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/mem"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestServerUpdateMetrics(t *testing.T) {
	store := mem.New()
	server := New(store)

	_, err := server.UpdateMetrics(context.Background(), &metricspb.UpdateMetricsRequest{
		Metrics: []*metricspb.Metric{
			{Id: "Load", Type: metricspb.Metric_GAUGE, Value: 12.5},
			{Id: "PollCount", Type: metricspb.Metric_COUNTER, Delta: 3},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if got, ok := store.GetGauge(context.Background(), "Load"); !ok || got != 12.5 {
		t.Fatalf("gauge Load = %v, %v; want 12.5, true", got, ok)
	}
	if got, ok := store.GetCounter(context.Background(), "PollCount"); !ok || got != 3 {
		t.Fatalf("counter PollCount = %v, %v; want 3, true", got, ok)
	}
}

func TestTrustedSubnetInterceptor(t *testing.T) {
	_, subnet, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		subnet  *net.IPNet
		ip      string
		wantErr codes.Code
	}{
		{name: "disabled", subnet: nil, wantErr: codes.OK},
		{name: "allowed", subnet: subnet, ip: "192.168.1.10", wantErr: codes.OK},
		{name: "denied", subnet: subnet, ip: "10.0.0.1", wantErr: codes.PermissionDenied},
		{name: "missing", subnet: subnet, wantErr: codes.PermissionDenied},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.ip != "" {
				ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(realIPMetadataKey, tt.ip))
			}

			interceptor := TrustedSubnetInterceptor(tt.subnet)
			_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, func(context.Context, any) (any, error) {
				return &metricspb.UpdateMetricsResponse{}, nil
			})

			if status.Code(err) != tt.wantErr {
				t.Fatalf("status code = %v, want %v", status.Code(err), tt.wantErr)
			}
		})
	}
}
