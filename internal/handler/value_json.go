package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"
)

type ValueJSONHandler struct {
	storage repository.Storage
}

func NewValueJSON(storage repository.Storage) *ValueJSONHandler {
	return &ValueJSONHandler{storage: storage}
}

func (h *ValueJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}

	var req model.Metrics
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	resp := model.Metrics{
		ID:    req.ID,
		MType: req.MType,
	}

	switch req.MType {
	case "gauge":
		v, ok := h.storage.GetGauge(req.ID)
		if !ok {
			http.NotFound(w, r)
			return
		}
		resp.Value = &v

	case "counter":
		v, ok := h.storage.GetCounter(req.ID)
		if !ok {
			http.NotFound(w, r)
			return
		}
		resp.Delta = &v

	default:
		http.Error(w, "unknown metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
