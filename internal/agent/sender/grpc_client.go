package sender

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/proto/metricspb"
	"github.com/ilyinon/go-musthave-metrics/internal/realip"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GRPCClient sends metrics to the server over gRPC.
type GRPCClient struct {
	conn   *grpc.ClientConn
	client metricspb.MetricsClient
	realIP string
}

// NewGRPC creates a TLS-enabled gRPC metrics client.
func NewGRPC(address string, caFile string) (*GRPCClient, error) {
	caFile = strings.TrimSpace(caFile)
	if caFile == "" {
		return nil, errors.New("gRPC CA file is required")
	}

	creds, err := credentials.NewClientTLSFromFile(caFile, tlsServerName(address))
	if err != nil {
		return nil, err
	}

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	return &GRPCClient{
		conn:   conn,
		client: metricspb.NewMetricsClient(conn),
		realIP: localIPForURL(address),
	}, nil
}

func tlsServerName(address string) string {
	host, _ := targetHostPort(address)
	return host
}

// Close closes the underlying gRPC connection.
func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

// SetRealIP overrides the IP sent in the real IP metadata.
func (c *GRPCClient) SetRealIP(ip string) {
	c.realIP = ip
}

// Batch sends multiple metrics in a single gRPC request.
func (c *GRPCClient) Batch(ctx context.Context, metrics []model.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	var lastErr error
	for i := 0; i <= len(retryDelays); i++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = c.batchOnce(ctx, metrics)
		if lastErr == nil {
			return nil
		}
		if !isRetriableGRPCError(lastErr) || i == len(retryDelays) {
			break
		}

		if err := waitRetry(ctx, retryDelays[i]); err != nil {
			return err
		}
	}

	return lastErr
}

func (c *GRPCClient) batchOnce(ctx context.Context, metrics []model.Metrics) error {
	pbMetrics := make([]*metricspb.Metric, 0, len(metrics))

	for _, metric := range metrics {
		pbMetric, err := metricToProto(metric)
		if err != nil {
			return err
		}
		pbMetrics = append(pbMetrics, pbMetric)
	}

	req := metricspb.UpdateMetricsRequest_builder{
		Metrics: pbMetrics,
	}.Build()

	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if c.realIP != "" {
		reqCtx = metadata.AppendToOutgoingContext(reqCtx, realip.MetadataKey, c.realIP)
	}

	_, err := c.client.UpdateMetrics(reqCtx, req)
	return err
}

func metricToProto(metric model.Metrics) (*metricspb.Metric, error) {
	switch metric.MType {
	case model.MetricGauge:
		if metric.Value == nil {
			return nil, fmt.Errorf("missing gauge value for %q", metric.ID)
		}
		return metricspb.Metric_builder{
			Id:    metric.ID,
			Type:  metricspb.Metric_GAUGE,
			Value: *metric.Value,
		}.Build(), nil

	case model.MetricCounter:
		if metric.Delta == nil {
			return nil, fmt.Errorf("missing counter delta for %q", metric.ID)
		}
		return metricspb.Metric_builder{
			Id:    metric.ID,
			Type:  metricspb.Metric_COUNTER,
			Delta: *metric.Delta,
		}.Build(), nil

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
