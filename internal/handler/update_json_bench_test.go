package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/mem"
)

func BenchmarkUpdateJSON(b *testing.B) {
	store := mem.New()
	h := NewUpdateJSON(store, nil)

	val := 123.45
	body, _ := json.Marshal(model.Metrics{
		ID:    "cpu",
		MType: model.MetricGauge,
		Value: &val,
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
	}
}
