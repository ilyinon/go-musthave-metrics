package sender

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
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

type recordingMetricsServer struct {
	metricspb.UnimplementedMetricsServer

	metrics []*metricspb.Metric
	realIP  string
}

func (s *recordingMetricsServer) UpdateMetrics(ctx context.Context, req *metricspb.UpdateMetricsRequest) (*metricspb.UpdateMetricsResponse, error) {
	s.metrics = req.GetMetrics()

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get(realip.MetadataKey)
		if len(values) > 0 {
			s.realIP = values[0]
		}
	}

	return metricspb.UpdateMetricsResponse_builder{}.Build(), nil
}

func TestGRPCClientSendsBatchAndRealIPMetadata(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	certFile, keyFile := newSelfSignedCert(t)
	grpcCreds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		t.Fatal(err)
	}

	grpcServer := grpc.NewServer(grpc.Creds(grpcCreds))
	recordingServer := &recordingMetricsServer{}
	metricspb.RegisterMetricsServer(grpcServer, recordingServer)

	go func() {
		_ = grpcServer.Serve(listener)
	}()
	defer grpcServer.Stop()

	client, err := NewGRPC(listener.Addr().String(), certFile)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	client.SetRealIP("192.168.1.10")

	gaugeValue := 12.5
	counterValue := int64(3)
	err = client.Batch(context.Background(), []model.Metrics{
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

func TestGRPCClientBatchStopsDuringRetryDelay(t *testing.T) {
	originalRetryDelays := retryDelays
	retryDelays = []time.Duration{time.Hour}
	defer func() {
		retryDelays = originalRetryDelays
	}()

	failingClient := &unavailableMetricsClient{called: make(chan struct{})}
	client := &GRPCClient{client: failingClient}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)

	value := 12.5
	go func() {
		errCh <- client.Batch(ctx, []model.Metrics{
			{ID: "Load", MType: model.MetricGauge, Value: &value},
		})
	}()

	<-failingClient.called
	cancel()

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want context.Canceled", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Batch did not stop after context cancellation")
	}
}

type unavailableMetricsClient struct {
	called chan struct{}
}

func (c *unavailableMetricsClient) UpdateMetrics(
	context.Context,
	*metricspb.UpdateMetricsRequest,
	...grpc.CallOption,
) (*metricspb.UpdateMetricsResponse, error) {
	close(c.called)
	return nil, status.Error(codes.Unavailable, "temporary unavailable")
}

func newSelfSignedCert(t *testing.T) (string, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		t.Fatal(err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}

	certDER, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&privateKey.PublicKey,
		privateKey,
	)
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	certFile := filepath.Join(dir, "grpc-server.crt")
	keyFile := filepath.Join(dir, "grpc-server.key")

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err := os.WriteFile(certFile, certPEM, 0600); err != nil {
		t.Fatal(err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		t.Fatal(err)
	}

	return certFile, keyFile
}
