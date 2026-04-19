package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/audit"
	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"

	"github.com/go-chi/chi/v5"
)

// UpdateHandler handles HTTP requests for updating metrics via URL parameters.
type UpdateHandler struct {
	storage repository.Storage
	auditor *audit.Auditor
}

// NewUpdate creates a new UpdateHandler.
func NewUpdate(storage repository.Storage, auditor *audit.Auditor) *UpdateHandler {
	return &UpdateHandler{storage: storage, auditor: auditor}
}

// ServeHTTP processes metric updates received via URL.
func (h *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metricType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	valueStr := chi.URLParam(r, "value")

	switch metricType {
	case model.MetricGauge:
		v, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			http.Error(w, "bad value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateGauge(ctx, name, v)

	case model.MetricCounter:
		v, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			http.Error(w, "bad value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateCounter(ctx, name, v)

	default:
		http.Error(w, "bad metric type", http.StatusBadRequest)
		return
	}

	if h.auditor != nil {
		h.auditor.Notify(audit.Event{
			TS:        time.Now().Unix(),
			Metrics:   []string{name},
			IPAddress: extractIP(r),
		})
	}

	w.WriteHeader(http.StatusOK)
}
