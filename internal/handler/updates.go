package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/audit"
	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"
)

// UpdatesHandler handles batch updates of metrics.
type UpdatesHandler struct {
	storage repository.Storage
	auditor *audit.Auditor
}

// NewUpdates creates a new UpdatesHandler.
func NewUpdates(storage repository.Storage, auditor *audit.Auditor) *UpdatesHandler {
	return &UpdatesHandler{storage: storage, auditor: auditor}
}

// ServeHTTP processes a batch of metrics sent in JSON format.
func (h *UpdatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var metrics []model.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := h.storage.UpdateBatch(ctx, metrics); err != nil {
		if errors.Is(err, repository.ErrInvalidMetric) {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
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
