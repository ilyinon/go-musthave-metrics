package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Storage - интерфес хранилища метрик.
type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
}

// MemStorage - простая in-memory реализация Storage.
type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

// Конструктор in-memory хранилища.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

// Установка значения gauge-метрик.
func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.gauges[name] = value
}

// Увеличение counter-метрики.
func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.counters[name] += value
}

// Конфигурация HTTP-сервера и маршрутов.
func ServerMetrics(storage Storage) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", getIndex)
	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", sendMetric(storage))
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	return r
}

// Простой health/info endpoint.
func getIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Server metrics endpoint is /update"))
}

// Основной обработчик POST /update/{type}/{name}/{value}
func sendMetric(storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")
		valueStr := chi.URLParam(r, "value")

		switch metricType {
		case "gauge":
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				http.Error(w, "bad value", http.StatusBadRequest)
				return
			}
			storage.UpdateGauge(name, value)

		case "counter":
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				http.Error(w, "bad value", http.StatusBadRequest)
				return
			}
			storage.UpdateCounter(name, value)

		default:
			http.Error(w, "bad metric type", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func main() {
	storage := NewMemStorage()

	// Запуск HTTP-сервера.
	log.Fatal(http.ListenAndServe(":8080", ServerMetrics(storage)))

}
