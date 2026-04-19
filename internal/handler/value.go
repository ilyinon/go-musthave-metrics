package handler

import (
	"net/http"
	"strconv"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"

	"github.com/go-chi/chi/v5"
)

// ValueHandler handles requests for retrieving metric values via URL parameters.
type ValueHandler struct {
	storage repository.Storage
}

// NewValue creates a new ValueHandler.
func NewValue(storage repository.Storage) *ValueHandler {
	return &ValueHandler{storage: storage}
}

// ServeHTTP returns a metric value based on type and name from URL parameters.
func (h *ValueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metricType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch metricType {
	case model.MetricGauge:
		v, ok := h.storage.GetGauge(ctx, name)
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(strconv.FormatFloat(v, 'f', -1, 64)))

	case model.MetricCounter:
		v, ok := h.storage.GetCounter(ctx, name)
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(strconv.FormatInt(v, 10)))

	default:
		http.Error(w, "bad metric type", http.StatusBadRequest)
	}
}
