package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"
)

type UpdateJSONHandler struct {
	storage repository.Storage
}

func NewUpdateJSON(storage repository.Storage) *UpdateJSONHandler {
	return &UpdateJSONHandler{storage: storage}
}

func (h *UpdateJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}

	var m model.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	switch m.MType {
	case "gauge":
		if m.Value == nil {
			http.Error(w, "missing value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateGauge(m.ID, *m.Value)

	case "counter":
		if m.Delta == nil {
			http.Error(w, "missing delta", http.StatusBadRequest)
			return
		}
		h.storage.UpdateCounter(m.ID, *m.Delta)

	default:
		http.Error(w, "bad metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(m)
}
