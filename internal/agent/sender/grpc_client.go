package sender

import (
	"context"
	"fmt"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/proto/metricspb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const realIPMetadataKey = "x-real-ip"

// GRPCClient sends metrics to the server over gRPC.
type GRPCClient struct {
	conn   *grpc.ClientConn
	client metricspb.MetricsClient
	realIP string
}

// NewGRPC creates a gRPC metrics client.
func NewGRPC(address string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GRPCClient{
		conn:   conn,
		client: metricspb.NewMetricsClient(conn),
		realIP: localIPForURL(address),
	}, nil
}

// Close closes the underlying gRPC connection.
func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

// SetRealIP overrides the IP sent in the x-real-ip metadata.
func (c *GRPCClient) SetRealIP(ip string) {
	c.realIP = ip
}

// Batch sends multiple metrics in a single gRPC request.
func (c *GRPCClient) Batch(metrics []model.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	var lastErr error
	for i := 0; i <= len(retryDelays); i++ {
		lastErr = c.batchOnce(metrics)
		if lastErr == nil {
			return nil
		}
		if !isRetriableGRPCError(lastErr) || i == len(retryDelays) {
			break
		}

		time.Sleep(retryDelays[i])
	}

	return lastErr
}

func (c *GRPCClient) batchOnce(metrics []model.Metrics) error {
	req := &metricspb.UpdateMetricsRequest{
		Metrics: make([]*metricspb.Metric, 0, len(metrics)),
	}

	for _, metric := range metrics {
		pbMetric, err := metricToProto(metric)
		if err != nil {
			return err
		}
		req.Metrics = append(req.Metrics, pbMetric)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if c.realIP != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, realIPMetadataKey, c.realIP)
	}

	_, err := c.client.UpdateMetrics(ctx, req)
	return err
}

func metricToProto(metric model.Metrics) (*metricspb.Metric, error) {
	switch metric.MType {
	case model.MetricGauge:
		if metric.Value == nil {
			return nil, fmt.Errorf("missing gauge value for %q", metric.ID)
		}
		return &metricspb.Metric{
			Id:    metric.ID,
			Type:  metricspb.Metric_GAUGE,
			Value: *metric.Value,
		}, nil

	case model.MetricCounter:
		if metric.Delta == nil {
			return nil, fmt.Errorf("missing counter delta for %q", metric.ID)
		}
		return &metricspb.Metric{
			Id:    metric.ID,
			Type:  metricspb.Metric_COUNTER,
			Delta: *metric.Delta,
		}, nil

	default:
		return nil, fmt.Errorf("unknown metric type %q", metric.MType)
	}
}

func isRetriableGRPCError(err error) bool {
	code := status.Code(err)
	switch code {
	case codes.DeadlineExceeded, codes.ResourceExhausted, codes.Unavailable:
		return true
	case codes.OK, codes.Canceled, codes.InvalidArgument, codes.PermissionDenied, codes.Unauthenticated:
		return false
	default:
		return false
	}
}
