package handler

import (
	"context"
	"html"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/repository"
)

type IndexHandler struct {
	storage repository.Storage
}

func NewIndex(storage repository.Storage) *IndexHandler {
	return &IndexHandler{storage: storage}
}

// Ping checks database connection.
// GET /ping
func (h *IndexHandler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	if err := h.storage.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ServeHTTP renders HTML page with metrics.
// GET /
func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("<html><head><title>Metrics</title></head><body>"))
	w.Write([]byte("<h1>Current Metrics</h1><ul>"))

	gauges := h.storage.ListGauges()
	gKeys := make([]string, 0, len(gauges))
	for k := range gauges {
		gKeys = append(gKeys, k)
	}
	sort.Strings(gKeys)

	for _, name := range gKeys {
		val := gauges[name]
		w.Write([]byte(
			"<li>Gauge " +
				html.EscapeString(name) +
				": " +
				strconv.FormatFloat(val, 'f', -1, 64) +
				"</li>",
		))
	}

	counters := h.storage.ListCounters()
	cKeys := make([]string, 0, len(counters))
	for k := range counters {
		cKeys = append(cKeys, k)
	}
	sort.Strings(cKeys)

	for _, name := range cKeys {
		val := counters[name]
		w.Write([]byte(
			"<li>Counter " +
				html.EscapeString(name) +
				": " +
				strconv.FormatInt(val, 10) +
				"</li>",
		))
	}

	w.Write([]byte("</ul></body></html>"))
}
