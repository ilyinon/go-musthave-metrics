package handler

import (
	"net/http"
	"strconv"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"

	"github.com/go-chi/chi/v5"
)

type UpdateHandler struct {
	storage repository.Storage
}

func NewUpdate(storage repository.Storage) *UpdateHandler {
	return &UpdateHandler{storage: storage}
}

func (h *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		h.storage.UpdateGauge(name, v)

	case model.MetricCounter:
		v, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			http.Error(w, "bad value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateCounter(name, v)

	default:
		http.Error(w, "bad metric type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
