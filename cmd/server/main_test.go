package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ==================== MemStorage ====================

func TestMemStorage_UpdateAndGetGauge(t *testing.T) {
	store := NewMemStorage()

	store.UpdateGauge("cpu", 1.23)
	val, ok := store.GetGauge("cpu")
	if !ok || val != 1.23 {
		t.Fatalf("expected 1.23, got %v, ok=%v", val, ok)
	}
}

func TestMemStorage_UpdateAndGetCounter(t *testing.T) {
	store := NewMemStorage()

	store.UpdateCounter("requests", 5)
	val, ok := store.GetCounter("requests")
	if !ok || val != 5 {
		t.Fatalf("expected 5, got %v, ok=%v", val, ok)
	}

	store.UpdateCounter("requests", 3)
	val, _ = store.GetCounter("requests")
	if val != 3 { // Обрати внимание: в твоём коде UpdateCounter перезаписывает, не накапливает
		t.Fatalf("expected 3 after overwrite, got %v", val)
	}
}

func TestMemStorage_ListGaugesAndCounters(t *testing.T) {
	store := NewMemStorage()
	store.UpdateGauge("cpu", 1.1)
	store.UpdateCounter("req", 10)

	gauges := store.ListGauges()
	if len(gauges) != 1 || gauges["cpu"] != 1.1 {
		t.Fatalf("gauges mismatch: %v", gauges)
	}

	counters := store.ListCounters()
	if len(counters) != 1 || counters["req"] != 10 {
		t.Fatalf("counters mismatch: %v", counters)
	}
}

// ==================== HTTP Handlers ====================

func TestServer_Index(t *testing.T) {
	store := NewMemStorage()
	store.UpdateGauge("g1", 0.5)
	store.UpdateCounter("c1", 7)

	server := httptest.NewServer(ServerMetrics(store))
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	if !strings.Contains(content, "Gauge g1: 0.5") {
		t.Fatalf("expected Gauge g1 in index, got %s", content)
	}
	if !strings.Contains(content, "Counter c1: 7") {
		t.Fatalf("expected Counter c1 in index, got %s", content)
	}
}

func TestServer_UpdateMetricAndGetValue(t *testing.T) {
	store := NewMemStorage()
	server := httptest.NewServer(ServerMetrics(store))
	defer server.Close()

	// POST gauge
	resp, err := http.Post(server.URL+"/update/gauge/Load/12.5", "text/plain", nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// GET gauge value
	valResp, err := http.Get(server.URL + "/value/gauge/Load")
	if err != nil {
		t.Fatal(err)
	}
	defer valResp.Body.Close()
	body, _ := io.ReadAll(valResp.Body)
	if string(body) != "12.5" {
		t.Fatalf("expected 12.5, got %s", string(body))
	}

	// POST counter
	resp, _ = http.Post(server.URL+"/update/counter/Poll/3", "text/plain", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// GET counter value
	valResp, _ = http.Get(server.URL + "/value/counter/Poll")
	body, _ = io.ReadAll(valResp.Body)
	if string(body) != "3" {
		t.Fatalf("expected 3, got %s", string(body))
	}
}

func TestServer_GetUnknownMetric(t *testing.T) {
	store := NewMemStorage()
	server := httptest.NewServer(ServerMetrics(store))
	defer server.Close()

	resp, _ := http.Get(server.URL + "/value/gauge/unknown")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown gauge, got %d", resp.StatusCode)
	}

	resp, _ = http.Get(server.URL + "/value/counter/unknown")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown counter, got %d", resp.StatusCode)
	}
}

func TestServer_BadMetricType(t *testing.T) {
	store := NewMemStorage()
	server := httptest.NewServer(ServerMetrics(store))
	defer server.Close()

	resp, _ := http.Post(server.URL+"/update/unknown/test/1", "text/plain", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad metric type, got %d", resp.StatusCode)
	}

	respGet, _ := http.Get(server.URL + "/value/unknown/test")
	if respGet.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad metric type GET, got %d", respGet.StatusCode)
	}
}

func TestServer_BadMetricValue(t *testing.T) {
	store := NewMemStorage()
	server := httptest.NewServer(ServerMetrics(store))
	defer server.Close()

	resp, _ := http.Post(server.URL+"/update/gauge/test/abc", "text/plain", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad gauge value, got %d", resp.StatusCode)
	}

	resp, _ = http.Post(server.URL+"/update/counter/test/xyz", "text/plain", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad counter value, got %d", resp.StatusCode)
	}
}

func TestServer_MethodNotAllowed(t *testing.T) {
	store := NewMemStorage()
	server := httptest.NewServer(ServerMetrics(store))
	defer server.Close()

	resp, _ := http.Get(server.URL + "/update/gauge/test/1") // GET на POST endpoint
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}
