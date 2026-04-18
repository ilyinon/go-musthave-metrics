package handler

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/audit"
	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"
)

type UpdatesHandler struct {
	storage repository.Storage
	auditor *audit.Auditor
}

func NewUpdates(storage repository.Storage, auditor *audit.Auditor) *UpdatesHandler {
	return &UpdatesHandler{storage: storage, auditor: auditor}
}

func (h *UpdatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var metrics []model.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	for _, m := range metrics {
		switch m.MType {
		case model.MetricGauge:
			if m.Value == nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			h.storage.UpdateGauge(ctx, m.ID, *m.Value)

		case model.MetricCounter:
			if m.Delta == nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			h.storage.UpdateCounter(ctx, m.ID, *m.Delta)

		default:
			http.Error(w, "unknown metric type", http.StatusBadRequest)
			return
		}
	}

	if h.auditor != nil {
		names := make([]string, 0, len(metrics))
		for _, m := range metrics {
			names = append(names, m.ID)
		}

		sort.Strings(names)

		h.auditor.Notify(audit.Event{
			TS:        time.Now().Unix(),
			Metrics:   names,
			IPAddress: extractIP(r),
		})
	}

	w.WriteHeader(http.StatusOK)
}
