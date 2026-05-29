package sender

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	appcrypto "github.com/ilyinon/go-musthave-metrics/internal/crypto"
	"github.com/ilyinon/go-musthave-metrics/internal/model"
)

func TestNewClient(t *testing.T) {
	c := New("http://localhost:8000", "")

	if c.baseURL != "http://localhost:8000" {
		t.Errorf("baseURL mismatch: got %s", c.baseURL)
	}

	if c.client == nil {
		t.Fatal("resty client is nil")
	}
}

func TestSendGaugeAndCounter(t *testing.T) {
	var paths []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL, "")

	client.Gauge("TestGauge", 12.34)
	client.Counter("TestCounter", 5)

	if len(paths) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(paths))
	}

	want := []string{
		"/update/gauge/TestGauge/12.34",
		"/update/counter/TestCounter/5",
	}

	for i := range want {
		if paths[i] != want[i] {
			t.Errorf("path mismatch: got %s, want %s", paths[i], want[i])
		}
	}
}

func TestClientSendsRealIPHeader(t *testing.T) {
	var gotHeader string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Real-IP")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL, "")
	if err := client.Gauge("TestGauge", 12.34); err != nil {
		t.Fatal(err)
	}

	if net.ParseIP(gotHeader) == nil {
		t.Fatalf("X-Real-IP = %q, want valid IP", gotHeader)
	}
}

func TestBatchWithPublicKeySendsEncryptedSignedJSONMetrics(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	const key = "test-key"

	var got []model.Metrics
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/updates/" {
			t.Errorf("path mismatch: got %s, want /updates/", r.URL.Path)
		}
		if r.Header.Get("Content-Encoding") != "gzip" {
			t.Errorf("Content-Encoding = %q, want gzip", r.Header.Get("Content-Encoding"))
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		decrypted, err := appcrypto.DecryptHybridRSA(privateKey, body)
		if err != nil {
			t.Fatal(err)
		}

		if want := appcrypto.HashSHA256(decrypted, key); r.Header.Get("HashSHA256") != want {
			t.Fatalf("hash mismatch: got %s, want %s", r.Header.Get("HashSHA256"), want)
		}

		gr, err := gzip.NewReader(bytes.NewReader(decrypted))
		if err != nil {
			t.Fatal(err)
		}
		defer gr.Close()

		payload, err := io.ReadAll(gr)
		if err != nil {
			t.Fatal(err)
		}

		if err := json.Unmarshal(payload, &got); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	gaugeValue := 12.34
	counterValue := int64(5)
	metrics := []model.Metrics{
		{ID: "TestGauge", MType: model.MetricGauge, Value: &gaugeValue},
		{ID: "TestCounter", MType: model.MetricCounter, Delta: &counterValue},
	}

	client := New(server.URL, key)
	client.SetPublicKey(&privateKey.PublicKey)

	if err := client.Batch(metrics); err != nil {
		t.Fatal(err)
	}

	if len(got) != len(metrics) {
		t.Fatalf("expected %d metrics, got %d", len(metrics), len(got))
	}
	for i := range metrics {
		if got[i].ID != metrics[i].ID || got[i].MType != metrics[i].MType {
			t.Fatalf("metric mismatch: got %+v, want %+v", got[i], metrics[i])
		}
	}
}
