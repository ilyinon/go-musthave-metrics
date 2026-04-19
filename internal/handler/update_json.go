package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/audit"
	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"
)

// UpdateJSONHandler handles HTTP requests for updating metrics in JSON format.
type UpdateJSONHandler struct {
	storage repository.Storage
	auditor *audit.Auditor
}

// NewUpdateJSON creates a new UpdateJSONHandler.
func NewUpdateJSON(storage repository.Storage, auditor *audit.Auditor) *UpdateJSONHandler {
	return &UpdateJSONHandler{storage: storage, auditor: auditor}
}

// ServeHTTP processes a JSON request with metric data and updates storage.
func (h *UpdateJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}

	var m model.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	switch m.MType {
	case model.MetricGauge:
		if m.Value == nil {
			http.Error(w, "missing value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateGauge(ctx, m.ID, *m.Value)

	case model.MetricCounter:
		if m.Delta == nil {
			http.Error(w, "missing delta", http.StatusBadRequest)
			return
		}
		h.storage.UpdateCounter(ctx, m.ID, *m.Delta)

	default:
		http.Error(w, "unknown metric type", http.StatusBadRequest)
		return
	}

	if h.auditor != nil {
		h.auditor.Notify(audit.Event{
			TS:        time.Now().Unix(),
			Metrics:   []string{m.ID},
			IPAddress: extractIP(r),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(m)
}
