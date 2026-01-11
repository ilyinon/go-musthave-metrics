package handler

import (
	"net/http"
	"strconv"

	"github.com/ilyinon/go-musthave-metrics/internal/repository"

	"github.com/go-chi/chi/v5"
)

type ValueHandler struct {
	storage repository.Storage
}

func NewValue(storage repository.Storage) *ValueHandler {
	return &ValueHandler{storage: storage}
}

func (h *ValueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch metricType {
	case "gauge":
		v, ok := h.storage.GetGauge(name)
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(strconv.FormatFloat(v, 'f', -1, 64)))

	case "counter":
		v, ok := h.storage.GetCounter(name)
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(strconv.FormatInt(v, 10)))

	default:
		http.Error(w, "bad metric type", http.StatusBadRequest)
	}
}
