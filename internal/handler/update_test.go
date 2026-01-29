package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/mem"
)

func TestUpdateHandler_BadMetricType(t *testing.T) {
	store := mem.New()
	h := NewUpdate(store)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.ServeHTTP)

	server := httptest.NewServer(r)
	defer server.Close()

	resp, _ := http.Post(server.URL+"/update/unknown/test/1", "text/plain", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	defer resp.Body.Close()
}

func TestUpdateHandler_BadMetricValue(t *testing.T) {
	store := mem.New()
	h := NewUpdate(store)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.ServeHTTP)

	server := httptest.NewServer(r)
	defer server.Close()

	resp, _ := http.Post(server.URL+"/update/gauge/test/abc", "text/plain", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	defer resp.Body.Close()
}
