package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/mem"
)

func TestValueHandler_UnknownMetric(t *testing.T) {
	store := mem.New()
	h := NewValue(store)

	r := chi.NewRouter()
	r.Get("/value/{type}/{name}", h.ServeHTTP)

	server := httptest.NewServer(r)
	defer server.Close()

	resp, _ := http.Get(server.URL + "/value/gauge/unknown")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

