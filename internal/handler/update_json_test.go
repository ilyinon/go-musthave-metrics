package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/mem"
)

func TestUpdateJSONHandler_GaugeOK(t *testing.T) {
	store := mem.New()
	h := NewUpdateJSON(store)

	val := 1.23
	body, _ := json.Marshal(model.Metrics{
		ID:    "cpu",
		MType: model.MetricGauge,
		Value: &val,
	})

	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.Background())

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	v, ok := store.GetGauge(context.Background(), "cpu")
	if !ok || v != val {
		t.Fatalf("expected gauge %v, got %v", val, v)
	}
}

func TestUpdateJSONHandler_CounterOK(t *testing.T) {
	store := mem.New()
	h := NewUpdateJSON(store)

	delta := int64(5)
	body, _ := json.Marshal(model.Metrics{
		ID:    "requests",
		MType: model.MetricCounter,
		Delta: &delta,
	})

	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	v, _ := store.GetCounter(context.Background(), "requests")
	if v != delta {
		t.Fatalf("expected counter %d, got %d", delta, v)
	}
}

func TestUpdateJSONHandler_InvalidContentType(t *testing.T) {
	store := mem.New()
	h := NewUpdateJSON(store)

	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestUpdateJSONHandler_BadJSON(t *testing.T) {
	store := mem.New()
	h := NewUpdateJSON(store)

	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader([]byte(`{bad json`)))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
