package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"
)

type UpdatesHandler struct {
	storage repository.Storage
}

func NewUpdates(storage repository.Storage) *UpdatesHandler {
	return &UpdatesHandler{storage: storage}
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

	w.WriteHeader(http.StatusOK)
}
