package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilyinon/go-musthave-metrics/internal/repository/mem"
)

func TestUpdatesHandlerRejectsBatchAtomically(t *testing.T) {
	store := mem.New()
	handler := NewUpdates(store, nil)

	body := []byte(`[
		{"id":"Load","type":"gauge","value":12.5},
		{"id":"Broken","type":"gauge"}
	]`)
	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if _, ok := store.GetGauge(context.Background(), "Load"); ok {
		t.Fatal("valid metric from rejected batch was stored")
	}
}
