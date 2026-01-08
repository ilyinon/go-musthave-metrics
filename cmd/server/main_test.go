package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMemStorage_UpdateGauge(t *testing.T) {
	store := NewMemStorage()
	store.UpdateGauge("metric1", 1.23)

	if got := store.gauges["metric1"]; got != 1.23 {
		t.Errorf("gauge metric mismatch: got %v, want %v", got, 1.23)
	}
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	store := NewMemStorage()
	store.UpdateCounter("counter1", 1)
	store.UpdateCounter("counter1", 2)

	if got := store.counters["counter1"]; got != 3 {
		t.Errorf("counter value mismatch: got %v, want %v", got, 3)
	}
}

func TestServer_GetIndex(t *testing.T) {
	store := NewMemStorage()
	srv := httptest.NewServer(ServerMetrics(store))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected status code: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestServer_UpdateGauge(t *testing.T) {
	store := NewMemStorage()
	srv := httptest.NewServer(ServerMetrics(store))
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/update/gauge/Load/12.5", "text/plain", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if got := store.gauges["Load"]; got != 12.5 {
		t.Errorf("stored gauge value mismatch: got %v, want %v", got, 12.5)
	}
}

func TestServer_UpdateCounter(t *testing.T) {
	store := NewMemStorage()
	srv := httptest.NewServer(ServerMetrics(store))
	defer srv.Close()

	http.Post(srv.URL+"/update/counter/Poll/1", "text/plain", nil)
	http.Post(srv.URL+"/update/counter/Poll/2", "text/plain", nil)

	if got := store.counters["Poll"]; got != 3 {
		t.Errorf("stored counter value mismatch: got %v, want %v", got, 3)
	}
}

func TestServer_BadMetricType(t *testing.T) {
	store := NewMemStorage()
	srv := httptest.NewServer(ServerMetrics(store))
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/update/unknown/test/1", "text/plain", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400 for unknown metric type, got %d", resp.StatusCode)
	}
}

func TestServer_BadMetricValue(t *testing.T) {
	store := NewMemStorage()
	srv := httptest.NewServer(ServerMetrics(store))
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/update/gauge/test/abc", "text/plain", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid value, got %d", resp.StatusCode)
	}
}

func TestServer_MethodNotAllowed(t *testing.T) {
	store := NewMemStorage()
	srv := httptest.NewServer(ServerMetrics(store))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/update/gauge/test/1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405 for wrong method, got %d", resp.StatusCode)
	}
}
