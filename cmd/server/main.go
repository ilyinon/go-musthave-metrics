package main

import (
	"flag"
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Storage - интерфес хранилища метрик.
type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)

	// новые методы для чтения
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)

	// для списка всех метрик
	ListGauges() map[string]float64
	ListCounters() map[string]int64
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

func (s *MemStorage) GetGauge(name string) (float64, bool) {
	val, ok := s.gauges[name]
	return val, ok
}

func (s *MemStorage) GetCounter(name string) (int64, bool) {
	val, ok := s.counters[name]
	return val, ok
}

func (s *MemStorage) ListGauges() map[string]float64 {
	// возвращаем копию, чтобы не ломать инкапсуляцию
	c := make(map[string]float64, len(s.gauges))
	for k, v := range s.gauges {
		c[k] = v
	}
	return c
}

func (s *MemStorage) ListCounters() map[string]int64 {
	c := make(map[string]int64, len(s.counters))
	for k, v := range s.counters {
		c[k] = v
	}
	return c
}

// Конфигурация HTTP-сервера и маршрутов.
func ServerMetrics(storage Storage) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", getIndex(storage))
	// POST для обновления метрик
	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", sendMetric(storage))
	})

	// GET для чтения текущего значения метрики
	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", getMetricValue(storage))
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	return r
}

func getIndex(storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		w.Write([]byte("<html><head><title>Metrics</title></head><body>"))
		w.Write([]byte("<h1>Current Metrics</h1><ul>"))

		// gauges
		gauges := storage.ListGauges()
		gKeys := make([]string, 0, len(gauges))
		for k := range gauges {
			gKeys = append(gKeys, k)
		}
		sort.Strings(gKeys) // сортируем по алфавиту
		for _, name := range gKeys {
			value := gauges[name]
			w.Write([]byte("<li>Gauge " + html.EscapeString(name) + ": " + strconv.FormatFloat(value, 'f', -1, 64) + "</li>"))
		}

		// counters
		counters := storage.ListCounters()
		cKeys := make([]string, 0, len(counters))
		for k := range counters {
			cKeys = append(cKeys, k)
		}
		sort.Strings(cKeys)
		for _, name := range cKeys {
			value := counters[name]
			w.Write([]byte("<li>Counter " + html.EscapeString(name) + ": " + strconv.FormatInt(value, 10) + "</li>"))
		}

		w.Write([]byte("</ul></body></html>"))
	}
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

// GET /value/{type}/{name} — возвращает текущее значение метрики
func getMetricValue(storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		switch metricType {
		case "gauge":
			value, ok := storage.GetGauge(name)
			if !ok {
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))

		case "counter":
			value, ok := storage.GetCounter(name)
			if !ok {
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(strconv.FormatInt(value, 10)))

		default:
			http.Error(w, "bad metric type", http.StatusBadRequest)
		}
	}
}

type NetAddress struct {
	Host string
	Port int
}

func (a *NetAddress) String() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

func (a *NetAddress) Set(value string) error {
	host, portStr, err := net.SplitHostPort(value)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	a.Host = host
	a.Port = port
	return nil
}

func main() {
	storage := NewMemStorage()

	a := &NetAddress{
		Host: "localhost",
		Port: 8080,
	}

	flag.Var(a, "a", "Net address host:port")
	flag.Parse()

	// Запуск HTTP-сервера.
	log.Fatal(http.ListenAndServe(a.String(), ServerMetrics(storage)))

}
