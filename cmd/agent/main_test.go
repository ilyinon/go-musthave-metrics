package main

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestNewMetricsClient(t *testing.T) {
	c := NewMetricsClient("http://localhost:8000")

	if c.baseURL != "http://localhost:8000" {
		t.Errorf("baseURL mismatch: got %s, want %s", c.baseURL, "http://localhost:8000")
	}

	if c.client == nil {
		t.Error("expected resty client to be initialized, got nil")
	}
}

func TestCollectRuntimeMetrics(t *testing.T) {
	m := CollectRuntimeMetrics()

	if len(m) == 0 {
		t.Fatal("runtime metrics map is empty")
	}

	checkKeys := []string{
		"Alloc", "TotalAlloc", "Sys", "Lookups", "Mallocs", "Frees",
		"HeapAlloc", "HeapSys", "HeapIdle", "HeapInuse", "HeapReleased", "HeapObjects",
		"RandomValue",
	}

	for _, key := range checkKeys {
		if _, exists := m[key]; !exists {
			t.Errorf("expected metric %q to exist", key)
		}
	}
}

func TestCollectCustomMetrics_IncrementsCounter(t *testing.T) {
	atomic.StoreInt64(&pollCount, 0)

	first := CollectCustomMetrics()
	second := CollectCustomMetrics()
	third := CollectCustomMetrics()

	if first["PollCount"] != 1 {
		t.Errorf("first call: got %d, want 1", first["PollCount"])
	}
	if second["PollCount"] != 2 {
		t.Errorf("second call: got %d, want 2", second["PollCount"])
	}
	if third["PollCount"] != 3 {
		t.Errorf("third call: got %d, want 3", third["PollCount"])
	}
}

func TestSendGaugeAndCounter_RequestsSent(t *testing.T) {
	var paths []string

	// Мини-сервер, который просто собирает пути запросов
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewMetricsClient(server.URL)

	client.SendGauge("TestGauge", 12.34)
	client.SendCounter("TestCounter", 5)

	if len(paths) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(paths))
	}

	want := []string{
		"/update/gauge/TestGauge/12.34",
		"/update/counter/TestCounter/5",
	}

	for i, p := range want {
		if paths[i] != p {
			t.Errorf("request path mismatch: got %s, want %s", paths[i], p)
		}
	}
}
