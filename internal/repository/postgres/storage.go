package postgres

import (
	"context"
	"database/sql"
)

type Storage struct {
	db *sql.DB
}

func New(db *sql.DB) *Storage {
	return &Storage{db: db}
}

// Ping checks database connection.
func (s *Storage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// ---- Gauge ----

func (s *Storage) UpdateGauge(name string, value float64) {
	// not implemented yet
}

func (s *Storage) GetGauge(name string) (float64, bool) {
	return 0, false
}

func (s *Storage) ListGauges() map[string]float64 {
	return map[string]float64{}
}

func (s *Storage) GetAllGauges() map[string]float64 {
	return map[string]float64{}
}

// ---- Counter ----

func (s *Storage) UpdateCounter(name string, value int64) {
	// not implemented yet
}

func (s *Storage) GetCounter(name string) (int64, bool) {
	return 0, false
}

func (s *Storage) ListCounters() map[string]int64 {
	return map[string]int64{}
}

func (s *Storage) GetAllCounters() map[string]int64 {
	return map[string]int64{}
}

