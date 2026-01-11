package sender

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := New("http://localhost:8000")

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

	client := New(server.URL)

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

